package openid

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"

	//"github.com/auth0/go-jwt-middleware/v2/validator".
	"github.com/coreos/go-oidc"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

const (
	userNameKeyString = "userNameKey"
	admin             = "Test Max"

	AuthorizationHeader = "Authorization"
	bearer              = "Bearer "
	CookieName          = "access_token"

	// state = "somestate".
)

type providerImpl struct {
	client       Provider // *gocloak.GoCloak
	clientID     string
	clientSecret string
	realm        string
	state        string
	verifier     *oidc.IDTokenVerifier
	oauth2Config oauth2.Config
}

type Config struct {
	Addr string

	ClientID     string
	ClientSecret string

	RedirectURL string // "http://localhost:8080/auth/callback"

	Issuer   string `yaml:"issuer" envconfig:"AUTH0_DOMAIN"` // issuer = "https://accounts.google.com"
	Audience string `yaml:"audience" envconfig:"AUTH0_AUDIENCE"`
}

type Provider interface {
	//TokenIntroSpector
	DecodeAccessToken(ctx context.Context, accessToken string, realm string) (*jwt.Token, *jwt.MapClaims, error)
	Provide(ctx context.Context, state, code string) (AuthorizeResponse, error)
	AuthURL() string
	Middleware() echo.MiddlewareFunc
	//LoginClient(ctx context.Context, clientID string, clientSecret string,
	//	realm string, scopes ...string) (*gocloak.JWT, error)
}

type TokenIntroSpector interface {
	//RetrospectToken(ctx context.Context, accessToken string, clientID string, clientSecret string, realm string) (*gocloak.IntroSpectTokenResult, error)
}

func NewProvider(cfg Config) (*providerImpl, error) {
	const (
		state = "state"
	)
	provider, err := oidc.NewProvider(context.Background(), cfg.Issuer)
	if err != nil {
		return nil, err
	}
	oauth2Config := oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	verifier := provider.Verifier(&oidc.Config{ClientID: cfg.ClientID})
	return &providerImpl{
		//client:       client,
		clientID:     cfg.ClientID,
		clientSecret: cfg.ClientSecret,
		//realm:        cfg.Realm,
		state:        state,
		verifier:     verifier,
		oauth2Config: oauth2Config,
	}, nil
}

//func (a *providerImpl) Auth(ctx context.Context, clientID string, clientSecret string,
//	realm string, scopes ...string) (*gocloak.JWT, error) {
//	return a.client.LoginClient(ctx, clientID, clientSecret, realm, scopes...)
//	//http.Redirect(w, r, oauth2Config.AuthCodeURL(state), http.StatusFound)
//
//}
//
//func (a *providerImpl) TokenIntroSpector(ctx context.Context, accessToken string,
//	clientID string, clientSecret string, realm string) (*gocloak.IntroSpectTokenResult, error) {
//	return a.client.RetrospectToken(ctx, accessToken, clientID, clientSecret, realm)
//}

func ValidateToken() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Validate token

			// authorization := c.Request().Header.Get(AuthorizationHeader)
			//validate()
			return next(c)
		}
	}
}

func (p *providerImpl) Middleware() echo.MiddlewareFunc {

	//provider := jwks.NewCachingProvider(issuerURL, time.Minute*5)
	//jwtValidator, err := validator.New(
	//	provider.KeyFunc,
	//	validator.RS256,
	//	issuerURL.String(),
	//	[]string{cfg.Audience},
	//	//validator.WithCustomClaims(func() validator.CustomClaims { return &CustomClaims{} }),
	//	validator.WithAllowedClockSkew(time.Minute),
	//)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := c.Request().Context()
			authorization := c.Request().Header.Get(AuthorizationHeader)
			if authorization == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "No Authorization Header")
			}
			if !strings.HasPrefix(authorization, bearer) {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid Authorization Header")
			}

			token := strings.TrimPrefix(authorization, bearer)

			//claims, err := jwtValidator.ValidateToken(c.Request().Context(), token)
			//if err != nil {
			//	return echo.NewHTTPError(http.StatusUnauthorized, "Invalid Token")
			//}

			//http.Redirect(w, r, a.oauth2Config.AuthCodeURL(a.state), http.StatusUnauthorized)

			//res, err := p.client.RetrospectToken(ctx, token, p.clientID, p.clientSecret, p.realm)
			//if err != nil {
			//	slog.Error("invalid token", "error", err, "token", token)
			//	return echo.NewHTTPError(http.StatusUnauthorized, "retrospect token", err)
			//}

			//if !*res.Active {
			//	slog.Error("invalid token", "error", err, "token", token)
			//	return echo.NewHTTPError(http.StatusUnauthorized, "in active token", err)
			//}
			_, claims, err := p.client.DecodeAccessToken(ctx, token, p.realm)
			if err != nil {
				slog.Error("DecodeAccessToken", "error", err)
				return echo.NewHTTPError(http.StatusUnauthorized, "decode AccessToken", err)
			}

			userID, name, email := NewJwtHelper(*claims).GetUser()
			slog.Info("AuthenticateMW",
				slog.String("userID", userID),
				slog.String("name", name),
				slog.String("email", email))

			c.Set("claims", claims)
			c.Set(userNameKeyString, admin)

			return next(c)
		}
	}
}

func (p *providerImpl) AuthURL() string {
	return p.oauth2Config.AuthCodeURL(p.state)
}

func (p *providerImpl) Provide(ctx context.Context, state, code string) (AuthorizeResponse, error) {
	if p.state != state {
		return AuthorizeResponse{}, errors.New("state did not match")
	}
	oauth2Token, err := p.oauth2Config.Exchange(ctx, code)
	if err != nil {
		return AuthorizeResponse{}, errors.New("failed to exchange token: " + err.Error())
	}
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		return AuthorizeResponse{}, errors.New("no id_token field in oauth2 token")
	}
	idToken, err := p.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return AuthorizeResponse{}, errors.New("failed to verify ID Token: " + err.Error())
	}

	resp := AuthorizeResponse{
		OAuth2Token:   oauth2Token,
		IDTokenClaims: new(json.RawMessage),
	}

	if err := idToken.Claims(&resp.IDTokenClaims); err != nil {
		return AuthorizeResponse{}, err
	}
	return resp, nil
}

type AuthorizeResponse struct {
	OAuth2Token   *oauth2.Token
	IDTokenClaims *json.RawMessage
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
