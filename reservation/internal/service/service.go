package service

import (
	"context"
	"github.com/Astemirdum/library-service/reservation/internal/model"

	"github.com/Astemirdum/library-service/reservation/internal/repository"
	"go.uber.org/zap"
)

type Service struct {
	log  *zap.Logger
	repo repository.Repository
}

func NewService(repo repository.Repository, log *zap.Logger) *Service {
	return &Service{
		log:  log,
		repo: repo,
	}
}

func (s *Service) CreateReservation(ctx context.Context, req model.CreateReservationRequest) (model.Reservation, error) {
	return s.repo.CreateReservation(ctx, req)
}

func (s *Service) GetReservation(ctx context.Context) error {
	return nil
}
