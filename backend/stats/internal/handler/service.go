package handler

import (
	"context"

	"github.com/Astemirdum/library-service/backend/pkg/kafka"

	statsModel "github.com/Astemirdum/library-service/backend/stats/internal/model"
	"github.com/Astemirdum/library-service/backend/stats/internal/service"
)

//go:generate go run github.com/golang/mock/mockgen -source=service.go -destination=mocks/mock.go

type StatsService interface {
	GetStats(ctx context.Context) (statsModel.StatsInfo, error)
	Stats(ctx context.Context, eventStats kafka.EventStats) error
}

var _ StatsService = (*service.Service)(nil)
