package org

import (
	"context"
	"errors"
	"github.com/RyanBard/go-service-ex/pkg/org"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"net/http"
)

type OrgService interface {
	GetByID(ctx context.Context, id string) (org.Org, error)
	GetAll(ctx context.Context, name string) ([]org.Org, error)
	Save(ctx context.Context, o org.Org) (org.Org, error)
	Delete(ctx context.Context, o org.DeleteOrg) error
}

type ctrl struct {
	log      logrus.FieldLogger
	validate *validator.Validate
	service  OrgService
}

func NewController(log logrus.FieldLogger, validate *validator.Validate, service OrgService) *ctrl {
	return &ctrl{
		log:      log.WithField("SVC", "OrgCTL"),
		validate: validate,
		service:  service,
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
	o, err := ctr.service.GetByID(ctx, id)
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
	log.WithField("org", o).Debug("success")
	c.JSON(http.StatusOK, o)
}

func (ctr ctrl) GetAll(c *gin.Context) {
	ctx := c.Request.Context()
	name := c.Query("name")
	log := ctr.log.WithFields(logrus.Fields{
		"reqID":          ctx.Value("reqID"),
		"loggedInUserID": ctx.Value("userID"),
		"fn":             "GetAll",
		"name":           name,
	})
	log.Debug("called")
	o, err := ctr.service.GetAll(ctx, name)
	if err != nil {
		log.WithError(err).Error("service call failed")
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	log.WithField("orgsLen", len(o)).Debug("success")
	c.JSON(http.StatusOK, o)
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
	var o org.Org
	if err := c.ShouldBindJSON(&o); err != nil {
		log.WithError(err).Warn("unmarshalling failed")
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	if pathID != "" {
		o.ID = pathID
	}
	if err := ctr.validate.Struct(o); err != nil {
		log.WithError(err).Warn("invalid org body")
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	log = log.WithFields(logrus.Fields{
		"org": o,
	})
	log.Debug("body processed, about to call service")
	o, err := ctr.service.Save(ctx, o)
	if err != nil {
		var statusCode int
		var notFound NotFoundErr
		var modSysOrg CannotModifySysOrgErr
		var optLock OptimisticLockErr
		var dupName NameAlreadyInUseErr
		if errors.As(err, &notFound) {
			log.WithError(err).Warn("resource not found")
			statusCode = http.StatusNotFound
		} else if errors.As(err, &modSysOrg) {
			log.WithError(err).Warn("cannot modify system org")
			statusCode = http.StatusForbidden
		} else if errors.As(err, &optLock) {
			log.WithError(err).Warn("optimistic lock error")
			statusCode = http.StatusConflict
		} else if errors.As(err, &dupName) {
			log.WithError(err).Warn("duplicate name error")
			statusCode = http.StatusConflict
		} else {
			log.WithError(err).Error("service call failed")
			statusCode = http.StatusInternalServerError
		}
		c.JSON(statusCode, gin.H{"message": err.Error()})
		return
	}
	log.Debug("success")
	c.JSON(http.StatusOK, o)
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
	var o org.DeleteOrg
	if err := c.ShouldBindJSON(&o); err != nil {
		log.WithError(err).Warn("unmarshalling failed")
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	if pathID != "" {
		o.ID = pathID
	}
	if err := ctr.validate.Struct(o); err != nil {
		log.WithError(err).Warn("invalid org body")
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	log = log.WithFields(logrus.Fields{
		"org": o,
	})
	log.Debug("body processed, about to call service")
	if err := ctr.service.Delete(ctx, o); err != nil {
		var statusCode int
		var notFound NotFoundErr
		var modSysOrg CannotModifySysOrgErr
		var optLock OptimisticLockErr
		if errors.As(err, &notFound) {
			log.WithError(err).Warn("resource already gone, not deleting")
			c.Status(http.StatusNoContent)
			return
		} else if errors.As(err, &modSysOrg) {
			log.WithError(err).Warn("cannot modify system org")
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
	log.Debug("Success")
	c.Status(http.StatusNoContent)
}
