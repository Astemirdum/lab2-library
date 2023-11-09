package handler

import (
	"net/http"

	"github.com/Astemirdum/library-service/pkg/kafka"
	"github.com/IBM/sarama"

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
	echoSwagger "github.com/swaggo/echo-swagger"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type Handler struct {
	librarySvc     LibraryService
	ratingSvc      RatingService
	reservationSvc ReservationService
	enqueuer       Enqueuer
	log            *zap.Logger
}

func New(log *zap.Logger, cfg config.Config, producer sarama.SyncProducer) *Handler {
	h := &Handler{
		librarySvc:     library.NewService(log, cfg),
		ratingSvc:      rating.NewService(log, cfg),
		reservationSvc: reservation.NewService(log, cfg),
		enqueuer:       NewEnqueuer(producer),
		log:            log,
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
	api.GET("/authorize", h.Authorize)
	api.GET("/callback", h.Callback)
	api = api.Group("", authMW)

	api.GET("/rating", h.GetRating)

	api.GET("/libraries", h.GetLibraries)
	api.GET("/libraries/:libraryUid/books", h.GetBooks)

	api.POST("/reservations", h.CreateReservation)
	api.GET("/reservations", h.GetReservations)
	api.POST("/reservations/:reservationUid/return", h.ReservationReturn)

	return e
}

func (h *Handler) Health(c echo.Context) error {
	return c.String(http.StatusOK, "OK")
}

func (h *Handler) Authorize(c echo.Context) error {
	return nil
}

func (h *Handler) Callback(c echo.Context) error {
	return nil
}

func (h *Handler) GetReservations(c echo.Context) error {
	userName := c.Request().Header.Get(XUserName)
	if userName == "" {
		return echo.NewHTTPError(http.StatusBadRequest, errs.ErrUserName)
	}
	ctx := c.Request().Context()

	var reserves []model.GetReservation
	if err := h.reservationSvc.CB().Call(func() error {
		list, code, err := h.reservationSvc.GetReservation(ctx, userName)
		if err != nil {
			return echo.NewHTTPError(code, err.Error())
		}
		reserves = list
		return nil
	}); err != nil {
		return err
	}

	gg, ctx := errgroup.WithContext(ctx)
	libs := make([]model.Library, 0, len(reserves))
	gg.Go(func() error {
		for _, reserve := range reserves {
			if err := h.librarySvc.CB().Call(func() error {
				lib, code, err := h.librarySvc.GetLibrary(ctx, reserve.LibraryUid)
				if err != nil {
					return echo.NewHTTPError(code, err.Error())
				}
				libs = append(libs, lib.Library)
				return nil
			}); err != nil {
				return err
			}
		}
		return nil
	})
	books := make([]model.Book, 0, len(reserves))
	gg.Go(func() error {
		for _, reserve := range reserves {
			if err := h.librarySvc.CB().Call(func() error {
				book, code, err := h.librarySvc.GetBook(ctx, reserve.LibraryUid, reserve.BookUid)
				if err != nil {
					return echo.NewHTTPError(code, err.Error())
				}
				books = append(books, book.Book)
				return nil
			}); err != nil {
				return err
			}
		}
		return nil
	})

	if err := gg.Wait(); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, getReservationResponse(reserves, books, libs))
}

func getReservationResponse(reserves []model.GetReservation, books []model.Book, libs []model.Library) []model.GetReservationResponse {
	items := make([]model.GetReservationResponse, 0, len(reserves))
	for i := range reserves {
		items = append(items, model.GetReservationResponse{
			Reservation: model.Reservation{
				ReservationUid: reserves[i].ReservationUid,
				Status:         reserves[i].Status,
				StartDate:      reserves[i].StartDate,
				TillDate:       reserves[i].TillDate,
			},
			Library: libs[i],
			Book:    books[i],
		})
	}
	return items
}

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
	var (
		lib  model.GetLibrary
		book model.GetBook
		rat  model.Rating
		code int
		err  error
	)
	gg, ctxCancel := errgroup.WithContext(ctx)
	gg.Go(func() error {
		return h.librarySvc.CB().Call(func() error {
			lib, code, err = h.librarySvc.GetLibrary(ctxCancel, createReservationRequest.LibraryUid)
			if err != nil {
				return echo.NewHTTPError(code, err.Error())
			}
			return nil
		})
	})

	gg.Go(func() error {
		return h.librarySvc.CB().Call(func() error {
			book, code, err = h.librarySvc.GetBook(ctxCancel, createReservationRequest.LibraryUid, createReservationRequest.BookUid)
			if err != nil {
				return echo.NewHTTPError(code, err.Error())
			}
			return nil
		})
	})

	gg.Go(func() error {
		return h.ratingSvc.CB().Call(func() error {
			rat, code, err = h.ratingSvc.GetRating(ctxCancel, createReservationRequest.UserName)
			if err != nil {
				return echo.NewHTTPError(code, err.Error())
			}
			return nil
		})
	})

	if err := gg.Wait(); err != nil {
		return err
	}
	createReservationRequest.Stars = rat.Stars
	rsv, code, err := h.reservationSvc.CreateReservation(ctx, createReservationRequest)
	if err != nil {
		return echo.NewHTTPError(code, err.Error())
	}

	if code, err := h.librarySvc.AvailableCount(ctx, model.AvailableCountRequest{
		LibraryID: lib.ID,
		BookID:    book.ID,
		IsReturn:  false,
	}); err != nil {
		if _, err := h.reservationSvc.RollbackReservation(ctx, rsv.ReservationUid); err != nil {
			h.log.Warn("RollbackReservation")
		}
		return echo.NewHTTPError(code, err.Error())
	}

	return c.JSON(http.StatusOK, model.CreateReservationResponse{
		ReservationUid: rsv.ReservationUid,
		Status:         rsv.Status,
		StartDate:      model.Date2{Time: rsv.StartDate},
		TillDate:       model.Date2{Time: rsv.TillDate},
		Library:        lib.Library,
		Book:           book.Book,
		Rating:         rat,
	})
}

func (h *Handler) ReservationReturn(c echo.Context) error {
	ctx := c.Request().Context()
	username := c.Request().Header.Get(XUserName)
	if username == "" {
		return echo.NewHTTPError(http.StatusBadRequest, errs.ErrUserName)
	}
	reservationUid := c.Param("reservationUid")
	var req model.ReservationReturnRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	res, code, err := h.reservationSvc.ReservationReturn(ctx, req, username, reservationUid)
	if err != nil {
		return echo.NewHTTPError(code, err.Error())
	}
	var (
		lib  model.GetLibrary
		book model.GetBook
	)
	gg, ctxCancel := errgroup.WithContext(ctx)
	gg.Go(func() error {
		return h.librarySvc.CB().Call(func() error {
			lib, code, err = h.librarySvc.GetLibrary(ctxCancel, res.LibraryUid)
			if err != nil {
				return echo.NewHTTPError(code, err.Error())
			}
			return nil
		})
	})

	gg.Go(func() error {
		return h.librarySvc.CB().Call(func() error {
			book, code, err = h.librarySvc.GetBook(ctxCancel, res.LibraryUid, res.BookUid)
			if err != nil {
				return echo.NewHTTPError(code, err.Error())
			}
			return nil
		})
	})
	if err := gg.Wait(); err != nil {
		return err
	}

	availableCountReq := model.AvailableCountRequest{
		LibraryID: lib.ID,
		BookID:    book.ID,
		IsReturn:  true,
	}
	if code, err := h.librarySvc.AvailableCount(ctx, availableCountReq); err != nil {
		if code == http.StatusServiceUnavailable {
			if err := h.enqueuer.Enqueue(kafka.LibraryTopic, availableCountReq); err != nil {
				h.log.Warn("AvailableCount h.enqueuer.Enqueue()", zap.Error(err))
			}
		} else {
			return echo.NewHTTPError(code, err.Error())
		}
	}

	stars := 1
	if book.Condition != req.Condition {
		stars = -10
	}

	if code, err := h.ratingSvc.Rating(ctx, username, stars); err != nil {
		if code == http.StatusServiceUnavailable {
			ratingMsg := model.RatingMsg{
				Name:  username,
				Stars: stars,
			}
			//h.enqueuer.EnqueueV2(ctx, h.ratingSvc.Rating, ratingMsg)
			if err := h.enqueuer.Enqueue(kafka.RatingTopic, ratingMsg); err != nil {
				h.log.Warn("Rating h.enqueuer.Enqueue()", zap.Error(err))
			}
		} else {
			return echo.NewHTTPError(code, err.Error())
		}
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) GetBooks(c echo.Context) error {
	var (
		code int
		data []byte
	)
	if err := h.librarySvc.CB().Call(func() error {
		var err error
		data, code, err = h.librarySvc.GetBooks(c)
		if err != nil {
			return echo.NewHTTPError(code, err.Error())
		}
		return nil
	}); err != nil {
		return err
	}

	return c.JSONBlob(code, data)
}

func (h *Handler) GetLibraries(c echo.Context) error {
	var (
		code int
		data []byte
	)
	if err := h.librarySvc.CB().Call(func() error {
		var err error
		data, code, err = h.librarySvc.GetLibraries(c)
		if err != nil {
			return echo.NewHTTPError(code, err.Error())
		}
		return nil
	}); err != nil {
		return err
	}

	return c.JSONBlob(code, data)
}

func (h *Handler) GetRating(c echo.Context) error {
	var (
		code int
		resp model.Rating
	)
	if err := h.ratingSvc.CB().Call(func() error {
		var err error
		resp, code, err = h.ratingSvc.GetRating(c.Request().Context(), c.Request().Header.Get(XUserName))
		if err != nil {
			return echo.NewHTTPError(code, err.Error())
		}
		return nil
	}); err != nil {
		return err
	}

	return c.JSON(code, resp)
}
