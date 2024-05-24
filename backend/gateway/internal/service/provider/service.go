package provider

import (
	"bytes"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/pkg/errors"

	"github.com/Astemirdum/library-service/backend/gateway/config"
	"github.com/Astemirdum/library-service/backend/pkg/circuit_breaker"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type Service struct {
	log    *zap.Logger
	client *http.Client
	cfg    config.ProviderHTTPServer
	cb     circuit_breaker.CircuitBreaker
}

func NewService(log *zap.Logger, cfg config.ProviderHTTPServer) *Service {
	return &Service{
		log:    log,
		client: &http.Client{Timeout: time.Minute},
		cfg:    cfg,
		cb:     circuit_breaker.New(100, time.Second, 0.2, 2),
	}
}

func (s *Service) CB() circuit_breaker.CircuitBreaker {
	return s.cb
}

func (s *Service) proxy(c echo.Context, body io.Reader) (data []byte, statusCode int, err error) {
	ur := c.Request().URL
	ur.Scheme = "http"
	ur.Host = net.JoinHostPort(s.cfg.Host, s.cfg.Port)
	req, err := http.NewRequestWithContext(c.Request().Context(), http.MethodPost, ur.String(), body)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	req.Header = c.Request().Header.Clone()
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, http.StatusServiceUnavailable, err
	}
	defer resp.Body.Close()

	data, err = io.ReadAll(resp.Body)
	if err != nil {
		return data, http.StatusBadRequest, err
	}
	return data, resp.StatusCode, nil
}

func (s *Service) Register(c echo.Context, body []byte) ([]byte, int, error) {
	data, code, err := s.proxy(c, bytes.NewBuffer(body))
	if err != nil {
		return data, code, err
	}
	if code >= 400 {
		return nil, code, errors.New(string(data))
	}
	return data, code, err
}

func (s *Service) Authorize(c echo.Context) ([]byte, int, error) {
	return s.proxy(c, c.Request().Body)
}
