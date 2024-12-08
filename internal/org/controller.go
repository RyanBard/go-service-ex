package org

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/RyanBard/go-service-ex/internal/logutil"
	"github.com/RyanBard/go-service-ex/pkg/org"
	"github.com/gin-gonic/gin"
)

type OrgService interface {
	GetByID(ctx context.Context, id string) (org.Org, error)
	GetAll(ctx context.Context, name string) ([]org.Org, error)
	Save(ctx context.Context, o org.Org) (org.Org, error)
	Delete(ctx context.Context, o org.DeleteOrg) error
}

type ctrl struct {
	log     *slog.Logger
	service OrgService
}

func NewController(log *slog.Logger, service OrgService) *ctrl {
	return &ctrl{
		log:     log.With(logutil.LogAttrSVC("OrgCTL")),
		service: service,
	}
}

func (ctr ctrl) GetByID(c *gin.Context) {
	ctx := c.Request.Context()
	id := c.Param("id")
	log := ctr.log.With(
		logutil.LogAttrReqID(ctx),
		logutil.LogAttrLoggedInUserID(ctx),
		logutil.LogAttrFN("GetByID"),
		logAttrOrgID(id),
	)
	log.Debug("called")
	o, err := ctr.service.GetByID(ctx, id)
	if err != nil {
		var statusCode int
		var notFound NotFoundErr
		if errors.As(err, &notFound) {
			log.With(logutil.LogAttrError(err)).Warn("resource not found")
			statusCode = http.StatusNotFound
		} else {
			log.With(logutil.LogAttrError(err)).Error("service call failed")
			statusCode = http.StatusInternalServerError
		}
		c.JSON(statusCode, gin.H{"message": err.Error()})
		return
	}
	log.With(logAttrOrg(o)).Debug("success")
	c.JSON(http.StatusOK, o)
}

func (ctr ctrl) GetAll(c *gin.Context) {
	ctx := c.Request.Context()
	name := c.Query("name")
	log := ctr.log.With(
		logutil.LogAttrReqID(ctx),
		logutil.LogAttrLoggedInUserID(ctx),
		logutil.LogAttrFN("GetAll"),
		logAttrOrgName(name),
	)
	log.Debug("called")
	o, err := ctr.service.GetAll(ctx, name)
	if err != nil {
		log.With(logutil.LogAttrError(err)).Error("service call failed")
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	log.With(logAttrOrgsLen(len(o))).Debug("success")
	c.JSON(http.StatusOK, o)
}

func (ctr ctrl) Save(c *gin.Context) {
	ctx := c.Request.Context()
	pathID := c.Param("id")
	log := ctr.log.With(
		logutil.LogAttrReqID(ctx),
		logutil.LogAttrLoggedInUserID(ctx),
		logutil.LogAttrFN("Save"),
		logAttrPathID(pathID),
	)
	log.Debug("called")
	var o org.Org
	if err := c.ShouldBindJSON(&o); err != nil {
		log.With(logutil.LogAttrError(err)).Warn("invalid request")
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	if pathID != "" {
		o.ID = pathID
	}
	log = log.With(logAttrOrg(o))
	log.Debug("body processed, about to call service")
	o, err := ctr.service.Save(ctx, o)
	if err != nil {
		var statusCode int
		var notFound NotFoundErr
		var modSysOrg CannotModifySysOrgErr
		var optLock OptimisticLockErr
		var dupName NameAlreadyInUseErr
		if errors.As(err, &notFound) {
			log.With(logutil.LogAttrError(err)).Warn("resource not found")
			statusCode = http.StatusNotFound
		} else if errors.As(err, &modSysOrg) {
			log.With(logutil.LogAttrError(err)).Warn("cannot modify system org")
			statusCode = http.StatusForbidden
		} else if errors.As(err, &optLock) {
			log.With(logutil.LogAttrError(err)).Warn("optimistic lock error")
			statusCode = http.StatusConflict
		} else if errors.As(err, &dupName) {
			log.With(logutil.LogAttrError(err)).Warn("duplicate name error")
			statusCode = http.StatusConflict
		} else {
			log.With(logutil.LogAttrError(err)).Error("service call failed")
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
	log := ctr.log.With(
		logutil.LogAttrReqID(ctx),
		logutil.LogAttrLoggedInUserID(ctx),
		logutil.LogAttrFN("Delete"),
		logAttrPathID(pathID),
	)
	log.Debug("called")
	var o org.DeleteOrg
	if err := c.ShouldBindJSON(&o); err != nil {
		log.With(logutil.LogAttrError(err)).Warn("invalid request")
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	if pathID != "" {
		o.ID = pathID
	}
	log = log.With(logAttrOrg(o))
	log.Debug("body processed, about to call service")
	if err := ctr.service.Delete(ctx, o); err != nil {
		var statusCode int
		var notFound NotFoundErr
		var modSysOrg CannotModifySysOrgErr
		var optLock OptimisticLockErr
		if errors.As(err, &notFound) {
			log.With(logutil.LogAttrError(err)).Warn("resource already gone, not deleting")
			c.Status(http.StatusNoContent)
			return
		} else if errors.As(err, &modSysOrg) {
			log.With(logutil.LogAttrError(err)).Warn("cannot modify system org")
			statusCode = http.StatusForbidden
		} else if errors.As(err, &optLock) {
			log.With(logutil.LogAttrError(err)).Warn("optimistic lock error")
			statusCode = http.StatusConflict
		} else {
			log.With(logutil.LogAttrError(err)).Error("service call failed")
			statusCode = http.StatusInternalServerError
		}
		c.JSON(statusCode, gin.H{"message": err.Error()})
		return
	}
	log.Debug("Success")
	c.Status(http.StatusNoContent)
}
