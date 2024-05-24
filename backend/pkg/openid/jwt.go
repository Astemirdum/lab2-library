package openid

import (
	"fmt"
	"net/http"
	"strings"

	jwt "github.com/golang-jwt/jwt/v4"
)

type contextKey int

const (
	ContextKeyClaims contextKey = iota + 1
)

type JwtHelper struct {
	claims       jwt.MapClaims
	realmRoles   []string
	accountRoles []string
	scopes       []string
}

func NewJwtHelper(claims jwt.MapClaims) *JwtHelper {
	return &JwtHelper{
		claims:       claims,
		realmRoles:   parseRealmRoles(claims),
		accountRoles: parseAccountRoles(claims),
		scopes:       parseScopes(claims),
	}
}

func GetKeyCloakToken(r *http.Request) (string, error) {
	cookie, err := r.Cookie(CookieName)
	if err != nil {
		return "", err
	}

	return cookie.Value, nil
}

func (j *JwtHelper) GetUserID() string {
	if sid, ok := j.claims["sid"].(string); ok {
		return sid
	}
	return ""
}

func (j *JwtHelper) GetName() string {
	if name, ok := j.claims["name"].(string); ok {
		return name
	}
	return ""
}

func (j *JwtHelper) GetEmail() string {
	if email, ok := j.claims["email"].(string); ok {
		return email
	}
	return ""
}

func (j *JwtHelper) GetUser() (sid, name, email string) {
	return j.GetUserID(), j.GetName(), j.GetEmail()
}

func (j *JwtHelper) IsUserInRealmRole(role string) bool {
	return contains(j.realmRoles, role)
}

func (j *JwtHelper) TokenHasScope(scope string) bool {
	return contains(j.scopes, scope)
}

func parseRealmRoles(claims jwt.MapClaims) []string {
	realmRoles := make([]string, 0)

	if claim, ok := claims["realm_access"]; ok {
		if roles, ok := claim.(map[string]interface{})["roles"]; ok {
			for _, role := range roles.([]interface{}) { //nolint:forcetypeassert
				realmRoles = append(realmRoles, role.(string)) //nolint:forcetypeassert
			}
		}
	}

	return realmRoles
}

func parseAccountRoles(claims jwt.MapClaims) []string {
	var accountRoles []string

	if claim, ok := claims["resource_access"]; ok {
		if acc, ok := claim.(map[string]interface{})["account"]; ok {
			if roles, ok := acc.(map[string]interface{})["roles"]; ok {
				for _, role := range roles.([]interface{}) { //nolint:forcetypeassert
					accountRoles = append(accountRoles, role.(string)) //nolint:forcetypeassert
				}
			}
		}
	}

	return accountRoles
}

func parseScopes(claims jwt.MapClaims) []string {
	scopeStr, err := parseString(claims, "scope")
	if err != nil {
		return make([]string, 0)
	}
	scopes := strings.Split(scopeStr, " ")

	return scopes
}

func parseString(claims jwt.MapClaims, key string) (string, error) {
	var (
		ok  bool
		raw interface{}
		iss string
	)
	raw, ok = claims[key]
	if !ok {
		return "", nil
	}

	iss, ok = raw.(string)
	if !ok {
		return "", fmt.Errorf("key %s is invalid", key)
	}

	return iss, nil
}

func contains(arr []string, s string) bool {
	for i := range arr {
		if arr[i] == s {
			return true
		}
	}

	return false
}
