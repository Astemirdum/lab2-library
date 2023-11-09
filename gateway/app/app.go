package app

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Astemirdum/library-service/gateway/config"
	"github.com/Astemirdum/library-service/gateway/internal/handler"
	"github.com/Astemirdum/library-service/gateway/internal/server"
	"github.com/Astemirdum/library-service/pkg/logger"
	"go.uber.org/zap"
)

func Run(cfg config.Config) {
	log := logger.NewLogger(cfg.Log, "gateway")
	//producer, err := kafka.NewProducer(cfg.Kafka)
	//if err != nil {
	//	log.DPanic("kafka", zap.Error(err))
	//}
	h := handler.New(log, cfg, nil)

	srv := server.NewServer(cfg.Server, h.NewRouter(cfg.Auth0))
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
		log.DPanic("srv.Stop", zap.Error(err))
	}
	log.Info("Graceful shutdown finished")
	//_ = producer.Close()
}
