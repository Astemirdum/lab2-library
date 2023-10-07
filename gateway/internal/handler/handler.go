package handler

import (
	"github.com/Astemirdum/library-service/gateway/config"
	"github.com/Astemirdum/library-service/gateway/internal/errs"
	"github.com/Astemirdum/library-service/gateway/internal/model"
	"github.com/Astemirdum/library-service/gateway/internal/service/library"
	"github.com/Astemirdum/library-service/gateway/internal/service/rating"
	"github.com/Astemirdum/library-service/gateway/internal/service/reservation"
	"github.com/Astemirdum/library-service/pkg/validate"
	_ "github.com/Astemirdum/library-service/swagger"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pkg/errors"
	echoSwagger "github.com/swaggo/echo-swagger"
	"go.uber.org/zap"
	"net/http"
)

type Handler struct {
	librarySvc     LibraryService
	ratingSvc      RatingService      //nolint:unused
	reservationSvc ReservationService //nolint:unused
	//client         *http.Client
	log *zap.Logger
}

func New(log *zap.Logger, cfg config.Config) *Handler {
	h := &Handler{
		librarySvc:     library.NewService(log, cfg),
		ratingSvc:      rating.NewService(log, cfg),
		reservationSvc: reservation.NewService(log, cfg),
		// client:         &http.Client{Timeout: time.Minute},
		log: log,
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

	api.GET("/rating", h.GetRating)

	api.GET("/libraries", h.GetLibraries)
	api.GET("/libraries/:libraryUid/books", h.GetBooks)

	api.POST("/reservations", h.CreateReservation)
	api.GET("/reservations", h.GetReservation)
	api.POST("/reservations/{reservationUid}/return", h.ReservationReturn)

	return e
}

func (h *Handler) Health(c echo.Context) error {
	return c.String(http.StatusOK, "OK")
}

func (h *Handler) GetReservation(c echo.Context) error {
	data, code, err := h.reservationSvc.GetReservation(c)
	if err != nil {
		return echo.NewHTTPError(code, err.Error())
	}
	return c.JSONBlob(code, data)
}

const (
	XUserName = "X-User-Name"
)

func (h *Handler) CreateReservation(c echo.Context) error {
	var createReservationRequest model.CreateReservationRequest
	if err := c.Bind(&createReservationRequest); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	createReservationRequest.UserName = c.Request().Header.Get(XUserName)
	if err := c.Validate(createReservationRequest); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	ctx := c.Request().Context()
	lib, code, err := h.librarySvc.GetLibrary(ctx, createReservationRequest.LibraryUid)
	if err != nil {
		return echo.NewHTTPError(code, err.Error())
	}

	book, code, err := h.librarySvc.GetBook(ctx, createReservationRequest.LibraryUid, createReservationRequest.BookUid)
	if err != nil {
		return echo.NewHTTPError(code, err.Error())
	}

	rat, code, err := h.ratingSvc.GetRating(ctx, createReservationRequest.UserName)
	if err != nil {
		return echo.NewHTTPError(code, err.Error())
	}

	rsv, code, err := h.reservationSvc.CreateReservation(ctx, createReservationRequest)
	if err != nil {
		return echo.NewHTTPError(code, err.Error())
	}

	return c.JSON(http.StatusOK, model.CreateReservationResponse{
		Reservation: rsv,
		Library:     lib,
		Book:        book,
		Rating:      rat,
	})
}

func (h *Handler) ReservationReturn(c echo.Context) error {
	data, code, err := h.reservationSvc.ReservationReturn(c)
	if err != nil {
		return echo.NewHTTPError(code, err.Error())
	}
	return c.JSONBlob(code, data)
}

func (h *Handler) GetBooks(c echo.Context) error {
	data, code, err := h.librarySvc.GetBooks(c)
	if err != nil {
		return echo.NewHTTPError(code, err.Error())
	}
	return c.JSONBlob(code, data)
}

func (h *Handler) GetRating(c echo.Context) error {
	resp, code, err := h.ratingSvc.GetRating(c.Request().Context(), c.Request().Header.Get(XUserName))
	if err != nil {
		return echo.NewHTTPError(code, err.Error())
	}
	return c.JSON(code, resp)
}

func (h *Handler) GetLibraries(c echo.Context) error {
	data, code, err := h.librarySvc.GetLibraries(c)
	if err != nil {
		return echo.NewHTTPError(code, err.Error())
	}
	return c.JSONBlob(code, data)
}

func (h *Handler) CreateLibrary(c echo.Context) error {
	var pers model.Library
	if err := c.Bind(&pers); err != nil {
		return httpValidationError(c, http.StatusBadRequest, err)
	}
	if err := c.Validate(pers); err != nil {
		return httpValidationError(c, http.StatusBadRequest, err)
	}
	//ctx := c.Request().Context()
	//id, err := h.reservationSvc.CreateReservation(ctx)
	//if err != nil {
	//	return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	//}
	//c.Response().Header().Set("Location", fmt.Sprintf("/api/v1/librarys/%d", id))

	return c.String(http.StatusCreated, "OK")
}

func httpValidationError(c echo.Context, code int, err error) error {
	c.Response().WriteHeader(code)
	_ = c.JSON(code, &errs.ValidationErrorResponse{ //nolint:errcheck
		Message: err.Error(),
		Errors: struct {
			AdditionalProperties string `json:"additionalProperties"`
		}{
			AdditionalProperties: "",
		},
	})
	return errors.New("")
}
