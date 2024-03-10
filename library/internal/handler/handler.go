package handler

import (
	"net/http"
	"strconv"
	"time"

	md "github.com/Astemirdum/library-service/pkg/middleware"

	"github.com/Astemirdum/library-service/library/internal/errs"
	"github.com/Astemirdum/library-service/pkg/validate"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type Handler struct {
	librarySvc LibraryService
	client     *http.Client
	log        *zap.Logger
}

func New(librarySrv LibraryService, log *zap.Logger) *Handler {
	h := &Handler{
		librarySvc: librarySrv,
		log:        log,
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

	api.GET("/libraries/:libraryUid/books", h.GetBooks)
	api.GET("/libraries/:libraryUid/books/:bookUid", h.GetBook)
	api.PATCH("/libraries/books", h.AvailableCount)

	api.GET("/libraries", h.GetLibraries)
	api.GET("/libraries/:libraryUid", h.GetLibrary)

	return e
}

func (h *Handler) Health(c echo.Context) error {
	return c.String(http.StatusOK, "OK")
}

func (h *Handler) AvailableCount(c echo.Context) error {
	type Req struct {
		LibraryID int  `json:"libraryID"`
		BookID    int  `json:"bookID"`
		IsReturn  bool `json:"isReturn"`
	}
	var req Req
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.librarySvc.AvailableCount(c.Request().Context(), req.LibraryID, req.BookID, req.IsReturn); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusOK)
}

func (h *Handler) GetBook(c echo.Context) error {
	libraryUid := c.Param("libraryUid")
	bookUid := c.Param("bookUid")
	book, err := h.librarySvc.GetBook(c.Request().Context(), libraryUid, bookUid)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, book)
}

func (h *Handler) GetLibrary(c echo.Context) error {
	libraryUid := c.Param("libraryUid")
	lib, err := h.librarySvc.GetLibrary(c.Request().Context(), libraryUid)
	if err != nil {
		if errors.Is(err, errs.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, err.Error())
		}
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, lib)
}

func (h *Handler) GetBooks(c echo.Context) error {
	ctx := c.Request().Context()

	libraryUid := c.Param("libraryUid")
	if libraryUid == "" {
		return echo.NewHTTPError(http.StatusBadRequest, errors.New("empty libraryUid"))
	}
	var (
		err     error
		page    int
		size    int
		showAll bool
	)
	if pageParam := c.QueryParam("page"); pageParam != "" {
		if page, err = strconv.Atoi(pageParam); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, errors.New("page is invalid"))
		}
	}
	if sizeParam := c.QueryParam("size"); sizeParam != "" {
		if size, err = strconv.Atoi(sizeParam); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, errors.New("size is invalid"))
		}
	}
	if showAllParam := c.QueryParam("showAll"); showAllParam != "" {
		if showAll, err = strconv.ParseBool(showAllParam); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, errors.New("showAll is invalid"))
		}
	}

	books, err := h.librarySvc.ListBooks(ctx, libraryUid, showAll, page, size)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, books)
}

func (h *Handler) GetLibraries(c echo.Context) error {
	ctx := c.Request().Context()

	city := c.QueryParam("city")
	if city == "" {
		return echo.NewHTTPError(http.StatusBadRequest, errors.New("city is required"))
	}
	var (
		err  error
		page int
		size int
	)
	if pageParam := c.QueryParam("page"); pageParam != "" {
		if page, err = strconv.Atoi(pageParam); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, errors.New("page is invalid"))
		}
	}
	if sizeParam := c.QueryParam("size"); sizeParam != "" {
		if size, err = strconv.Atoi(sizeParam); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, errors.New("size is invalid"))
		}
	}

	library, err := h.librarySvc.ListLibrary(ctx, city, page, size)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, library)
}
