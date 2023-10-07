package handler

import (
	"net/http"
	"time"

	"github.com/Astemirdum/library-service/rating/internal/errs"
	"github.com/pkg/errors"

	"github.com/Astemirdum/library-service/pkg/validate"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"

	_ "github.com/Astemirdum/library-service/swagger"
)

type Handler struct {
	ratingSvc RatingService
	client    *http.Client
	log       *zap.Logger
}

func New(ratingSvc RatingService, log *zap.Logger) *Handler {
	h := &Handler{
		ratingSvc: ratingSvc,
		log:       log,
		client: &http.Client{
			Timeout: time.Minute,
		},
	}
	return h
}

func (h *Handler) NewRouter() *echo.Echo {
	e := echo.New()
	const (
		baseRPS = 10
		apiRPS  = 100
	)
	e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
		StackSize: 4 << 10, // 4 KB
	}))
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{http.MethodGet, http.MethodOptions, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
		AllowCredentials: true,
	}))

	base := e.Group("", newRateLimiterMW(baseRPS))
	base.GET("/manage/health", h.Health)

	e.Validator = validate.NewCustomValidator()
	api := e.Group("/api/v1",
		middleware.RequestLoggerWithConfig(requestLoggerConfig()),
		middleware.RequestID(),
		newRateLimiterMW(apiRPS),
	)

	api.GET("/rating", h.GetRating)

	return e
}

func (h *Handler) Health(c echo.Context) error {
	return c.String(http.StatusOK, "OK")
}

func (h *Handler) GetRating(c echo.Context) error {
	ctx := c.Request().Context()

	userName := c.Request().Header.Get("X-User-Name")
	if userName == "" {
		return echo.NewHTTPError(http.StatusNotFound, "username is empty")
	}

	stars, err := h.ratingSvc.GetRating(ctx, userName)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, stars)
}
