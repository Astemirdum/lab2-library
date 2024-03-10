package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/Astemirdum/library-service/pkg/auth"
	"github.com/Astemirdum/library-service/pkg/logger"
	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/time/rate"
)

const (
	AuthorizationHeader = "Authorization"
	bearer              = "Bearer "
)

func JwtAuthentication(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authorization := c.Request().Header.Get(AuthorizationHeader)
		if authorization == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "No Authorization Header")
		}
		if !strings.HasPrefix(authorization, bearer) {
			return echo.NewHTTPError(http.StatusUnauthorized, "Invalid Authorization Header")
		}
		tokenStr := strings.TrimPrefix(authorization, bearer)
		claims := new(auth.Claims)

		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return auth.JWTKey, nil
		})
		if err != nil || !token.Valid {
			return echo.NewHTTPError(http.StatusUnauthorized, "JwtAccessDenied")
		}
		if time.Now().After(claims.ExpiresAt.Time) {
			return echo.NewHTTPError(http.StatusUnauthorized, "TokenExpired")
		}

		req := c.Request()
		ctx := auth.SetAuthContext(req.Context(), claims.Profile.Username, claims.Profile.Role)
		req = req.WithContext(ctx)
		c.SetRequest(req)

		return next(c)
	}
}

func AuthContext(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		req := c.Request()
		userName := req.Header.Get(auth.XUserNameHeader)
		if userName == "" {
			return errors.New("user-name is empty")
		}
		userRole := req.Header.Get(auth.XUserRoleHeader)
		if userRole == "" {
			return errors.New("user-role is empty")
		}
		ctx := auth.SetAuthContext(req.Context(), userName, userRole)
		req = req.WithContext(ctx)
		c.SetRequest(req)
		return next(c)
	}
}

func NewRateLimiter(rps rate.Limit) echo.MiddlewareFunc {
	return middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(rps))
}

func RequestLoggerConfig() middleware.RequestLoggerConfig {
	cfg := logger.Log{LogLevel: zapcore.DebugLevel, Sink: ""}
	log := logger.NewLogger(cfg, "echo")
	c := middleware.RequestLoggerConfig{
		LogURI:       true,
		LogStatus:    true,
		HandleError:  true,
		LogError:     true,
		LogLatency:   true,
		LogRequestID: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			level := zapcore.InfoLevel
			if v.Error != nil {
				level = zapcore.ErrorLevel
			}
			log.Log(level, "request",
				zap.String("URI", v.URI),
				zap.String("Method", v.Method),
				zap.Int("status", v.Status),
				zap.Duration("latency", v.Latency),
				zap.Error(v.Error),
				zap.String("request_id", v.RequestID),
			)
			return nil
		},
	}
	return c
}
