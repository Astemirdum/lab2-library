package service

import (
	"context"

	"github.com/Astemirdum/library-service/backend/reservation/internal/errs"

	"github.com/Astemirdum/library-service/backend/reservation/internal/model"

	"github.com/Astemirdum/library-service/backend/reservation/internal/repository"
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
	rented, err := s.repo.GetRented(ctx, req.UserName)
	if err != nil {
		return model.Reservation{}, err
	}
	if rented >= req.Stars {
		return model.Reservation{}, errs.ErrNoStars
	}
	return s.repo.CreateReservation(ctx, req)
}

func (s *Service) GetReservations(ctx context.Context, username string) ([]model.Reservation, error) {
	return s.repo.GetReservations(ctx, username)
}

func (s *Service) ReservationsReturn(ctx context.Context, username, reservationUID string) (model.ReservationReturnResponse, error) {
	return s.repo.ReservationsReturn(ctx, username, reservationUID)
}

func (s *Service) RollbackReservation(ctx context.Context, uid string) error {
	return s.repo.DeleteReservation(ctx, uid)
}
