package rating

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	"github.com/Astemirdum/library-service/gateway/internal/model"

	"github.com/Astemirdum/library-service/gateway/config"
	"go.uber.org/zap"
)

type Service struct {
	log    *zap.Logger
	client *http.Client
	cfg    config.RatingHTTPServer
}

func NewService(log *zap.Logger, cfg config.Config) *Service {
	return &Service{
		log:    log,
		client: &http.Client{},
		cfg:    cfg.RatingHTTPServer,
	}
}

func (s *Service) GetRating(ctx context.Context, userName string) (model.Rating, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("http://%s/api/v1/rating", net.JoinHostPort(s.cfg.Host, s.cfg.Port)), http.NoBody)
	if err != nil {
		return model.Rating{}, http.StatusBadRequest, err
	}
	req.Header.Set("X-User-Name", userName)
	resp, err := s.client.Do(req)
	if err != nil {
		return model.Rating{}, http.StatusBadRequest, err
	}
	defer resp.Body.Close()

	var rat model.Rating
	if err := json.NewDecoder(resp.Body).Decode(&rat); err != nil {
		return model.Rating{}, http.StatusBadRequest, err
	}

	return rat, resp.StatusCode, nil
}
