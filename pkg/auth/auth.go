package auth

import (
	"context"
	"net/http"

	"github.com/golang-jwt/jwt/v4"
	"github.com/pkg/errors"
)

const (
	secretKey       = "my_secret_key"
	XUserNameHeader = "X-User-Name"
	XUserRoleHeader = "X-User-Role"

	adminRole = "admin"
)

type CtxUserKey int

const (
	userNameKey CtxUserKey = 1
	userRoleKey CtxUserKey = 2
)

var JWTKey = []byte(secretKey)

type Claims struct {
	Profile struct {
		Username string `json:"username"`
		Role     string `json:"role"`
	} `json:"profile"`
	OpenID string `json:"openid"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

func (c *Claims) Valid() error {
	valid := c.RegisteredClaims.Valid()
	return valid
}

type Get interface{ Value(any) any }

func IsAdmin(getter Get) bool {
	role, err := GetUserRole(getter)
	if err != nil {
		return false
	}
	return role == adminRole
}

func GetUserName(getter Get) (string, error) {
	userName, ok := getter.Value(userNameKey).(string)
	if !ok {
		return "", errors.New("no user name")
	}
	return userName, nil
}

func GetUserRole(getter Get) (string, error) {
	userRole, ok := getter.Value(userRoleKey).(string)
	if !ok {
		return "", errors.New("no user role")
	}
	return userRole, nil
}

func SetAuthContext(ctx context.Context, userName, userRole string) context.Context {
	ctx = context.WithValue(ctx, userNameKey, userName)
	ctx = context.WithValue(ctx, userRoleKey, userRole)
	return ctx
}

func SetAuthHeader(req *http.Request) {
	ctx := req.Context()
	if name, err := GetUserName(ctx); err == nil {
		req.Header.Add(XUserNameHeader, name)
	}
	if role, err := GetUserRole(ctx); err == nil {
		req.Header.Add(XUserRoleHeader, role)
	}
}
