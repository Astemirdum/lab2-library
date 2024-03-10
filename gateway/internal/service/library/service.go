package library

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/Astemirdum/library-service/pkg/circuit_breaker"

	"github.com/Astemirdum/library-service/gateway/internal/errs"

	"github.com/Astemirdum/library-service/gateway/config"
	"github.com/Astemirdum/library-service/gateway/internal/model"
	"github.com/labstack/echo/v4"

	"go.uber.org/zap"
)

type Service struct {
	log    *zap.Logger
	client *http.Client
	cfg    config.LibraryHTTPServer
	cb     circuit_breaker.CircuitBreaker
}

func NewService(log *zap.Logger, cfg config.LibraryHTTPServer) *Service {
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

func (s *Service) GetBook(ctx context.Context, libUid, bookUid string) (model.GetBook, int, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("http://%s/api/v1/libraries/%s/books/%s", net.JoinHostPort(s.cfg.Host, s.cfg.Port), libUid, bookUid),
		nil)
	if err != nil {
		return model.GetBook{}, http.StatusBadRequest, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return model.GetBook{}, http.StatusServiceUnavailable, err
	}
	defer resp.Body.Close()
	var book model.GetBook
	if err := json.NewDecoder(resp.Body).Decode(&book); err != nil {
		return model.GetBook{}, http.StatusBadRequest, err
	}
	if resp.StatusCode >= 400 {
		err = errs.ErrDefault
	}
	return book, resp.StatusCode, err
}

func (s *Service) AvailableCount(ctx context.Context, inp model.AvailableCountRequest) (status int, err error) {
	b := bytes.NewBuffer(nil)
	if err := json.NewEncoder(b).Encode(inp); err != nil {
		return http.StatusBadRequest, err
	}
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPatch,
		fmt.Sprintf("http://%s/api/v1/libraries/books", net.JoinHostPort(s.cfg.Host, s.cfg.Port)),
		b)
	if err != nil {
		return http.StatusBadRequest, err
	}
	req.Header.Set("Content-Type", echo.MIMEApplicationJSONCharsetUTF8)
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

func (s *Service) GetLibrary(ctx context.Context, libUid string) (model.GetLibrary, int, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("http://%s/api/v1/libraries/%s", net.JoinHostPort(s.cfg.Host, s.cfg.Port), libUid),
		nil)
	if err != nil {
		return model.GetLibrary{}, http.StatusBadRequest, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return model.GetLibrary{}, http.StatusServiceUnavailable, err
	}
	defer resp.Body.Close()
	var lib model.GetLibrary
	if err := json.NewDecoder(resp.Body).Decode(&lib); err != nil {
		return model.GetLibrary{}, http.StatusBadRequest, err
	}
	if resp.StatusCode >= 400 {
		err = errs.ErrDefault
	}
	return lib, resp.StatusCode, err
}

func (s *Service) GetBooks(c echo.Context) (data []byte, statusCode int, err error) {
	return s.proxy(c)
}

func (s *Service) GetLibraries(c echo.Context) (data []byte, statusCode int, err error) {
	return s.proxy(c)
}

func (s *Service) proxy(c echo.Context) (data []byte, statusCode int, err error) {
	ur := c.Request().URL
	ur.Scheme = "http"
	ur.Host = net.JoinHostPort(s.cfg.Host, s.cfg.Port)
	req, err := http.NewRequestWithContext(c.Request().Context(), http.MethodGet, ur.String(), http.NoBody)
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
		return nil, http.StatusBadRequest, err
	}
	return data, resp.StatusCode, nil
}
