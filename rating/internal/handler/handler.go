package handler

import (
	"net/http"
	"time"

	"github.com/Astemirdum/library-service/rating/internal/model"

	"github.com/Astemirdum/library-service/pkg/auth"
	md "github.com/Astemirdum/library-service/pkg/middleware"

	"github.com/Astemirdum/library-service/rating/internal/errs"
	"github.com/pkg/errors"

	"github.com/Astemirdum/library-service/pkg/validate"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
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

	base := e.Group("", md.NewRateLimiter(baseRPS))
	base.GET("/manage/health", h.Health)

	e.Validator = validate.NewCustomValidator()

	api := e.Group("/api/v1",
		middleware.RequestLoggerWithConfig(md.RequestLoggerConfig()),
		middleware.RequestID(),
		md.NewRateLimiter(apiRPS),
	)
	api.POST("/rating", h.Rating)

	api = api.Group("", md.AuthContext)
	api.GET("/rating", h.GetRating)
	api.PATCH("/rating", h.Rating)

	return e
}

func (h *Handler) Health(c echo.Context) error {
	return c.String(http.StatusOK, "OK")
}

func (h *Handler) Rating(c echo.Context) error {
	ctx := c.Request().Context()
	userName, err := auth.GetUserName(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	var ratingReq model.Rating
	if err := c.Bind(&ratingReq); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.ratingSvc.Rating(ctx, userName, ratingReq.Stars); err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) CreateRating(c echo.Context) error {
	ctx := c.Request().Context()

	var ratingReq model.CreateRating
	if err := c.Bind(&ratingReq); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.ratingSvc.CreateRating(ctx, ratingReq.Name, ratingReq.Stars); err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) GetRating(c echo.Context) error {
	ctx := c.Request().Context()

	userName, err := auth.GetUserName(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
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
