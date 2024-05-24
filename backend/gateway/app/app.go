package app

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Astemirdum/library-service/backend/gateway/config"
	"github.com/Astemirdum/library-service/backend/gateway/internal/handler"
	"github.com/Astemirdum/library-service/backend/gateway/internal/server"
	"github.com/Astemirdum/library-service/backend/pkg/kafka"
	"github.com/Astemirdum/library-service/backend/pkg/logger"
	"go.uber.org/zap"
)

func Run(cfg config.Config) error {
	log := logger.NewLogger(cfg.Log, "gateway")

	if err := kafka.CreateTopics(cfg.Kafka); err != nil {
		log.Error("create topics", zap.Error(err))
	}
	producer, err := kafka.NewSyncProducer(cfg.Kafka)
	if err != nil {
		log.DPanic("kafka", zap.Error(err))
		return err
	}
	defer producer.Close()

	asyncProducer, err := kafka.NewAsyncProducer(cfg.Kafka)
	if err != nil {
		log.DPanic("kafka", zap.Error(err))
		return err
	}
	defer asyncProducer.Close()

	h := handler.New(log, cfg, producer, asyncProducer)

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

	if err := srv.Stop(closeCtx); err != nil {
		log.Error("srv.Stop", zap.Error(err))
	}

	log.Info("Graceful shutdown finished")
	return nil
}
