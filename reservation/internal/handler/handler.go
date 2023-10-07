package handler

import (
	"github.com/Astemirdum/library-service/reservation/internal/model"
	"net/http"
	"time"

	"github.com/Astemirdum/library-service/pkg/validate"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
	"go.uber.org/zap"

	_ "github.com/Astemirdum/library-service/swagger"
)

type Handler struct {
	reservationSvc LibraryService
	client         *http.Client
	log            *zap.Logger
}

func New(reservationSrv LibraryService, log *zap.Logger) *Handler {
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
	base.GET("/swagger/*", echoSwagger.WrapHandler)

	e.Validator = validate.NewCustomValidator()
	api := e.Group("/api/v1",
		middleware.RequestLoggerWithConfig(requestLoggerConfig()),
		middleware.RequestID(),
		newRateLimiterMW(apiRPS),
	)

	api.GET("/reservations", h.GetReservation)
	api.POST("/reservations", h.CreateReservation)
	api.POST("/reservations/{reservationUid}/return", h.ReservationReturn)

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
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) GetReservation(c echo.Context) error {
	ctx := c.Request().Context()
	err := h.reservationSvc.GetReservation(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	// c.Response().Header().Set("Location", fmt.Sprintf("/api/v1/librarys/%d", id))

	return c.JSON(http.StatusOK, nil)
}

func (h *Handler) ReservationReturn(c echo.Context) error {
	return nil
}
