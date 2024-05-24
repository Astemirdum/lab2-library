package auth

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/MicahParks/keyfunc"
	"github.com/golang-jwt/jwt/v4"
)

const (
	bearer = "Bearer "
)

func NewJWKs(rawJWKS string) *keyfunc.JWKS {
	jwksJSON := json.RawMessage(rawJWKS)
	jwks, err := keyfunc.NewJSON(jwksJSON)
	if err != nil {
		panic(err)
	}
	return jwks
}

var DefaultKeyFunc = func(token *jwt.Token) (interface{}, error) { return JWTKey, nil }

func RetrieveToken(reqToken string, keyFunc func(*jwt.Token) (interface{}, error)) (*Token, error) {
	if len(reqToken) == 0 {
		return nil, fmt.Errorf("token is missing")
	}
	if !strings.HasPrefix(reqToken, bearer) {
		return nil, fmt.Errorf("need bearer token")
	}
	var tk = new(Token)
	token, err := jwt.ParseWithClaims(strings.TrimPrefix(reqToken, bearer), tk, keyFunc)
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("jwt access denied %v", err)
	}
	if time.Now().After(tk.ExpiresAt.Time) {
		return nil, fmt.Errorf("token is expired")
	}

	return tk, nil
}
