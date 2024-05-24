package app

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Astemirdum/library-service/backend/library/config"
	"github.com/Astemirdum/library-service/backend/library/internal/handler"
	"github.com/Astemirdum/library-service/backend/library/internal/repository"
	"github.com/Astemirdum/library-service/backend/library/internal/server"
	"github.com/Astemirdum/library-service/backend/library/internal/service"
	"github.com/Astemirdum/library-service/backend/library/migrations"
	"github.com/Astemirdum/library-service/backend/pkg/kafka"
	"github.com/Astemirdum/library-service/backend/pkg/logger"
	"github.com/Astemirdum/library-service/backend/pkg/postgres"
	"go.uber.org/zap"
)

func Run(cfg *config.Config) error {
	log := logger.NewLogger(cfg.Log, "library")
	db, err := postgres.NewPostgresDB(context.Background(), &cfg.Database, migrations.MigrationFiles)
	if err != nil {
		return fmt.Errorf("db init %v", err)
	}
	repo, err := repository.NewRepository(db, log)
	if err != nil {
		return fmt.Errorf("repo %v", err)
	}
	svc := service.NewService(repo, log)

	consumer, err := kafka.NewConsumer(cfg.Kafka, kafka.LibraryConsumerGroup)
	if err != nil {
		return fmt.Errorf("kafka.NewConsumer %v", err)
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
		log.Error("srv.Stop", zap.Error(err))
	}
	db.Close()
	log.Info("Graceful shutdown finished")
	return nil
}
