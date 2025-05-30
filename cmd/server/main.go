package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	logutil "github.com/RyanBard/go-log-util/pkg"
	"github.com/RyanBard/go-service-ex/internal/config"
	"github.com/RyanBard/go-service-ex/internal/idgen"
	"github.com/RyanBard/go-service-ex/internal/mdlw"
	"github.com/RyanBard/go-service-ex/internal/org"
	"github.com/RyanBard/go-service-ex/internal/timer"
	"github.com/RyanBard/go-service-ex/internal/tx"
	"github.com/RyanBard/go-service-ex/internal/user"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	log := slog.Default()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.With(logutil.LogAttrError(err)).Error("invalid config")
		panic(err)
	}

	lvl, err := logutil.ParseLevel(cfg.LogLevel)
	if err != nil {
		log.With(
			logutil.LogAttrError(err),
			slog.String("logLevel", cfg.LogLevel),
		).Error("invalid log level")
		panic(err)
	}
	slogOpts := slog.HandlerOptions{
		Level: lvl,
	}

	if cfg.Mode != "local" {
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slogOpts))
	} else {
		log = slog.New(slog.NewTextHandler(os.Stdout, &slogOpts))
	}

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

	orgDAO := org.NewDAO(log, cfg.DB.QueryTimeout, dbx)
	orgService := org.NewService(log, orgDAO, txMGR, timer, idGenerator)
	orgCtrl := org.NewController(log, orgService)

	userDAO := user.NewDAO(log, cfg.DB.QueryTimeout, dbx)
	userService := user.NewService(log, orgService, userDAO, txMGR, timer, idGenerator)
	userCtrl := user.NewController(log, userService)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(mdlw.ReqID(log))

	r.GET("/health", func(c *gin.Context) {
		log.With(logutil.LogAttrReqID(c.Request.Context())).Debug("Health called")
		c.Status(http.StatusNoContent)
	})

	r.GET("/readiness", func(c *gin.Context) {
		log.With(logutil.LogAttrReqID(c.Request.Context())).Debug("Readiness called")
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
