package mdlw

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	ctxutil "github.com/RyanBard/go-ctx-util/pkg"
	logutil "github.com/RyanBard/go-log-util/pkg"
	"github.com/RyanBard/go-service-ex/internal/config"
	"github.com/RyanBard/go-service-ex/internal/testutil"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
)

var (
	cfg = config.AuthConfig{
		JWTSecret:   "jwt-secret",
		JWTAudience: "jwt-audience",
		JWTIssuer:   "jwt-issuer",
	}
)

const (
	nonAdminUserID = "non-admin-id"
	adminUserID    = "admin-id"
)

func basic(token string) string {
	return fmt.Sprintf("Basic %s", token)
}

func bearer(token string) string {
	return fmt.Sprintf("Bearer %s", token)
}

func validClaims(userID string) jwt.MapClaims {
	return jwt.MapClaims{
		"sub": userID,
		"aud": cfg.JWTAudience,
		"iss": cfg.JWTIssuer,
		"exp": time.Now().AddDate(0, 0, 1).Unix(),
	}
}

func hmacJWT(claims jwt.MapClaims) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		testutil.GetLogger().With(logutil.LogAttrError(err)).Error("failed to sign jwt")
		panic(err)
	}
	return tokenStr
}

func validNonAdminJWT() string {
	return hmacJWT(validClaims(nonAdminUserID))
}

func validNonAdminExplicitFalseJWT() string {
	claims := validClaims(nonAdminUserID)
	claims["admin"] = false
	return hmacJWT(claims)
}

func validAdminJWT() string {
	claims := validClaims(adminUserID)
	claims["admin"] = true
	return hmacJWT(claims)
}

func invalidExpiredJWT() string {
	claims := validClaims(adminUserID)
	claims["exp"] = time.Now().AddDate(0, 0, -1).Unix()
	return hmacJWT(claims)
}

func invalidIssuerJWT() string {
	claims := validClaims(adminUserID)
	claims["iss"] = "wrong"
	return hmacJWT(claims)
}

func invalidAudienceJWT() string {
	claims := validClaims(adminUserID)
	claims["aud"] = "wrong"
	return hmacJWT(claims)
}

func invalidHMACSigningMethodJWT() string {
	claims := validClaims(adminUserID)
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenStr, err := token.SignedString([]byte(cfg.JWTSecret))
	if err != nil {
		testutil.GetLogger().With(logutil.LogAttrError(err)).Error("failed to sign jwt")
		panic(err)
	}
	return tokenStr
}

func invalidRSAJWT() string {
	claims := validClaims(adminUserID)
	signBytes := []byte(`
-----BEGIN RSA PRIVATE KEY-----
MIICWgIBAAKBgGKGCOfH0qBzdZpCQJQlz6F455KgOieMkS0FIcQj0pd2mejAU1Rw
A07bUhLRtuNqJKdfDPG5dGsLyV31avGYzJ2QGNIKjyCzIYBxPKoT4JbRZkRpdZOb
fYGZ3Ru215o9/sLCzFik0bx+RrXEszd74tgLi8CpiZlICaOemG79SmSvAgMBAAEC
gYA8OAKPcMJbkdaqv53rLyU2Y8jfBRImhDNj2gQmd2LLcxFlgtAsBv7unv0ORaJM
Y98dcepegOUYXK7qwAtqueMuFPWl4u29GUKvV9ux0iipmCXAYnTnlKRiTp260Oom
BnaFceROpYJmfP7A3XBGzZ0RClHQvccjYHMtpA0yDldAUQJBAJ9UXnRbh7jQaNsc
COe+hT4Co+f6XOXcaC6LyDZtRGeM4MSPMAAu46VU79FE09fC5RzIhK5ukumpxOa6
/19wvOcCQQCeTRBw7QXT1q1XPARAWCmqCuyjoEFNVidTZxjhVwJ27fGB8AevuvxI
TDy7FpOpuVuwkUAwTdtnwljAOAufCxj5AkAOulPI2bUgBlPK/TptgZT7eG8CQIhZ
zxfqRY2KSmtqTwFv6fR779mnLMTGSWBzr1ZSZM6u+RWnd8P1uA9nGRq9AkB2OfQX
gs4hYmnfhvFd5PppBvOpWNysl7WTMqKAWW17yUXf15bGBg65KEcLK1dpIQh7nF+m
M9+zZJDILpNvWuhxAkBcWGe4crEQsCGREdc+Jqx8EApAfqQRhLb5vCyVNXJudglw
rqiuE4WB9nCxKD4Es+3u5dJzmWqRK/gYtKurCWVq
-----END RSA PRIVATE KEY-----
`)
	signKey, err := jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		testutil.GetLogger().With(logutil.LogAttrError(err)).Error("failed to parse PEM")
		panic(err)
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenStr, err := token.SignedString(signKey)
	if err != nil {
		testutil.GetLogger().With(logutil.LogAttrError(err)).Error("failed to sign jwt")
		panic(err)
	}
	return tokenStr
}

func invalidSecretJWT() string {
	claims := validClaims(adminUserID)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte("wrong"))
	if err != nil {
		testutil.GetLogger().With(logutil.LogAttrError(err)).Error("failed to sign jwt")
		panic(err)
	}
	return tokenStr
}

func ginContext(headers map[string]string) (*gin.Context, *httptest.ResponseRecorder, error) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	gc, _ := gin.CreateTestContext(w)
	req, err := http.NewRequest("POST", "/", nil)
	if err != nil {
		return nil, w, err
	}
	gc.Request = req
	for k, v := range headers {
		gc.Request.Header.Set(k, v)
	}
	return gc, w, nil
}

func TestReqID_NoHeader(t *testing.T) {
	gc, w, err := ginContext(map[string]string{})
	assert.Nil(t, err)

	mw := ReqID(testutil.GetLogger())
	mw(gc)
	c := gc.Request.Context()

	assert.Equal(t, 200, w.Result().StatusCode)
	actual := c.Value(ctxutil.ContextKeyReqID{})
	assert.Contains(t, actual, "generated-")
}

func TestReqID_Header(t *testing.T) {
	expected := "expected-req-id"

	gc, w, err := ginContext(map[string]string{"X-Request-Id": expected})
	assert.Nil(t, err)

	mw := ReqID(testutil.GetLogger())
	mw(gc)
	c := gc.Request.Context()

	assert.Equal(t, 200, w.Result().StatusCode)
	actual := c.Value(ctxutil.ContextKeyReqID{})
	assert.Equal(t, expected, actual)
}

func TestAuth_NoHeader(t *testing.T) {
	gc, w, err := ginContext(map[string]string{})
	assert.Nil(t, err)

	mw := Auth(testutil.GetLogger(), cfg)
	mw(gc)
	c := gc.Request.Context()

	assert.Equal(t, 401, w.Result().StatusCode)
	actual := c.Value(ctxutil.ContextKeyUserID{})
	assert.Nil(t, actual)
}

func TestAuth_NotBearer(t *testing.T) {
	gc, w, err := ginContext(map[string]string{"Authorization": basic(validAdminJWT())})
	assert.Nil(t, err)

	mw := Auth(testutil.GetLogger(), cfg)
	mw(gc)
	c := gc.Request.Context()

	assert.Equal(t, 401, w.Result().StatusCode)
	actual := c.Value(ctxutil.ContextKeyUserID{})
	assert.Nil(t, actual)
}

func TestAuth_MalformedTooManyParts(t *testing.T) {
	gc, w, err := ginContext(map[string]string{"Authorization": bearer(bearer(validAdminJWT()))})
	assert.Nil(t, err)

	mw := Auth(testutil.GetLogger(), cfg)
	mw(gc)
	c := gc.Request.Context()

	assert.Equal(t, 401, w.Result().StatusCode)
	actual := c.Value(ctxutil.ContextKeyUserID{})
	assert.Nil(t, actual)
}

func TestAuth_InvalidExpiredToken(t *testing.T) {
	gc, w, err := ginContext(map[string]string{"Authorization": bearer(invalidExpiredJWT())})
	assert.Nil(t, err)

	mw := Auth(testutil.GetLogger(), cfg)
	mw(gc)
	c := gc.Request.Context()

	assert.Equal(t, 401, w.Result().StatusCode)
	actual := c.Value(ctxutil.ContextKeyUserID{})
	assert.Nil(t, actual)
}

func TestAuth_InvalidIssuerToken(t *testing.T) {
	gc, w, err := ginContext(map[string]string{"Authorization": bearer(invalidIssuerJWT())})
	assert.Nil(t, err)

	mw := Auth(testutil.GetLogger(), cfg)
	mw(gc)
	c := gc.Request.Context()

	assert.Equal(t, 401, w.Result().StatusCode)
	actual := c.Value(ctxutil.ContextKeyUserID{})
	assert.Nil(t, actual)
}

func TestAuth_InvalidAudienceToken(t *testing.T) {
	gc, w, err := ginContext(map[string]string{"Authorization": bearer(invalidAudienceJWT())})
	assert.Nil(t, err)

	mw := Auth(testutil.GetLogger(), cfg)
	mw(gc)
	c := gc.Request.Context()

	assert.Equal(t, 401, w.Result().StatusCode)
	actual := c.Value(ctxutil.ContextKeyUserID{})
	assert.Nil(t, actual)
}

func TestAuth_InvalidSecretToken(t *testing.T) {
	gc, w, err := ginContext(map[string]string{"Authorization": bearer(invalidSecretJWT())})
	assert.Nil(t, err)

	mw := Auth(testutil.GetLogger(), cfg)
	mw(gc)
	c := gc.Request.Context()

	assert.Equal(t, 401, w.Result().StatusCode)
	actual := c.Value(ctxutil.ContextKeyUserID{})
	assert.Nil(t, actual)
}

func TestAuth_InvalidHMACAlgToken(t *testing.T) {
	gc, w, err := ginContext(map[string]string{"Authorization": bearer(invalidHMACSigningMethodJWT())})
	assert.Nil(t, err)

	mw := Auth(testutil.GetLogger(), cfg)
	mw(gc)
	c := gc.Request.Context()

	assert.Equal(t, 401, w.Result().StatusCode)
	actual := c.Value(ctxutil.ContextKeyUserID{})
	assert.Nil(t, actual)
}

func TestAuth_InvalidRSAToken(t *testing.T) {
	gc, w, err := ginContext(map[string]string{"Authorization": bearer(invalidRSAJWT())})
	assert.Nil(t, err)

	mw := Auth(testutil.GetLogger(), cfg)
	mw(gc)
	c := gc.Request.Context()

	assert.Equal(t, 401, w.Result().StatusCode)
	actual := c.Value(ctxutil.ContextKeyUserID{})
	assert.Nil(t, actual)
}

func TestAuth_ValidNonAdminToken(t *testing.T) {
	gc, w, err := ginContext(map[string]string{"Authorization": bearer(validNonAdminJWT())})
	assert.Nil(t, err)

	mw := Auth(testutil.GetLogger(), cfg)
	mw(gc)
	c := gc.Request.Context()

	assert.Equal(t, 200, w.Result().StatusCode)
	actual := c.Value(ctxutil.ContextKeyUserID{})
	assert.Equal(t, nonAdminUserID, actual)
	claims := c.Value(ctxutil.ContextKeyJWTClaims{}).(jwt.MapClaims)
	admin, _ := claims["admin"].(bool)
	assert.Equal(t, false, admin)
}

func TestAuth_ValidAdminToken(t *testing.T) {
	gc, w, err := ginContext(map[string]string{"Authorization": bearer(validAdminJWT())})
	assert.Nil(t, err)

	mw := Auth(testutil.GetLogger(), cfg)
	mw(gc)
	c := gc.Request.Context()

	assert.Equal(t, 200, w.Result().StatusCode)
	actual := c.Value(ctxutil.ContextKeyUserID{})
	assert.Equal(t, adminUserID, actual)
	claims := c.Value(ctxutil.ContextKeyJWTClaims{}).(jwt.MapClaims)
	admin, _ := claims["admin"].(bool)
	assert.Equal(t, true, admin)
}

func TestAuth_ValidExplicitAdminFalseToken(t *testing.T) {
	gc, w, err := ginContext(map[string]string{"Authorization": bearer(validNonAdminExplicitFalseJWT())})
	assert.Nil(t, err)

	mw := Auth(testutil.GetLogger(), cfg)
	mw(gc)
	c := gc.Request.Context()

	assert.Equal(t, 200, w.Result().StatusCode)
	actual := c.Value(ctxutil.ContextKeyUserID{})
	assert.Equal(t, nonAdminUserID, actual)
	claims := c.Value(ctxutil.ContextKeyJWTClaims{}).(jwt.MapClaims)
	admin, _ := claims["admin"].(bool)
	assert.Equal(t, false, admin)
}

func TestRequiresAdmin_Admin(t *testing.T) {
	gc, w, err := ginContext(map[string]string{})
	assert.Nil(t, err)

	claims := jwt.MapClaims{
		"admin": true,
	}
	gc.Request = gc.Request.WithContext(context.WithValue(gc.Request.Context(), ctxutil.ContextKeyJWTClaims{}, claims))

	mw := RequiresAdmin(testutil.GetLogger())
	mw(gc)

	assert.Equal(t, 200, w.Result().StatusCode)
}

func TestRequiresAdmin_NonAdmin(t *testing.T) {
	gc, w, err := ginContext(map[string]string{})
	assert.Nil(t, err)

	claims := jwt.MapClaims{}
	gc.Request = gc.Request.WithContext(context.WithValue(gc.Request.Context(), ctxutil.ContextKeyJWTClaims{}, claims))

	mw := RequiresAdmin(testutil.GetLogger())
	mw(gc)

	assert.Equal(t, 403, w.Result().StatusCode)
}

func TestRequiresAdmin_AdminFalse(t *testing.T) {
	gc, w, err := ginContext(map[string]string{})
	assert.Nil(t, err)

	claims := jwt.MapClaims{
		"admin": false,
	}
	gc.Request = gc.Request.WithContext(context.WithValue(gc.Request.Context(), ctxutil.ContextKeyJWTClaims{}, claims))

	mw := RequiresAdmin(testutil.GetLogger())
	mw(gc)

	assert.Equal(t, 403, w.Result().StatusCode)
}

func TestRequiresAdmin_NoClaimsAvailable(t *testing.T) {
	gc, w, err := ginContext(map[string]string{})
	assert.Nil(t, err)

	mw := RequiresAdmin(testutil.GetLogger())
	mw(gc)

	assert.Equal(t, 403, w.Result().StatusCode)
}
