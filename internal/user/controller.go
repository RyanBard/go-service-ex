package user

import (
	"context"
	"errors"
	"net/http"

	"github.com/RyanBard/go-service-ex/internal/org"
	"github.com/RyanBard/go-service-ex/pkg/user"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type UserService interface {
	GetByID(ctx context.Context, id string) (user.User, error)
	GetAll(ctx context.Context) ([]user.User, error)
	GetAllByOrgID(ctx context.Context, orgID string) ([]user.User, error)
	Save(ctx context.Context, u user.User) (user.User, error)
	Delete(ctx context.Context, u user.DeleteUser) error
}

type ctrl struct {
	log     logrus.FieldLogger
	service UserService
}

func NewController(log logrus.FieldLogger, service UserService) *ctrl {
	return &ctrl{
		log:     log.WithField("svc", "UserCTL"),
		service: service,
	}
}

func (ctr ctrl) GetByID(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	log := ctr.log.WithFields(logrus.Fields{
		"reqID":          ctx.Value("reqID"),
		"loggedInUserID": ctx.Value("userID"),
		"fn":             "GetByID",
		"id":             id,
	})
	log.Debug("called")
	u, err := ctr.service.GetByID(ctx, id)
	if err != nil {
		var statusCode int
		var notFound NotFoundErr
		if errors.As(err, &notFound) {
			log.WithError(err).Warn("resource not found")
			statusCode = http.StatusNotFound
		} else {
			log.WithError(err).Error("service call failed")
			statusCode = http.StatusInternalServerError
		}
		c.JSON(statusCode, gin.H{"message": err.Error()})
		return
	}
	log.WithField("user", u).Debug("success")
	c.JSON(http.StatusOK, u)
}

func (ctr ctrl) GetAll(c *gin.Context) {
	ctx := c.Request.Context()
	log := ctr.log.WithFields(logrus.Fields{
		"reqID":          ctx.Value("reqID"),
		"loggedInUserID": ctx.Value("userID"),
		"fn":             "GetAll",
	})
	log.Debug("called")
	u, err := ctr.service.GetAll(ctx)
	if err != nil {
		log.WithError(err).Error("service call failed")
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	log.WithField("usersLen", len(u)).Debug("success")
	c.JSON(http.StatusOK, u)
}

func (ctr ctrl) GetAllByOrgID(c *gin.Context) {
	ctx := c.Request.Context()
	orgID := c.Param("id")
	log := ctr.log.WithFields(logrus.Fields{
		"reqID":          ctx.Value("reqID"),
		"loggedInUserID": ctx.Value("userID"),
		"fn":             "GetAll",
		"id":             orgID,
	})
	log.Debug("called")
	u, err := ctr.service.GetAllByOrgID(ctx, orgID)
	if err != nil {
		var statusCode int
		var orgNotFound org.NotFoundErr
		if errors.As(err, &orgNotFound) {
			log.WithError(err).Warn("org resource not found")
			statusCode = http.StatusNotFound
		} else {
			log.WithError(err).Error("service call failed")
			statusCode = http.StatusInternalServerError
		}
		c.JSON(statusCode, gin.H{"message": err.Error()})
		return
	}
	log.WithField("usersLen", len(u)).Debug("success")
	c.JSON(http.StatusOK, u)
}

func (ctr ctrl) Save(c *gin.Context) {
	ctx := c.Request.Context()
	pathID := c.Param("id")
	log := ctr.log.WithFields(logrus.Fields{
		"reqID":          ctx.Value("reqID"),
		"loggedInUserID": ctx.Value("userID"),
		"fn":             "Save",
		"pathID":         pathID,
	})
	log.Debug("called")
	var u user.User
	if err := c.ShouldBindJSON(&u); err != nil {
		log.WithError(err).Warn("invalid request")
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	if pathID != "" {
		u.ID = pathID
	}
	log = log.WithFields(logrus.Fields{
		"user": u,
	})
	log.Debug("body processed, about to call service")
	u, err := ctr.service.Save(ctx, u)
	if err != nil {
		var statusCode int
		var notFound NotFoundErr
		var modSysUser CannotModifySysUserErr
		var assocSysOrg CannotAssociateSysOrgErr
		var orgNotFound org.NotFoundErr
		var optLock OptimisticLockErr
		var dupEmail EmailAlreadyInUseErr
		if errors.As(err, &notFound) {
			log.WithError(err).Warn("resource not found")
			statusCode = http.StatusNotFound
		} else if errors.As(err, &modSysUser) {
			log.WithError(err).Warn("cannot modify system user")
			statusCode = http.StatusForbidden
		} else if errors.As(err, &optLock) {
			log.WithError(err).Warn("optimistic lock error")
			statusCode = http.StatusConflict
		} else if errors.As(err, &dupEmail) {
			log.WithError(err).Warn("duplicate email error")
			statusCode = http.StatusConflict
		} else if errors.As(err, &orgNotFound) {
			log.WithError(err).Warn("org resource not found")
			statusCode = http.StatusBadRequest
		} else if errors.As(err, &assocSysOrg) {
			// TODO - you could choose to 404 this to obfuscate for security reasons, but I'm letting error details go through in the response atm, so probably not worth it right now
			log.WithError(err).Warn("cannot associate system org")
			statusCode = http.StatusForbidden
		} else {
			log.WithError(err).Error("service call failed")
			statusCode = http.StatusInternalServerError
		}
		c.JSON(statusCode, gin.H{"message": err.Error()})
		return
	}
	log.Debug("success")
	c.JSON(http.StatusOK, u)
}

func (ctr ctrl) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	pathID := c.Param("id")
	log := ctr.log.WithFields(logrus.Fields{
		"reqID":          ctx.Value("reqID"),
		"loggedInUserID": ctx.Value("userID"),
		"fn":             "Delete",
		"pathID":         pathID,
	})
	log.Debug("called")
	var u user.DeleteUser
	if err := c.ShouldBindJSON(&u); err != nil {
		log.WithError(err).Warn("invalid request")
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	if pathID != "" {
		u.ID = pathID
	}
	log = log.WithFields(logrus.Fields{
		"user": u,
	})
	log.Debug("body processed, about to call service")
	if err := ctr.service.Delete(ctx, u); err != nil {
		var statusCode int
		var notFound NotFoundErr
		var modSysUser CannotModifySysUserErr
		var optLock OptimisticLockErr
		if errors.As(err, &notFound) {
			log.WithError(err).Warn("resource already gone, not deleting")
			c.Status(http.StatusNoContent)
			return
		} else if errors.As(err, &modSysUser) {
			log.WithError(err).Warn("cannot modify system user")
			statusCode = http.StatusForbidden
		} else if errors.As(err, &optLock) {
			log.WithError(err).Warn("optimistic lock error")
			statusCode = http.StatusConflict
		} else {
			log.WithError(err).Error("service call failed")
			statusCode = http.StatusInternalServerError
		}
		c.JSON(statusCode, gin.H{"message": err.Error()})
		return
	}
	log.Debug("success")
	c.Status(http.StatusNoContent)
}
