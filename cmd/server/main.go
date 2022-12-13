package main

import (
	"fmt"
	"github.com/RyanBard/gin-ex/internal/pkg/config"
	"github.com/RyanBard/gin-ex/internal/pkg/db"
	"github.com/RyanBard/gin-ex/internal/pkg/idgen"
	"github.com/RyanBard/gin-ex/internal/pkg/mdlw"
	"github.com/RyanBard/gin-ex/internal/pkg/org"
	"github.com/RyanBard/gin-ex/internal/pkg/timer"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"net/http"
)

func main() {
	log := logrus.StandardLogger()
	cfg, err := config.LoadConfig()
	if err != nil {
		log.WithError(err).Fatal("invalid config")
	}
	logLvl, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		log.WithError(err).Fatal("invalid log level")
	}
	log.SetLevel(logLvl)
	dbx := sqlx.MustConnect(
		"postgres",
		fmt.Sprintf(
			"user=%s password=%s dbname=%s sslmode=%s",
			cfg.DB.User,
			cfg.DB.Password,
			cfg.DB.DBName,
			cfg.DB.SSLMode,
		),
	)
	orgDAO := org.NewOrgDAO(log, dbx)
	txMGR := db.NewTXMGR(log, dbx)
	timer := timer.New()
	idGenerator := idgen.New()
	orgService := org.NewOrgService(log, orgDAO, txMGR, timer, idGenerator)
	validate := validator.New()
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

	r.Run(fmt.Sprintf(":%v", cfg.Port))
}
