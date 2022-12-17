package org

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
)

type OrgService interface {
	GetByID(ctx context.Context, id string) (Org, error)
	GetAll(ctx context.Context, name string) ([]Org, error)
	Save(ctx context.Context, o Org) (Org, error)
	Delete(ctx context.Context, o DeleteOrg) error
}

type ctrl struct {
	log      logrus.FieldLogger
	validate *validator.Validate
	service  OrgService
}

func NewOrgController(log logrus.FieldLogger, validate *validator.Validate, service OrgService) *ctrl {
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
		log.WithError(err).Error("Service call failed")
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	log.WithField("org", o).Debug("Success")
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
		log.WithError(err).Error("Service call failed")
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	log.WithField("orgsLen", len(o)).Debug("Success")
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
		log.WithError(err).Error("Read of req body failed")
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	defer c.Request.Body.Close()
	var o Org
	if err = json.Unmarshal(bytes, &o); err != nil {
		log.WithError(err).Error("Unmarshalling failed")
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	if pathID != "" {
		o.ID = pathID
	}
	if err = ctr.validate.Struct(o); err != nil {
		log.WithError(err).Error("Invalid org body")
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	log = ctr.log.WithFields(logrus.Fields{
		"org": o,
	})
	log.Debug("body processed, about to call service")
	o, err = ctr.service.Save(c.Request.Context(), o)
	if err != nil {
		log.WithError(err).Error("Service call failed")
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	log.Debug("Success")
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
	var o DeleteOrg
	if err = json.Unmarshal(bytes, &o); err != nil {
		log.WithError(err).Error("Unmarshalling failed")
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	if pathID != "" {
		o.ID = pathID
	}
	if err = ctr.validate.Struct(o); err != nil {
		log.WithError(err).Error("Invalid org body")
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	log = ctr.log.WithFields(logrus.Fields{
		"org": o,
	})
	log.Debug("body processed, about to call service")
	if err := ctr.service.Delete(c.Request.Context(), o); err != nil {
		log.WithError(err).Error("Service call failed")
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	log.Debug("Success")
	c.Status(http.StatusNoContent)
}
