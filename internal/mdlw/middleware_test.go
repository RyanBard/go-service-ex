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
	// ssh-keygen -t rsa -b 4096 -f foo_rsa -m PEM < /dev/null
	signBytes := []byte(`
-----BEGIN RSA PRIVATE KEY-----
MIIJJwIBAAKCAgEA4aIC+LRLdENP+Iz03iLzAtBh7jAPYb9Tk5IbBGe/skEREzPx
p6b4Dxv9V8bjkEHWPo4wLoooVgXSYcW+LwC9HLZKyoux+bYIdmYVEq2FvGlXh+Ev
XG42XokGNdQTi5i/t9Ylpx7+bwdJAhtfnCzj6NZnPX0O5nizQWKomaw21l5rEQYz
c75y16jGY1DXrWFjoPvgceo9oT0BDJlmlfItL7HoUtYmtt4p0Ic2vrr5WrIgpNyK
mvNvkQggSIne1idpijnqm8IGzGh9oWotTwrhicFji7TZbJPMAU9ILTumYXqSS+VC
Zh9Y70KerDRMr5EQl7w/IqtbXb42AAGtbW9mpq0nL4labmY/HcG7SSjXwxjNmBeO
R3oxpBaH5SfV/io4MRkY2EKHeTbTt8djRrWHL3bTkQLzfuY0vacmMjWPNM4NwX24
V+ymvSoSSbi77QGE3DjmlsDDtBYPICp6dOT8AYyCNiKTbN4LLWHXBlwYfWwGwk7i
OXOeEWQKtd3FSMB3EhJvEHyw0lh1gbMphEAQrUOEiTnj+29zc3fdKEW0cNvem9+w
OGmtIcAGRyKLKXYTkqH8YSgoTGlA8ZUSHAJBx2xBjNK9MXtaPDa2d/js6imVs2WA
yGoYY+m1QvZLBIrOBbS7VcUANvyId4Tozu6OKSbmhFbXlgzMHB37fqWFknECAwEA
AQKCAgAj/1TBiIQ0PYv778MeUI2cMJI7LcQ6eEuVct0oGARLpFlcyq/+ayNOx8yk
rSi/ql1LuCQkPKsARGgC7MkxlwjU1Jl8lIp1uC7D8vfgNbptK5JJLapAcR3v+aT5
wAbJQfquUfZKYiW+IXeqpCLeGARZcKFifDn9F5sjrqGO+Nx72W7z9Z0OVX9dfEIA
dNSgBvZd5+lwSyp/d6vb7nphKFTAs/3If32INcwOho/7oHlpWJtgKgZb/8QYK6d4
4zQfzwRlxuUw6eTrRi1q6/VkE6yVbXrb/mOk0LSerd8qRka+bQ97l80+3TfzXhnS
B/sHFJoE10XsWlHWlCLp39qjJ7+EwcP+ks8CB5mvpROQ4zJVrDj/VG+hdwPfExg6
adu0NmL57KOTxwZND9I0tz6yhT+/77IlPbkO3bLm7KSnY+J8ASHqkoLkAlFLIJuL
VtCam7GtLJUW4B/2R1zO7mhXZm2S/rGUFV7aXG6AIAVG9iIWZGOBkDySip9kANmp
lTmcaf28+gJcuKDMEcyCBl5sihXX276LIWQEdpNbtugfz79YRyILbkgvPzf+lTei
8+x81QzAsld8vXFojOiila0Zq98HRF5qYHk78LT+VKaB0pLfHvCZ6bCpagz5Vs3V
NXOM7L4c9HY2hAtq55GbRKhy5qHjj9r85Mw2f4sH2M1jMYlapQKCAQEA+X3AfyZ3
QG6wKeFhz0KiLrsP/SgpTDDpy2eGY/VWC61DrxTcAODbwbb0ogHlH3m2n0crCeND
LLihdPIVnZYQ2wnsuezYK4axXfSWIhU+RHomGkcAId+pJbsrqbqXNhZBQgBAH5+A
AY6L3yIRH4kXAEEsVvJ+XitpGZjbJgfYWR+HcRrnYrA4V3dNcJQuFQ7zFL3Qf0O3
z/OXPUFQVW94TGgzr7/lReOVuivvwfVtsnYN17prVG7p1MVgI3lx+LtbfTlcJ8AU
I/wPqyOnIdz5n7/34H7FQKBJgMt1ALHab+nn6xX0ZRO/LJb1ZQ4FlV0asIsZ10VF
L/AvvV6tluULdQKCAQEA54Trav/ANwPekAqtT1rfLEw9TkNtIZ++bQ3V4my9sAdn
O2wWfRvP+KvaULm+evK8qIMs7AY/alVRcR6oGnI0d5Lyt4E51biezXCpRAfhXtaZ
o2NmuXXGUGU703f0YrjJtGsJe9nWfPDeMakgXhiaOVAAfjdjmOFgr+pJj2dcl/By
HAEcYE1ON43R9ZT9bM37P0ZcR01YmPkBvP8x3aU5PFyRi5Lt6AiyR78ffzQp/aQA
P78UDF1UvRwMsHLe3MgUHJSLp1gQI8A1IrfRbo2X8nsT1TQtg6vjyoiQwqvLNzlz
cyCRFPR1C8ByMngxjpi/iCAPdAS/WnZEleimdOPXjQKCAQAoBoalX3RuP9O/REs+
xv/h41zTTmghsw8u9LLYwnlFckyTgBMziN94mnNwskEh5ZLoxrAe6/jQ6eXtLxFM
sNCPc24o/dhIU1mNKzoSybmtnrMSMCXTSWCHjxmYJzkvEi7x+bxP1nTgyw+hgoOg
vYuVHN+SK2EelnRMCPvPhB6vXaGLAbfxWzgOwDMKRpuGAVn6D+GtKR4KHJg82Yeb
zUBNccIpBe4wYiyZK4dRZKBuqwXZgCzL5OdanTepOUiRkO1Cj7s7HVOd2khhOl0L
z7m0aUXJHE1k6tIf+YQR0naq0anFS0ZkGODotGc2CSPYLJGllQtWP+SzpiLijJfo
x2IpAoIBAGLcdm6VUS3s4A/gD0uQgTX4REF0N1ihTv5gZxGt334YWzObnyrDQpZs
Q+guuLDlkDvg78DFljTAw+sq7RuF4vOgczZ63whtMpqFXjInr0wBFVAYtRfCu/E+
mUJeuaOiRrdc8YrKZXWqoSbRknm7xbdtexhqbIQu0UutHsY2m+phiUh+DPDucdT1
IU/d1C2u83d8gxELbSa9Rcm/qzp7QCqPWLMiBacMd5x9d2SpELk3W8fcvyHchL/Q
B3cdRh8+7J+bqtrFlva9L7NojOzJV7X6amijREAymS22DSKjmz40sSDbKjipfoWL
Ivh8LKU4gqkND6w6DsABOp5M3y+Qz00CggEASVV/QY4LnXA1HdrlHPnDiz2sR6Iq
bXwyQIKU3IgYXM0F9ot+bBLd0VTzjcsRh3w99l6JihmoTrBFi+BFDArDbFT8p4A+
Su7V8CHJ1EEpvgnqsLqj84ApDPJo8o6Aceb/R1MikkLDu/2yk7AFbjgjGtd6tUNS
iy7VnpOkPBaViFX5T5jnaU5/ucMyjxLZ+/VMiqZyAZXXWPyjIdAcL9fAGcYWg2jX
dltVblopfW8NKuW272RU1lxdV0xIEnyBlt80cQxLa4zIocxZ2dQOqGB0l4uQS8Yf
pejaRlDi6F2/ZfQs1P45gBz/c4DNXL/DiVJu6bvI5PJ4r40K+cZwd1vVSQ==
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
