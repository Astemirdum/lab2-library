package service

import (
	"context"

	"github.com/Astemirdum/library-service/library/internal/model"
	libraryRepo "github.com/Astemirdum/library-service/library/internal/repository"
	"go.uber.org/zap"
)

type Service struct {
	log  *zap.Logger
	repo libraryRepo.Repository
}

func NewService(repo libraryRepo.Repository, log *zap.Logger) *Service {
	return &Service{
		log:  log,
		repo: repo,
	}
}

func (s *Service) GetBook(ctx context.Context, libraryUid, bookUid string) (model.Book, error) {
	return s.repo.GetBook(ctx, libraryUid, bookUid)
}

func (s *Service) AvailableCount(ctx context.Context, libraryID, bookID int, isReturn bool) error {
	return s.repo.AvailableCount(ctx, libraryID, bookID, isReturn)
}

func (s *Service) GetLibrary(ctx context.Context, libraryUid string) (model.Library, error) {
	return s.repo.GetLibrary(ctx, libraryUid)
}

func (s *Service) ListLibrary(ctx context.Context, city string, page, size int) (model.ListLibraries, error) {
	return s.repo.ListLibrary(ctx, city, page, size)
}

func (s *Service) ListBooks(ctx context.Context, libraryUid string, showAll bool, page, size int) (model.ListBooks, error) {
	return s.repo.ListBooks(ctx, libraryUid, showAll, page, size)
}
