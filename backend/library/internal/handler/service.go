package handler

import (
	"context"

	"github.com/Astemirdum/library-service/backend/library/internal/model"
	"github.com/Astemirdum/library-service/backend/library/internal/service"
)

//go:generate go run github.com/golang/mock/mockgen -source=service.go -destination=mocks/mock.go

type LibraryService interface {
	ListLibrary(ctx context.Context, city string, page, size int) (model.ListLibraries, error)
	ListBooks(ctx context.Context, libraryUid string, showAll bool, page, size int) (model.ListBooks, error)
	GetBook(ctx context.Context, libraryUid, bookUid string) (model.Book, error)
	GetLibrary(ctx context.Context, libraryUid string) (model.Library, error)
	AvailableCount(ctx context.Context, libraryID, bookID int, isReturn bool) error
}

var _ LibraryService = (*service.Service)(nil)
