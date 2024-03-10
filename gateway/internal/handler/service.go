package handler

import (
	"context"

	"github.com/Astemirdum/library-service/gateway/internal/model"
	"github.com/Astemirdum/library-service/gateway/internal/service/library"
	"github.com/Astemirdum/library-service/gateway/internal/service/provider"
	"github.com/Astemirdum/library-service/gateway/internal/service/rating"
	"github.com/Astemirdum/library-service/gateway/internal/service/reservation"
	"github.com/Astemirdum/library-service/gateway/internal/service/stats"
	"github.com/Astemirdum/library-service/pkg/circuit_breaker"
	"github.com/labstack/echo/v4"
)

//go:generate go run github.com/golang/mock/mockgen -source=service.go -destination=mocks/mock.go

var (
	_ LibraryService     = (*library.Service)(nil)
	_ RatingService      = (*rating.Service)(nil)
	_ ReservationService = (*reservation.Service)(nil)
	_ StatsService       = (*stats.Service)(nil)
	_ ProviderService    = (*provider.Service)(nil)
)

type ProviderService interface {
	Register(c echo.Context) ([]byte, int, error)
	Authorize(c echo.Context) ([]byte, int, error)
	CB() circuit_breaker.CircuitBreaker
}

type LibraryService interface {
	GetLibraries(c echo.Context) ([]byte, int, error)
	GetLibrary(ctx context.Context, libUid string) (model.GetLibrary, int, error)
	GetBooks(c echo.Context) ([]byte, int, error)
	GetBook(ctx context.Context, libUid, bookUid string) (model.GetBook, int, error)
	AvailableCount(ctx context.Context, request model.AvailableCountRequest) (status int, err error)
	CB() circuit_breaker.CircuitBreaker
}

type RatingService interface {
	GetRating(ctx context.Context, userName string) (model.Rating, int, error)
	Rating(ctx context.Context, stars int) (int, error)
	CreateRating(ctx context.Context, userName string, stars int) (int, error)
	CB() circuit_breaker.CircuitBreaker
}

type StatsService interface {
	GetStats(ctx context.Context, userName string) (model.StatsInfo, int, error)
	CB() circuit_breaker.CircuitBreaker
}

type ReservationService interface {
	GetReservation(ctx context.Context, username string) ([]model.GetReservation, int, error)
	CreateReservation(ctx context.Context, request model.CreateReservationRequest) (model.Reservation, int, error)
	RollbackReservation(ctx context.Context, uuid string) (int, error)
	ReservationReturn(ctx context.Context, req model.ReservationReturnRequest, username, reservationUid string) (model.ReservationReturnResponse, int, error)
	CB() circuit_breaker.CircuitBreaker
}
