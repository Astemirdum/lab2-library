package service

import (
	"context"

	ratingModel "github.com/Astemirdum/library-service/rating/internal/model"
	ratingRepo "github.com/Astemirdum/library-service/rating/internal/repository"
	"go.uber.org/zap"
)

type Service struct {
	log  *zap.Logger
	repo ratingRepo.Repository
}

func NewService(repo ratingRepo.Repository, log *zap.Logger) *Service {
	return &Service{
		log:  log,
		repo: repo,
	}
}

func (s *Service) GetRating(ctx context.Context, name string) (ratingModel.Rating, error) {
	return s.repo.GetRating(ctx, name)
}
func (s *Service) Rating(ctx context.Context, name string, stars int) error {
	return s.repo.Rating(ctx, name, stars)
}
