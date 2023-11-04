package app

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Astemirdum/library-service/pkg/logger"
	"github.com/Astemirdum/library-service/pkg/postgres"
	"github.com/Astemirdum/library-service/rating/config"
	"github.com/Astemirdum/library-service/rating/internal/handler"
	"github.com/Astemirdum/library-service/rating/internal/repository"
	"github.com/Astemirdum/library-service/rating/internal/server"
	"github.com/Astemirdum/library-service/rating/internal/service"
	"github.com/Astemirdum/library-service/rating/migrations"
	"go.uber.org/zap"
)

func Run(cfg *config.Config) {
	log := logger.NewLogger(cfg.Log, "rating")
	db, err := postgres.NewPostgresDB(context.Background(), &cfg.Database, migrations.MigrationFiles)
	if err != nil {
		log.Fatal("db init", zap.Error(err))
	}
	repo, err := repository.NewRepository(db, log)
	if err != nil {
		log.Fatal("repo users", zap.Error(err))
	}
	svc := service.NewService(repo, log)

	//consumer, err := kafka.NewConsumer(cfg.Kafka, kafka.RatingConsumerGroup)
	//if err != nil {
	//  log.Fatal("kafka.NewConsumer", zap.Error(err))
	//}
	// go kafka.Consume(consumer, handler.NewConsumer(svc.Rating, log), kafka.RatingTopic)

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
