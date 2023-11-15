package app

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Astemirdum/library-service/pkg/kafka"

	"github.com/Astemirdum/library-service/library/migrations"

	"github.com/Astemirdum/library-service/library/config"
	"github.com/Astemirdum/library-service/library/internal/handler"
	"github.com/Astemirdum/library-service/library/internal/repository"
	"github.com/Astemirdum/library-service/library/internal/server"
	"github.com/Astemirdum/library-service/library/internal/service"
	"github.com/Astemirdum/library-service/pkg/logger"
	"github.com/Astemirdum/library-service/pkg/postgres"
	"go.uber.org/zap"
)

func Run(cfg *config.Config) {
	log := logger.NewLogger(cfg.Log, "library")
	db, err := postgres.NewPostgresDB(context.Background(), &cfg.Database, migrations.MigrationFiles)
	if err != nil {
		log.Fatal("db init", zap.Error(err))
	}
	repo, err := repository.NewRepository(db, log)
	if err != nil {
		log.Fatal("repo", zap.Error(err))
	}
	svc := service.NewService(repo, log)

	consumer, err := kafka.NewConsumer(cfg.Kafka, kafka.LibraryConsumerGroup)
	if err != nil {
		log.Fatal("kafka.NewConsumer", zap.Error(err))
	}
	go kafka.Consume(consumer, handler.NewConsumer(svc.AvailableCount, log), kafka.LibraryTopic)

	h := handler.New(svc, log)
	srv := server.NewServer(cfg.Server, h.NewRouter())
	log.Info("http server start ON: ",
		zap.String("addr",
			net.JoinHostPort(cfg.Server.Host, cfg.Server.Port)))
	go func() {
		if err := srv.Run(); err != nil {
			log.Error("server run", zap.Error(err))
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	termSig := <-sig

	log.Debug("Graceful shutdown", zap.Any("signal", termSig))

	closeCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	if err = srv.Stop(closeCtx); err != nil {
		log.DPanic("srv.Stop", zap.Error(err))
	}
	db.Close()
	log.Info("Graceful shutdown finished")
}
