package library

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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
	cfg    config.LibraryHTTPServer
}

func NewService(log *zap.Logger, cfg config.Config) *Service {
	return &Service{
		log:    log,
		client: &http.Client{Timeout: time.Minute},
		cfg:    cfg.LibraryHTTPServer,
	}
}

func (s *Service) GetBook(ctx context.Context, libUid, bookUid string) (model.Book, int, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("http://%s/api/v1/libraries/%s/books/%s", net.JoinHostPort(s.cfg.Host, s.cfg.Port), libUid, bookUid),
		nil)
	if err != nil {
		return model.Book{}, http.StatusBadRequest, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return model.Book{}, http.StatusBadRequest, err
	}
	var book model.Book
	if err := json.NewDecoder(resp.Body).Decode(&book); err != nil {
		return model.Book{}, http.StatusBadRequest, err
	}
	defer resp.Body.Close()
	return book, resp.StatusCode, nil
}

func (s *Service) GetLibrary(ctx context.Context, libUid string) (model.Library, int, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("http://%s/api/v1/libraries/%s", net.JoinHostPort(s.cfg.Host, s.cfg.Port), libUid),
		nil)
	if err != nil {
		return model.Library{}, http.StatusBadRequest, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return model.Library{}, http.StatusBadRequest, err
	}
	var lib model.Library
	if err := json.NewDecoder(resp.Body).Decode(&lib); err != nil {
		return model.Library{}, http.StatusBadRequest, err
	}
	defer resp.Body.Close()
	return lib, resp.StatusCode, nil
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
		return nil, http.StatusBadRequest, err
	}
	defer resp.Body.Close()

	data, err = io.ReadAll(resp.Body)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	return data, resp.StatusCode, nil
}
