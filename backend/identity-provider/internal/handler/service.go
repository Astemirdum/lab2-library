package handler

import (
	"context"

	"github.com/Astemirdum/library-service/backend/identity-provider/internal/model"
	"github.com/Astemirdum/library-service/backend/identity-provider/internal/service"
)

//go:generate go run github.com/golang/mock/mockgen -source=service.go -destination=mocks/mock.go

type AuthService interface {
	RegisterUser(ctx context.Context, user model.UserCreateRequest) error
	GetUser(ctx context.Context, username string) (model.User, error)
}

var _ AuthService = (*service.Service)(nil)
