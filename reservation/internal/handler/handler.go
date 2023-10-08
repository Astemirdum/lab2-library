package handler

import (
	"net/http"
	"time"

	"github.com/pkg/errors"

	"github.com/Astemirdum/library-service/reservation/internal/errs"
	"github.com/Astemirdum/library-service/reservation/internal/model"

	"github.com/Astemirdum/library-service/pkg/validate"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

type Handler struct {
	reservationSvc ReservationService
	client         *http.Client
	log            *zap.Logger
}

func New(reservationSrv ReservationService, log *zap.Logger) *Handler {
	h := &Handler{
		reservationSvc: reservationSrv,
		log:            log,
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

	api.GET("/reservations", h.GetReservations)
	api.POST("/reservations", h.CreateReservation)
	api.POST("/reservations/:reservationUid/return", h.ReservationsReturn)

	return e
}

func (h *Handler) Health(c echo.Context) error {
	return c.String(http.StatusOK, "OK")
}

const (
	XUserName = "X-User-Name"
)

func (h *Handler) CreateReservation(c echo.Context) error {
	var req model.CreateReservationRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	req.UserName = c.Request().Header.Get(XUserName)
	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	ctx := c.Request().Context()
	resp, err := h.reservationSvc.CreateReservation(ctx, req)
	if err != nil {
		if errors.Is(err, errs.ErrNoStars) {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) GetReservations(c echo.Context) error {
	ctx := c.Request().Context()
	username := c.Request().Header.Get(XUserName)
	if username == "" {
		return echo.NewHTTPError(http.StatusBadRequest, errs.ErrUserName)
	}
	rsv, err := h.reservationSvc.GetReservations(ctx, username)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, rsv)
}

func (h *Handler) ReservationsReturn(c echo.Context) error {
	ctx := c.Request().Context()
	username := c.Request().Header.Get(XUserName)
	if username == "" {
		return echo.NewHTTPError(http.StatusBadRequest, errs.ErrUserName)
	}
	reservationUid := c.Param("reservationUid")

	var req model.ReservationReturnRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	resp, err := h.reservationSvc.ReservationsReturn(ctx, username, reservationUid)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, resp)
}
