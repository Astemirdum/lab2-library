package handler

import (
	"net/http"
	"time"

	"github.com/Astemirdum/library-service/backend/pkg/auth"
	md "github.com/Astemirdum/library-service/backend/pkg/middleware"

	"github.com/Astemirdum/library-service/backend/pkg/validate"
	"github.com/Astemirdum/library-service/backend/reservation/internal/errs"
	"github.com/Astemirdum/library-service/backend/reservation/internal/model"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pkg/errors"
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

	base := e.Group("", md.NewRateLimiter(baseRPS))
	base.GET("/manage/health", h.Health)

	e.Validator = validate.NewCustomValidator()
	api := e.Group("/api/v1",
		middleware.RequestLoggerWithConfig(md.RequestLoggerConfig()),
		middleware.RequestID(),
		md.NewRateLimiter(apiRPS),
	)
	api.POST("/reservations/rollback", h.RollbackReservation)

	api = api.Group("", md.AuthContext)
	api.GET("/reservations", h.GetReservations)
	api.POST("/reservations", h.CreateReservation)
	api.POST("/reservations/:reservationUid/return", h.ReservationsReturn)

	return e
}

func (h *Handler) Health(c echo.Context) error {
	return c.String(http.StatusOK, "OK")
}

func (h *Handler) CreateReservation(c echo.Context) error {
	ctx := c.Request().Context()
	var req model.CreateReservationRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	userName, err := auth.GetUserName(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}
	req.UserName = userName

	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

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
	userName, err := auth.GetUserName(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}
	rsv, err := h.reservationSvc.GetReservations(ctx, userName)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, rsv)
}

func (h *Handler) ReservationsReturn(c echo.Context) error {
	ctx := c.Request().Context()
	userName, err := auth.GetUserName(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}
	reservationUid := c.Param("reservationUid")
	if reservationUid == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "reservationUid is empty")
	}
	var req model.ReservationReturnRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	resp, err := h.reservationSvc.ReservationsReturn(ctx, userName, reservationUid)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) RollbackReservation(c echo.Context) error {
	ctx := c.Request().Context()
	type req struct {
		Uuid string `json:"uuid"`
	}
	var r req
	if err := c.Bind(&r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.reservationSvc.RollbackReservation(ctx, r.Uuid); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusOK)
}
