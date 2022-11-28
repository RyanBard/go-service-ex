package main

import (
	"fmt"
	"github.com/RyanBard/gin-ex/internal/pkg/idgen"
	"github.com/RyanBard/gin-ex/internal/pkg/mdlw"
	"github.com/RyanBard/gin-ex/internal/pkg/org"
	"github.com/RyanBard/gin-ex/internal/pkg/timer"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"net/http"
)

func main() {
	port := 4000
	log := logrus.StandardLogger()
	log.SetLevel(logrus.DebugLevel)
	validate := validator.New()
	orgDAO := org.NewOrgDAO(log)
	timer := timer.New()
	idGenerator := idgen.New()
	orgService := org.NewOrgService(log, orgDAO, timer, idGenerator)
	orgCtrl := org.NewOrgController(log, validate, orgService)
	log.WithField("ctrl", orgCtrl).Info("Here")

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(mdlw.ReqID(log))

	r.GET("/health", func(c *gin.Context) {
		log.WithField("reqID", c.Request.Context().Value("reqID")).Debug("Health called")
		c.Status(http.StatusNoContent)
	})

	r.GET("/readiness", func(c *gin.Context) {
		log.WithField("reqID", c.Request.Context().Value("reqID")).Debug("Readiness called")
		c.Status(http.StatusNoContent)
	})

	authorized := r.Group("/api")
	authorized.Use(mdlw.Auth(log))

	authorized.GET("/orgs/:id", orgCtrl.GetByID)
	authorized.GET("/orgs", orgCtrl.GetAll)
	authorized.POST("/orgs", orgCtrl.Save)
	authorized.PUT("/orgs", orgCtrl.Save)
	authorized.POST("/orgs/:id", orgCtrl.Save)
	authorized.PUT("/orgs/:id", orgCtrl.Save)
	authorized.DELETE("/orgs/:id", orgCtrl.Delete)

	r.Run(fmt.Sprintf(":%v", port))
}
