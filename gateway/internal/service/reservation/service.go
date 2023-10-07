package reservation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/Astemirdum/library-service/gateway/config"
	"github.com/Astemirdum/library-service/gateway/internal/model"
	"github.com/labstack/echo/v4"

	"go.uber.org/zap"
)

type Service struct {
	log    *zap.Logger
	client *http.Client
	cfg    config.ReservationHTTPServer
}

func NewService(log *zap.Logger, cfg config.Config) *Service { //nolint:gocritic
	return &Service{
		log:    log,
		client: &http.Client{Timeout: time.Minute},
		cfg:    cfg.ReservationHTTPServer,
	}
}

const (
	XUserName = "X-User-Name"
)

func (s *Service) GetReservation(ctx context.Context, username string) ([]model.GetReservation, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("http://%s/api/v1/reservations", net.JoinHostPort(s.cfg.Host, s.cfg.Port)), http.NoBody)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	req.Header.Set(XUserName, username)
	req.Header.Set("Content-Type", echo.MIMEApplicationJSONCharsetUTF8)
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	defer resp.Body.Close()

	var rsv []model.GetReservation
	if err := json.NewDecoder(resp.Body).Decode(&rsv); err != nil {
		return nil, http.StatusBadRequest, err
	}
	return rsv, resp.StatusCode, nil
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
	req.Header.Set(XUserName, request.UserName)
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

func (s *Service) ReservationReturn(ctx context.Context, req model.ReservationReturnRequest, username, reservationUid string) (model.ReservationReturnResponse, int, error) {
	b := bytes.NewBuffer(nil)
	if err := json.NewEncoder(b).Encode(req); err != nil {
		return model.ReservationReturnResponse{}, http.StatusBadRequest, err
	}
	r, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("http://%s/api/v1/reservations/%s/return", net.JoinHostPort(s.cfg.Host, s.cfg.Port), reservationUid), b)
	if err != nil {
		return model.ReservationReturnResponse{}, http.StatusBadRequest, err
	}
	r.Header.Set("Content-Type", echo.MIMEApplicationJSONCharsetUTF8)
	r.Header.Set(XUserName, username)
	resp, err := s.client.Do(r)
	if err != nil {
		return model.ReservationReturnResponse{}, http.StatusBadRequest, err
	}
	defer resp.Body.Close()

	var res model.ReservationReturnResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return model.ReservationReturnResponse{}, http.StatusBadRequest, err
	}
	return res, resp.StatusCode, nil
}
