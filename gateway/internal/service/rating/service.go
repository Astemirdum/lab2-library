package rating

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	"github.com/Astemirdum/library-service/gateway/internal/errs"

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

	if resp.StatusCode >= 400 {
		err = errs.ErrDefault
	}

	return rat, resp.StatusCode, err
}

func (s *Service) Rating(ctx context.Context, userName string, stars int) (int, error) {
	b := bytes.NewBuffer(nil)
	ratingReq := struct {
		Stars int `json:"stars"`
	}{
		Stars: stars,
	}
	if err := json.NewEncoder(b).Encode(ratingReq); err != nil {
		return http.StatusBadRequest, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, fmt.Sprintf("http://%s/api/v1/rating", net.JoinHostPort(s.cfg.Host, s.cfg.Port)), b)
	if err != nil {
		return http.StatusBadRequest, err
	}
	req.Header.Set("X-User-Name", userName)
	resp, err := s.client.Do(req)
	if err != nil {
		return http.StatusBadRequest, err
	}
	defer resp.Body.Close()

	var rat model.Rating
	if err := json.NewDecoder(resp.Body).Decode(&rat); err != nil {
		return http.StatusBadRequest, err
	}
	if resp.StatusCode >= 400 {
		err = errs.ErrDefault
	}
	return resp.StatusCode, err
}
