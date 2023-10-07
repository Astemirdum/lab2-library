package app

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Astemirdum/library-service/reservation/migrations"

	"github.com/Astemirdum/library-service/pkg/logger"
	"github.com/Astemirdum/library-service/pkg/postgres"
	"github.com/Astemirdum/library-service/reservation/config"
	"github.com/Astemirdum/library-service/reservation/internal/handler"
	"github.com/Astemirdum/library-service/reservation/internal/repository"
	"github.com/Astemirdum/library-service/reservation/internal/server"
	"github.com/Astemirdum/library-service/reservation/internal/service"
	"go.uber.org/zap"
)

func Run(cfg *config.Config) {
	log := logger.NewLogger(cfg.Log, "reservation")
	db, err := postgres.NewPostgresDB(&cfg.Database, migrations.MigrationFiles)
	if err != nil {
		log.Fatal("db init", zap.Error(err))
	}
	repo, err := repository.NewRepository(db, log)
	if err != nil {
		log.Fatal("repo users", zap.Error(err))
	}
	userService := service.NewService(repo, log)

	h := handler.New(userService, log)

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
	if err = db.Close(); err != nil {
		log.DPanic(" db.Close()", zap.Error(err))
	}
	log.Info("Graceful shutdown finished")
}
