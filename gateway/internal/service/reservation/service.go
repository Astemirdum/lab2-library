package reservation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/Astemirdum/library-service/gateway/config"
	"github.com/Astemirdum/library-service/gateway/internal/model"
	"github.com/labstack/echo/v4"
	"net"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type Service struct {
	log    *zap.Logger
	client *http.Client
	cfg    config.ReservationHTTPServer
}

func NewService(log *zap.Logger, cfg config.Config) *Service {
	return &Service{
		log:    log,
		client: &http.Client{Timeout: time.Minute},
		cfg:    cfg.ReservationHTTPServer,
	}
}

func (s *Service) GetReservation(c echo.Context) ([]byte, int, error) {
	return nil, 0, nil
}

func (s *Service) CreateReservation(ctx context.Context, request model.CreateReservationRequest) (model.Reservation, int, error) {
	b := bytes.NewBuffer(nil)
	if err := json.NewEncoder(b).Encode(request); err != nil {
		return model.Reservation{}, http.StatusBadRequest, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("http://%s/api/v1/reservations", net.JoinHostPort(s.cfg.Host, s.cfg.Port)), b)
	if err != nil {
		return model.Reservation{}, http.StatusBadRequest, err
	}
	req.Header.Set("X-User-Name", request.UserName)
	req.Header.Set("Content-Type", echo.MIMEApplicationJSONCharsetUTF8)
	resp, err := s.client.Do(req)
	if err != nil {
		return model.Reservation{}, http.StatusBadRequest, err
	}
	defer resp.Body.Close()

	var rsv model.Reservation
	if err := json.NewDecoder(resp.Body).Decode(&rsv); err != nil {
		return model.Reservation{}, http.StatusBadRequest, err
	}
	return rsv, resp.StatusCode, nil
}

func (s *Service) ReservationReturn(c echo.Context) ([]byte, int, error) {
	return nil, 0, nil
}
