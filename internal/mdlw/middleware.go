package mdlw

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/RyanBard/go-service-ex/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

func ReqID(logger logrus.FieldLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		reqID := c.GetHeader("x-request-id")
		if reqID == "" {
			reqID = "generated-" + uuid.New().String()
		}
		log := logger.WithFields(logrus.Fields{
			"svc":   "Middleware",
			"reqID": reqID,
			"fn":    "ReqID",
		})
		log.Debug("called")
		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), "reqID", reqID))
	}
}

func Auth(logger logrus.FieldLogger, cfg config.AuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := logger.WithFields(logrus.Fields{
			"svc":   "Middleware",
			"reqID": c.Request.Context().Value("reqID"),
			"fn":    "Auth",
		})
		log.Debug("called")
		auth := c.GetHeader("authorization")
		if auth == "" {
			log.Warn("Authorization header was missing or empty")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		parts := strings.Split(auth, "Bearer ")
		partsLen := len(parts)
		if partsLen != 2 {
			log.WithField("partsLen", partsLen).Warn("Authorization header was malformed (not well formed Bearer)")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		token := strings.Trim(parts[1], " \t\r\n")
		claims, err := validateJWT(cfg, token)
		if err != nil {
			log.WithError(err).Warn("jwt validation failed")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		userID := claims["sub"]
		log.WithFields(logrus.Fields{
			"claims":         claims,
			"loggedInUserID": userID,
		}).Debug("Token was valid, proceeding")
		ctx := context.WithValue(c.Request.Context(), "userID", userID)
		ctx = context.WithValue(ctx, "jwtClaims", claims)
		c.Request = c.Request.WithContext(ctx)
	}
}

func validateJWT(cfg config.AuthConfig, tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		method, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, fmt.Errorf("unexpected signing algorithm: %v", token.Header["alg"])
		}
		if method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", method)
		}
		return []byte(cfg.JWTSecret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("could not cast to jwt.MapClaims")
	}
	if !claims.VerifyAudience(cfg.JWTAudience, true) {
		return nil, errors.New("audience was invalid")
	}
	if !claims.VerifyIssuer(cfg.JWTIssuer, true) {
		return nil, errors.New("issuer was invalid")
	}
	return claims, nil
}

func RequiresAdmin(logger logrus.FieldLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		log := logger.WithFields(logrus.Fields{
			"svc":            "Middleware",
			"reqID":          ctx.Value("reqID"),
			"loggedInUserID": ctx.Value("userID"),
			"fn":             "RequiresAdmin",
		})
		log.Debug("called")
		claims, ok := ctx.Value("jwtClaims").(jwt.MapClaims)
		if !ok {
			log.Warn("jwtClaims not in context")
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		isAdmin, ok := claims["admin"].(bool)
		if !ok {
			log.Warn("admin claim not in jwtClaims")
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		if !isAdmin {
			log.Warn("admin claim is false")
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
	}
}
