package handler

import (
	"net/http"
	"time"

	"github.com/Astemirdum/library-service/pkg/auth"

	"github.com/Astemirdum/library-service/pkg/auth0"
	md "github.com/Astemirdum/library-service/pkg/middleware"
	"github.com/Astemirdum/library-service/pkg/validate"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

type Handler struct {
	statsSvc StatsService
	client   *http.Client
	log      *zap.Logger
}

func New(statsSvc StatsService, log *zap.Logger) *Handler {
	h := &Handler{
		statsSvc: statsSvc,
		log:      log,
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
	e.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{StackSize: 4 << 10}))
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{http.MethodGet, http.MethodOptions, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
		AllowCredentials: true,
	}))

	base := e.Group("", md.NewRateLimiter(baseRPS))
	base.GET("/manage/health", h.Health)

	e.Validator = validate.NewCustomValidator()
	api := e.Group("/api/v1",
		middleware.RequestLoggerWithConfig(md.RequestLoggerConfig()),
		middleware.RequestID(),
		md.NewRateLimiter(apiRPS),
		auth0.MiddlewareUserName,
	)
	api = api.Group("", md.AuthContext)
	api.GET("/stats", h.GetStats)
	return e
}

func (h *Handler) Health(c echo.Context) error {
	return c.String(http.StatusOK, "OK")
}

func (h *Handler) GetStats(c echo.Context) error {
	ctx := c.Request().Context()
	if !auth.IsAdmin(ctx) {
		return echo.NewHTTPError(http.StatusUnauthorized, "no admin")
	}

	stat, err := h.statsSvc.GetStats(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, stat)
}
