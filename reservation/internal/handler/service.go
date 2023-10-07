package handler

import (
	"context"

	"github.com/Astemirdum/library-service/reservation/internal/model"
	"github.com/Astemirdum/library-service/reservation/internal/service"
)

//go:generate go run github.com/golang/mock/mockgen -source=service.go -destination=mocks/mock.go

type LibraryService interface {
	CreateReservation(ctx context.Context, req model.CreateReservationRequest) (model.Reservation, error)
	GetReservation(ctx context.Context) error
}

var _ LibraryService = (*service.Service)(nil)
