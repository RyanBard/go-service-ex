package user

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
)

type UserService interface {
	GetByID(ctx context.Context, id string) (User, error)
	GetAll(ctx context.Context) ([]User, error)
	GetAllByOrgID(ctx context.Context, orgID string) ([]User, error)
	Save(ctx context.Context, u User) (User, error)
	Delete(ctx context.Context, u DeleteUser) error
}

type ctrl struct {
	log      logrus.FieldLogger
	validate *validator.Validate
	service  UserService
}

func NewController(log logrus.FieldLogger, validate *validator.Validate, service UserService) *ctrl {
	return &ctrl{
		log:      log.WithField("SVC", "UserCTL"),
		validate: validate,
		service:  service,
	}
}

func (ctr ctrl) GetByID(c *gin.Context) {
	id := c.Param("id")
	log := ctr.log.WithFields(logrus.Fields{
		"reqID": c.Request.Context().Value("reqID"),
		"fn":    "GetByID",
		"id":    id,
	})
	log.Debug("called")
	u, err := ctr.service.GetByID(c.Request.Context(), id)
	if err != nil {
		log.WithError(err).Error("Service call failed")
		var statusCode int
		var notFound NotFoundErr
		if errors.As(err, &notFound) {
			statusCode = http.StatusNotFound
		} else {
			statusCode = http.StatusInternalServerError
		}
		c.JSON(statusCode, gin.H{"message": err.Error()})
		return
	}
	log.WithField("user", u).Debug("Success")
	c.JSON(http.StatusOK, u)
}

func (ctr ctrl) GetAll(c *gin.Context) {
	log := ctr.log.WithFields(logrus.Fields{
		"reqID": c.Request.Context().Value("reqID"),
		"fn":    "GetAll",
	})
	log.Debug("called")
	u, err := ctr.service.GetAll(c.Request.Context())
	if err != nil {
		log.WithError(err).Error("Service call failed")
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	log.WithField("usersLen", len(u)).Debug("Success")
	c.JSON(http.StatusOK, u)
}

func (ctr ctrl) GetAllByOrgID(c *gin.Context) {
	orgID := c.Param("id")
	log := ctr.log.WithFields(logrus.Fields{
		"reqID": c.Request.Context().Value("reqID"),
		"fn":    "GetAll",
		"id":    orgID,
	})
	log.Debug("called")
	u, err := ctr.service.GetAllByOrgID(c.Request.Context(), orgID)
	if err != nil {
		log.WithError(err).Error("Service call failed")
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	log.WithField("usersLen", len(u)).Debug("Success")
	c.JSON(http.StatusOK, u)
}

func (ctr ctrl) Save(c *gin.Context) {
	pathID := c.Param("id")
	log := ctr.log.WithFields(logrus.Fields{
		"reqID":  c.Request.Context().Value("reqID"),
		"fn":     "Save",
		"pathID": pathID,
	})
	log.Debug("called")
	bytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.WithError(err).Error("Read of req body failed")
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	defer c.Request.Body.Close()
	var u User
	if err = json.Unmarshal(bytes, &u); err != nil {
		log.WithError(err).Error("Unmarshalling failed")
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	if pathID != "" {
		u.ID = pathID
	}
	if err = ctr.validate.Struct(u); err != nil {
		log.WithError(err).Error("Invalid user body")
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	log = ctr.log.WithFields(logrus.Fields{
		"user": u,
	})
	log.Debug("body processed, about to call service")
	u, err = ctr.service.Save(c.Request.Context(), u)
	if err != nil {
		log.WithError(err).Error("Service call failed")
		var statusCode int
		var notFound NotFoundErr
		var modSysUser CannotModifySysUserErr
		var assocSysOrg CannotAssociateSysOrgErr
		if errors.As(err, &notFound) {
			statusCode = http.StatusNotFound
		} else if errors.As(err, &modSysUser) {
			statusCode = http.StatusForbidden
		} else if errors.As(err, &assocSysOrg) {
			// TODO - you could choose to 404 this to obfuscate for security reasons, but I'm letting error details go through in the response atm, so probably not worth it right now
			statusCode = http.StatusBadRequest
		} else {
			statusCode = http.StatusInternalServerError
		}
		c.JSON(statusCode, gin.H{"message": err.Error()})
		return
	}
	log.Debug("Success")
	c.JSON(http.StatusOK, u)
}

func (ctr ctrl) Delete(c *gin.Context) {
	pathID := c.Param("id")
	log := ctr.log.WithFields(logrus.Fields{
		"reqID":  c.Request.Context().Value("reqID"),
		"fn":     "Delete",
		"pathID": pathID,
	})
	log.Debug("called")
	bytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		log.WithError(err).Error("Read of req body failed")
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	defer c.Request.Body.Close()
	var u DeleteUser
	if err = json.Unmarshal(bytes, &u); err != nil {
		log.WithError(err).Error("Unmarshalling failed")
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	if pathID != "" {
		u.ID = pathID
	}
	if err = ctr.validate.Struct(u); err != nil {
		log.WithError(err).Error("Invalid user body")
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	log = ctr.log.WithFields(logrus.Fields{
		"user": u,
	})
	log.Debug("body processed, about to call service")
	if err := ctr.service.Delete(c.Request.Context(), u); err != nil {
		log.WithError(err).Error("Service call failed")
		var statusCode int
		var notFound NotFoundErr
		var modSysUser CannotModifySysUserErr
		if errors.As(err, &notFound) {
			statusCode = http.StatusNotFound
		} else if errors.As(err, &modSysUser) {
			statusCode = http.StatusForbidden
		} else {
			statusCode = http.StatusInternalServerError
		}
		c.JSON(statusCode, gin.H{"message": err.Error()})
		return
	}
	log.Debug("Success")
	c.Status(http.StatusNoContent)
}
