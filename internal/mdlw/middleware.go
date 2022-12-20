package mdlw

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"net/http"
	"strings"
)

func ReqID(logger logrus.FieldLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		reqID := c.GetHeader("x-request-id")
		if reqID == "" {
			reqID = "generated-" + uuid.New().String()
		}
		log := logger.WithField("reqID", reqID)
		log.Debug("ReqID called")
		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), "reqID", reqID))
	}
}

func Auth(logger logrus.FieldLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := logger.WithField("reqID", c.Request.Context().Value("reqID"))
		log.Debug("Auth called")
		auth := c.GetHeader("authorization")
		if auth == "" {
			log.Debug("Authorization header was missing or empty")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		parts := strings.Split(auth, "Bearer ")
		partsLen := len(parts)
		if partsLen != 2 {
			log.WithField("partsLen", partsLen).Debug("Authorization header was malformed (not well formed Bearer)")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		token := strings.Trim(parts[1], " \t\r\n")
		// TODO - jwt validate token instead of hardcoding
		if token != "foo" {
			log.Debug("Token was invalid")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		log.Debug("Token was valid, proceeding")
		userID := "TODO"
		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), "userID", userID))
	}
}
