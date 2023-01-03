package main

import (
	"fmt"
	"github.com/RyanBard/go-service-ex/internal/config"
	"github.com/RyanBard/go-service-ex/internal/idgen"
	"github.com/RyanBard/go-service-ex/internal/mdlw"
	"github.com/RyanBard/go-service-ex/internal/org"
	"github.com/RyanBard/go-service-ex/internal/timer"
	"github.com/RyanBard/go-service-ex/internal/tx"
	"github.com/RyanBard/go-service-ex/internal/user"
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

	timer := timer.New()
	idGenerator := idgen.New()

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

	txMGR := tx.NewTXMGR(log, dbx)

	validate := validator.New()

	orgDAO := org.NewDAO(log, cfg.DB.QueryTimeout, dbx)
	orgService := org.NewService(log, orgDAO, txMGR, timer, idGenerator)
	orgCtrl := org.NewController(log, validate, orgService)

	userDAO := user.NewDAO(log, cfg.DB.QueryTimeout, dbx)
	userService := user.NewService(log, orgService, userDAO, txMGR, timer, idGenerator)
	userCtrl := user.NewController(log, validate, userService)

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
	authorized.Use(mdlw.Auth(log, cfg.AuthConfig))

	adminPriv := r.Group("/api")
	adminPriv.Use(mdlw.Auth(log, cfg.AuthConfig))
	adminPriv.Use(mdlw.RequiresAdmin(log))

	authorized.GET("/orgs/:id", orgCtrl.GetByID)
	authorized.GET("/orgs", orgCtrl.GetAll)

	adminPriv.POST("/orgs", orgCtrl.Save)
	adminPriv.PUT("/orgs", orgCtrl.Save)
	adminPriv.POST("/orgs/:id", orgCtrl.Save)
	adminPriv.PUT("/orgs/:id", orgCtrl.Save)
	adminPriv.DELETE("/orgs/:id", orgCtrl.Delete)

	authorized.GET("/orgs/:id/users", userCtrl.GetAllByOrgID)

	authorized.GET("/users/:id", userCtrl.GetByID)
	authorized.GET("/users", userCtrl.GetAll)

	adminPriv.POST("/users", userCtrl.Save)
	adminPriv.PUT("/users", userCtrl.Save)
	adminPriv.POST("/users/:id", userCtrl.Save)
	adminPriv.PUT("/users/:id", userCtrl.Save)
	adminPriv.DELETE("/users/:id", userCtrl.Delete)

	r.Run(fmt.Sprintf(":%v", cfg.Port))
}
