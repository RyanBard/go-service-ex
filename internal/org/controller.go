package org

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/RyanBard/gin-ex/pkg/org"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"io"
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
	id := c.Param("id")
	log := ctr.log.WithFields(logrus.Fields{
		"reqID": c.Request.Context().Value("reqID"),
		"fn":    "GetByID",
		"id":    id,
	})
	log.Debug("called")
	o, err := ctr.service.GetByID(c.Request.Context(), id)
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
	name := c.Query("name")
	log := ctr.log.WithFields(logrus.Fields{
		"reqID": c.Request.Context().Value("reqID"),
		"fn":    "GetAll",
		"name":  name,
	})
	log.Debug("called")
	o, err := ctr.service.GetAll(c.Request.Context(), name)
	if err != nil {
		log.WithError(err).Error("service call failed")
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	log.WithField("orgsLen", len(o)).Debug("success")
	c.JSON(http.StatusOK, o)
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
		log.WithError(err).Error("read of req body failed")
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	defer c.Request.Body.Close()
	var o org.Org
	if err = json.Unmarshal(bytes, &o); err != nil {
		log.WithError(err).Warn("unmarshalling failed")
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	if pathID != "" {
		o.ID = pathID
	}
	if err = ctr.validate.Struct(o); err != nil {
		log.WithError(err).Warn("invalid org body")
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	log = log.WithFields(logrus.Fields{
		"org": o,
	})
	log.Debug("body processed, about to call service")
	o, err = ctr.service.Save(c.Request.Context(), o)
	if err != nil {
		var statusCode int
		var notFound NotFoundErr
		var modSysOrg CannotModifySysOrgErr
		var optLock OptimisticLockErr
		if errors.As(err, &notFound) {
			log.WithError(err).Warn("resource not found")
			statusCode = http.StatusNotFound
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
	log.Debug("success")
	c.JSON(http.StatusOK, o)
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
	var o org.DeleteOrg
	if err = json.Unmarshal(bytes, &o); err != nil {
		log.WithError(err).Warn("unmarshalling failed")
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	if pathID != "" {
		o.ID = pathID
	}
	if err = ctr.validate.Struct(o); err != nil {
		log.WithError(err).Warn("invalid org body")
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	log = log.WithFields(logrus.Fields{
		"org": o,
	})
	log.Debug("body processed, about to call service")
	if err := ctr.service.Delete(c.Request.Context(), o); err != nil {
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
