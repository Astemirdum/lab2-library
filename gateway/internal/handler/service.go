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
	GetLibrary(ctx context.Context, libUid string) (model.GetLibrary, int, error)
	GetBooks(c echo.Context) ([]byte, int, error)
	GetBook(ctx context.Context, libUid, bookUid string) (model.GetBook, int, error)
	AvailableCount(ctx context.Context, libraryID, bookID int, isReturn bool) (status int, err error)
}

type RatingService interface {
	GetRating(ctx context.Context, userName string) (model.Rating, int, error)
	Rating(ctx context.Context, userName string, stars int) (int, error)
}

type ReservationService interface {
	GetReservation(ctx context.Context, username string) ([]model.GetReservation, int, error)
	CreateReservation(ctx context.Context, request model.CreateReservationRequest) (model.Reservation, int, error)
	ReservationReturn(ctx context.Context, req model.ReservationReturnRequest, username, reservationUid string) (model.ReservationReturnResponse, int, error)
}
