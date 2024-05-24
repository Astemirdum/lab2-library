package service

import (
	"context"

	statsModel "github.com/Astemirdum/library-service/backend/pkg/kafka"

	"github.com/Astemirdum/library-service/backend/stats/internal/model"
	statsRepo "github.com/Astemirdum/library-service/backend/stats/internal/repository"
	"go.uber.org/zap"
)

type Service struct {
	log  *zap.Logger
	repo statsRepo.Repository
}

func NewService(repo statsRepo.Repository, log *zap.Logger) *Service {
	return &Service{
		log:  log,
		repo: repo,
	}
}

// GetStats get stats by user.
func (s *Service) GetStats(ctx context.Context) (model.StatsInfo, error) {
	return s.repo.GetStats(ctx)
}

// Stats used by kafka consumer.
func (s *Service) Stats(ctx context.Context, event statsModel.EventStats) error {
	return s.repo.Stats(ctx, event)
}
