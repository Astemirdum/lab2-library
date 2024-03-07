package auth0

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

const (
	userNameKeyString = "userNameKey"
	admin             = "Test Max"

	claimsKey = "claims"
	XUserName = "X-User-Name"

	AuthorizationHeader = "Authorization"
	bearer              = "Bearer "
)

type (
	Config struct {
		Issuer   string `yaml:"issuer" envconfig:"AUTH0_DOMAIN"`
		Audience string `yaml:"audience" envconfig:"AUTH0_AUDIENCE"`
		Enable   bool   `yaml:"enable" envconfig:"AUTH0_ENABLE"`
	}
)

type Validator interface {
	ValidateToken(ctx context.Context, tokenString string) (interface{}, error)
}

type validatorMock struct{}

func (v *validatorMock) ValidateToken(_ context.Context, _ string) (interface{}, error) {
	return new(validator.ValidatedClaims), nil
}

func NewValidator(cfg Config) (Validator, error) {
	if !cfg.Enable {
		return new(validatorMock), nil
	}
	issuerURL, err := url.Parse("https://" + cfg.Issuer + "/")
	if err != nil {
		return nil, fmt.Errorf("failed to parse the issuer url: %v", err)
	}
	provider := jwks.NewCachingProvider(issuerURL, time.Minute*5)
	jwtValidator, err := validator.New(
		provider.KeyFunc,
		validator.RS256,
		issuerURL.String(),
		[]string{cfg.Audience},
		validator.WithCustomClaims(func() validator.CustomClaims { return &CustomClaims{} }),
		validator.WithAllowedClockSkew(time.Minute),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to set up the jwt validator: %v", err)
	}
	return jwtValidator, nil
}

func MiddlewareUserName(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		userName := c.Request().Header.Get(XUserName)
		if userName == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, "username is empty")
		}
		c.Set(userNameKeyString, userName)
		return next(c)
	}
}

func Middleware(jwtValidator Validator) echo.MiddlewareFunc {
	//errorHandler := func(w http.ResponseWriter, r *http.Request, err error) {
	//	log.Printf("Encountered error while validating JWT: %v", err)
	//
	//	w.Header().Set("Content-Type", "application/json")
	//	w.WriteHeader(http.StatusUnauthorized)
	//	w.Write([]byte(`{"message":"Failed to validate JWT."}`))
	//}
	//middleware := jwtmiddleware.New(
	//	jwtValidator.ValidateToken,
	//	jwtmiddleware.WithErrorHandler(errorHandler),
	//)
	//return func(next http.Handler) http.Handler {
	//	return middleware.CheckJWT(next)
	//}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authorization := c.Request().Header.Get(AuthorizationHeader)
			if authorization == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "No Authorization Header")
			}
			if !strings.HasPrefix(authorization, bearer) {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid Authorization Header")
			}

			token := strings.TrimPrefix(authorization, bearer)

			claims, err := jwtValidator.ValidateToken(c.Request().Context(), token)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid Token")
			}

			c.Set(claimsKey, claims) // .(*validator.ValidatedClaims)
			c.Set(userNameKeyString, admin)

			return next(c)
		}
	}
}

// CustomClaims contains custom data we want from the token.
type CustomClaims struct {
	Scope string `json:"scope"`
}

func (c CustomClaims) Validate(ctx context.Context) error {
	//TODO:
	return nil
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

type Get interface {
	Get(string) any
}

func IsAdmin(userName string) bool {
	return userName == admin
}

func GetUserName(getter Get) (string, error) {
	userName, ok := getter.Get(userNameKeyString).(string)
	if !ok {
		return "", errors.New("no username")
	}
	return userName, nil
}

func GetClaims(c echo.Context) (*validator.ValidatedClaims, error) {
	claims, ok := c.Get(claimsKey).(*validator.ValidatedClaims)
	if !ok {
		return nil, errors.New("no claims")
	}
	return claims, nil
}

/*
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
*/
