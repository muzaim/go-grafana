package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-up-monitoring/internal/config"
	"go-up-monitoring/internal/database"
	"go-up-monitoring/internal/middleware"
	"go-up-monitoring/internal/modules/auth"
	"go-up-monitoring/internal/modules/heartbeat"
	"go-up-monitoring/internal/modules/incident"
	"go-up-monitoring/internal/modules/monitor"
	"go-up-monitoring/pkg/logger"
	"go-up-monitoring/pkg/metrics"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.LoadConfig("")
	if err != nil {
		panic("failed to load configuration: " + err.Error())
	}

	_, err = logger.InitLogger(logger.Config{
		Level:      cfg.Log.Level,
		Filename:   cfg.Log.Filename,
		MaxSizeMB:  cfg.Log.MaxSizeMB,
		MaxBackups: cfg.Log.MaxBackups,
		MaxAgeDays: cfg.Log.MaxAgeDays,
		Compress:   cfg.Log.Compress,
		Stdout:     true,
	})
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
	defer logger.Sync()

	logger.Info("starting Go Uptime Monitor API Server", zap.String("app", cfg.App.Name))

	// Connect to Database
	db, err := database.ConnectPostgres(cfg.Database)
	if err != nil {
		logger.Warn("database connection failed (continuing for healthz check)", zap.Error(err))
	} else {
		defer db.Close()
		_ = auth.AutoMigrate(db.DB)
		_ = monitor.AutoMigrate(db.DB)
		_ = heartbeat.AutoMigrate(db.DB)
		_ = incident.AutoMigrate(db.DB)
	}

	if cfg.App.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(metrics.GinMetricsMiddleware())

	// Health and Prometheus Metrics
	r.GET("/healthz", func(c *gin.Context) {
		dbStatus := "healthy"
		if db == nil || db.Ping() != nil {
			dbStatus = "disconnected"
		}
		c.JSON(http.StatusOK, gin.H{
			"status":   "ok",
			"database": dbStatus,
			"time":     time.Now().UTC(),
		})
	})
	r.GET("/metrics", metrics.PrometheusHandler())

	// API v1 group
	v1 := r.Group("/api/v1")
	authMw := middleware.AuthMiddleware(cfg.JWT.Secret)

	if db != nil {
		authRepo := auth.NewRepository(db.DB)
		authSvc := auth.NewService(authRepo, cfg.JWT)
		authHdlr := auth.NewHandler(authSvc)
		authHdlr.RegisterRoutes(v1, authMw)

		monRepo := monitor.NewRepository(db.DB)
		monSvc := monitor.NewService(monRepo)
		monHdlr := monitor.NewHandler(monSvc)
		monHdlr.RegisterRoutes(v1, authMw)

		hbRepo := heartbeat.NewRepository(db.DB)
		hbSvc := heartbeat.NewService(hbRepo, monRepo)
		hbHdlr := heartbeat.NewHandler(hbSvc)
		hbHdlr.RegisterRoutes(v1, authMw)
	}

	addr := fmt.Sprintf(":%d", cfg.App.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	go func() {
		logger.Info("HTTP server listening", zap.String("addr", addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server error", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down HTTP server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("server forced to shutdown", zap.Error(err))
	}
	logger.Info("server exited cleanly")
}
