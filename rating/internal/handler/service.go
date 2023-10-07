package handler

import (
	"context"
	ratingModel "github.com/Astemirdum/library-service/rating/internal/model"

	"github.com/Astemirdum/library-service/rating/internal/service"
)

//go:generate go run github.com/golang/mock/mockgen -source=service.go -destination=mocks/mock.go

type RatingService interface {
	GetRating(ctx context.Context, name string) (ratingModel.Rating, error)
}

var _ RatingService = (*service.Service)(nil)
