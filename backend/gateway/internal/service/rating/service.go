package rating

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/Astemirdum/library-service/backend/pkg/auth"

	"github.com/pkg/errors"

	"github.com/Astemirdum/library-service/backend/pkg/circuit_breaker"

	"github.com/labstack/echo/v4"

	"github.com/Astemirdum/library-service/backend/gateway/internal/errs"

	"github.com/Astemirdum/library-service/backend/gateway/internal/model"

	"github.com/Astemirdum/library-service/backend/gateway/config"
	"go.uber.org/zap"
)

type Service struct {
	log      *zap.Logger
	client   *http.Client
	cfg      config.RatingHTTPServer
	cb       circuit_breaker.CircuitBreaker
	endpoint string
}

func NewService(log *zap.Logger, cfg config.RatingHTTPServer) *Service {
	return &Service{
		log:      log,
		client:   &http.Client{Timeout: time.Minute},
		cfg:      cfg,
		cb:       circuit_breaker.New(100, time.Second, 0.2, 2),
		endpoint: fmt.Sprintf("http://%s/api/v1/rating", net.JoinHostPort(cfg.Host, cfg.Port)),
	}
}

func (s *Service) CB() circuit_breaker.CircuitBreaker {
	return s.cb
}

func (s *Service) GetRating(ctx context.Context) (model.Rating, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.endpoint, http.NoBody)
	if err != nil {
		return model.Rating{}, http.StatusBadRequest, err
	}
	auth.SetAuthHeader(req)
	resp, err := s.client.Do(req)
	if err != nil {
		return model.Rating{}, http.StatusServiceUnavailable, errors.New("rating unavailable")
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

func (s *Service) Rating(ctx context.Context, stars int) (int, error) {
	b := bytes.NewBuffer(nil)
	var ratingReq model.Rating
	ratingReq.Stars = stars
	if err := json.NewEncoder(b).Encode(ratingReq); err != nil {
		return http.StatusBadRequest, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, s.endpoint, b)
	if err != nil {
		return http.StatusBadRequest, err
	}
	auth.SetAuthHeader(req)
	req.Header.Set("Content-Type", echo.MIMEApplicationJSON)
	resp, err := s.client.Do(req)
	if err != nil {
		return http.StatusServiceUnavailable, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		err = errs.ErrDefault
	}
	return resp.StatusCode, err
}

func (s *Service) CreateRating(ctx context.Context, userName string, stars int) (int, error) {
	b := bytes.NewBuffer(nil)
	ratingReq := model.CreateRating{
		Name:  userName,
		Stars: stars,
	}
	if err := json.NewEncoder(b).Encode(ratingReq); err != nil {
		return http.StatusBadRequest, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.endpoint, b)
	if err != nil {
		return http.StatusBadRequest, err
	}
	req.Header.Set("Content-Type", echo.MIMEApplicationJSON)
	resp, err := s.client.Do(req)
	if err != nil {
		return http.StatusServiceUnavailable, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		d, _ := io.ReadAll(resp.Body) //nolint:errcheck
		err = errors.New(string(d))
	}
	return resp.StatusCode, err
}
