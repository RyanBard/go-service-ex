package mdlw

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	ctxutil "github.com/RyanBard/go-ctx-util/pkg"
	logutil "github.com/RyanBard/go-log-util/pkg"
	"github.com/RyanBard/go-service-ex/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

func logAttrSVC() slog.Attr {
	return logutil.LogAttrSVC("Middleware")
}

func ReqID(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		reqID := c.GetHeader("x-request-id")
		if reqID == "" {
			reqID = "generated-" + uuid.New().String()
		}
		ctx := context.WithValue(c.Request.Context(), ctxutil.ContextKeyReqID{}, reqID)
		c.Request = c.Request.WithContext(ctx)
		log := logger.With(
			logAttrSVC(),
			logutil.LogAttrReqID(ctx),
			logutil.LogAttrFN("ReqID"),
		)
		log.Debug("called")
	}
}

func Auth(logger *slog.Logger, cfg config.AuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := logger.With(
			logAttrSVC(),
			logutil.LogAttrReqID(c.Request.Context()),
			logutil.LogAttrFN("Auth"),
		)
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
			log.With(slog.Int("partsLen", partsLen)).Warn("Authorization header was malformed (not well formed Bearer)")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		token := strings.Trim(parts[1], " \t\r\n")
		claims, err := validateJWT(cfg, token)
		if err != nil {
			log.With(logutil.LogAttrError(err)).Warn("jwt validation failed")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		userID := claims["sub"]
		ctx := context.WithValue(c.Request.Context(), ctxutil.ContextKeyUserID{}, userID)
		ctx = context.WithValue(ctx, ctxutil.ContextKeyJWTClaims{}, claims)
		log.With(
			slog.Any("claims", claims),
			logutil.LogAttrLoggedInUserID(ctx),
		).Debug("Token was valid, proceeding")
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

func RequiresAdmin(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		log := logger.With(
			logAttrSVC(),
			logutil.LogAttrReqID(ctx),
			logutil.LogAttrLoggedInUserID(ctx),
			logutil.LogAttrFN("RequiresAdmin"),
		)
		log.Debug("called")
		claims, ok := ctx.Value(ctxutil.ContextKeyJWTClaims{}).(jwt.MapClaims)
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
