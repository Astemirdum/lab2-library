package handler

import (
	"encoding/json"
	"github.com/Astemirdum/library-service/pkg/logger"

	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/time/rate"

	"context"

	jwtmiddleware "github.com/auth0/go-jwt-middleware/v2"
	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"

	"log"
)

// CustomClaims contains custom data we want from the token.
type CustomClaims struct {
	Scope    string `json:"scope"`
	UserName string `json:"username"`
}

func (c CustomClaims) Validate(ctx context.Context) error {
	return nil
}

// EnsureValidToken is a middleware that will check the validity of our JWT.
func EnsureValidToken() func(next http.Handler) http.Handler {
	issuerURL, err := url.Parse("https://" + os.Getenv("AUTH0_DOMAIN") + "/")
	if err != nil {
		log.Fatalf("Failed to parse the issuer url: %v", err)
	}
	provider := jwks.NewCachingProvider(issuerURL, 5*time.Minute)

	jwtValidator, err := validator.New(
		provider.KeyFunc,
		validator.RS256,
		issuerURL.String(),
		[]string{os.Getenv("AUTH0_AUDIENCE")},
		validator.WithCustomClaims(
			func() validator.CustomClaims {
				return &CustomClaims{}
			},
		),
		validator.WithAllowedClockSkew(time.Minute),
	)
	if err != nil {
		log.Fatalf("Failed to set up the jwt validator")
	}

	errorHandler := func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("Encountered error while validating JWT: %v", err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"Failed to validate JWT."}`))
	}

	middleware := jwtmiddleware.New(
		jwtValidator.ValidateToken,
		jwtmiddleware.WithErrorHandler(errorHandler),
	)

	return func(next http.Handler) http.Handler {
		return middleware.CheckJWT(next)
	}
}

// HasScope checks whether our claims have a specific scope.
func (c CustomClaims) HasScope(expectedScope string) bool {
	// openid profile email
	result := strings.Split(c.Scope, " ")
	for i := range result {
		if result[i] == expectedScope {
			return true
		}
	}

	return false
}

func withScop(w http.ResponseWriter, r *http.Request) { //nolint:unused
	//router.Handle("/api/private", middleware.EnsureValidToken()(
	//		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	//			w.Header().Set("Content-Type", "application/json")
	//			w.WriteHeader(http.StatusOK)
	//			w.Write([]byte(`{"message":"Hello from a private endpoint! You need to be authenticated to see this."}`))
	//		}),
	//	))
	w.Header().Set("Content-Type", "application/json")

	token, ok := r.Context().Value(jwtmiddleware.ContextKey{}).(*validator.ValidatedClaims)
	if !ok {
		http.Error(w, "failed to get validated claims", http.StatusInternalServerError)
		return
	}
	claims, ok := token.CustomClaims.(*CustomClaims)
	if !ok {
		panic("token.CustomClaims.(*CustomClaims)")
	}
	if !claims.HasScope("read:messages") {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"message":"Insufficient scope."}`))
		return
	}

	w.WriteHeader(http.StatusOK)

	payload, err := json.Marshal(claims)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte(`{"message":"Hello from a private endpoint! You need to be authenticated to see this."}`))
	w.Write(payload)
}

//type userKeyType int
//const userNameKey userKeyType = 1

const userNameKeyString = "userNameKey"

func AuthenticateMW(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := strings.TrimPrefix(c.Request().Header.Get(AuthorizationHeader), Bearer)
		if token == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "no token in Authorization header")
		}

		//TODO: validate
		if token == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "token is invalid")
		}
		if token == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "token is expired")
		}

		//TODO: extract user from token
		username := ""

		//ctx := context.WithValue(c.Request().Context(), userKey, username)
		//c.SetRequest(c.Request().WithContext(ctx))
		c.Set(userNameKeyString, username)

		return next(c)
	}
}

func extractUserName(c echo.Context) (string, error) {
	username, ok := c.Get(userNameKeyString).(string)
	if !ok {
		return "", errors.New("invalid userNameKeyString")
	}
	return username, nil
}

func requestLoggerConfig() middleware.RequestLoggerConfig {
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

func newRateLimiterMW(rps rate.Limit) echo.MiddlewareFunc {
	return middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(rps))
}
