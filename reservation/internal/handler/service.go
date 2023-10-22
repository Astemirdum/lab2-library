package handler

import (
	"context"

	"github.com/Astemirdum/library-service/reservation/internal/model"
	"github.com/Astemirdum/library-service/reservation/internal/service"
)

//go:generate go run github.com/golang/mock/mockgen -source=service.go -destination=mocks/mock.go

type ReservationService interface {
	CreateReservation(ctx context.Context, req model.CreateReservationRequest) (model.Reservation, error)
	GetReservations(ctx context.Context, username string) ([]model.Reservation, error)
	ReservationsReturn(ctx context.Context, username, reservationUid string) (model.ReservationReturnResponse, error)
	RollbackReservation(ctx context.Context, uid string) error
}

var _ ReservationService = (*service.Service)(nil)
