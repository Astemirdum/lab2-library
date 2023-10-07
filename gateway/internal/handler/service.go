package handler

import (
	"context"
	"github.com/Astemirdum/library-service/gateway/internal/model"
	"github.com/labstack/echo/v4"

	"github.com/Astemirdum/library-service/gateway/internal/service/library"
	"github.com/Astemirdum/library-service/gateway/internal/service/rating"
	"github.com/Astemirdum/library-service/gateway/internal/service/reservation"
)

//go:generate go run github.com/golang/mock/mockgen -source=service.go -destination=mocks/mock.go

var (
	_ LibraryService     = (*library.Service)(nil)
	_ RatingService      = (*rating.Service)(nil)
	_ ReservationService = (*reservation.Service)(nil)
)

type LibraryService interface {
	GetLibraries(c echo.Context) ([]byte, int, error)
	GetLibrary(ctx context.Context, libUid string) (model.Library, int, error)
	GetBooks(c echo.Context) ([]byte, int, error)
	GetBook(ctx context.Context, libUid, bookUid string) (model.Book, int, error)
}

type RatingService interface {
	GetRating(ctx context.Context, userName string) (model.Rating, int, error)
}

type ReservationService interface {
	GetReservation(c echo.Context) ([]byte, int, error)
	CreateReservation(c context.Context, request model.CreateReservationRequest) (model.Reservation, int, error)
	ReservationReturn(c echo.Context) ([]byte, int, error)
}
