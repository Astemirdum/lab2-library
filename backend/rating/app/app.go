package app

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Astemirdum/library-service/backend/pkg/kafka"
	"github.com/Astemirdum/library-service/backend/pkg/logger"
	"github.com/Astemirdum/library-service/backend/pkg/postgres"
	"github.com/Astemirdum/library-service/backend/rating/config"
	"github.com/Astemirdum/library-service/backend/rating/internal/handler"
	"github.com/Astemirdum/library-service/backend/rating/internal/repository"
	"github.com/Astemirdum/library-service/backend/rating/internal/server"
	"github.com/Astemirdum/library-service/backend/rating/internal/service"
	"github.com/Astemirdum/library-service/backend/rating/migrations"
	"go.uber.org/zap"
)

func Run(cfg *config.Config) error {
	//fmt.Printf("%+v\n", cfg)
	log := logger.NewLogger(cfg.Log, "rating")
	db, err := postgres.NewPostgresDB(context.Background(), &cfg.Database, migrations.MigrationFiles)
	if err != nil {
		return fmt.Errorf("db init %w", err)
	}
	repo, err := repository.NewRepository(db, log)
	if err != nil {
		return fmt.Errorf("repo users %w", err)
	}
	svc := service.NewService(repo, log)

	consumer, err := kafka.NewConsumer(cfg.Kafka, kafka.RatingConsumerGroup)
	if err != nil {
		return fmt.Errorf("kafka.NewConsumer %w", err)
	}
	go kafka.Consume(consumer, handler.NewConsumer(svc.Rating, log), kafka.RatingTopic)

	h := handler.New(svc, log)
	srv := server.NewServer(cfg.Server, h.NewRouter())
	log.Info("http server start ON: ",
		zap.String("addr",
			net.JoinHostPort(cfg.Server.Host, cfg.Server.Port)))
	go func() {
		if err = srv.Run(); err != nil {
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
	return nil
}
