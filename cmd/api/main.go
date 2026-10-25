package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"openingjobs/internal/opening"
	"openingjobs/internal/platform/http/middleware"
	tenantpostgres "openingjobs/internal/tenant/infrastructure/postgres"
	tenanthttp "openingjobs/internal/tenant/transport/http"
	"openingjobs/pkg/config"

	"github.com/gin-gonic/gin"
)

func main() {
	logger := config.GetLogger()

	if err := config.Init(); err != nil {
		logger.Errorf("config initialization error: %v", err)
		return
	}

	defer config.GetDB().Close()
	requestTimeout, err := config.GetDuration("HTTP_REQUEST_TIMEOUT", 10*time.Second)
	if err != nil {
		logger.Errorf("invalid HTTP request timeout: %v", err)
		return
	}

	router := gin.New()
	if err := router.SetTrustedProxies(nil); err != nil {
		logger.Errorf("configure trusted proxies: %v", err)
		return
	}

	router.Use(gin.Logger(), gin.Recovery())
	router.Use(middleware.RequestTimeout(requestTimeout))
	router.Use(func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20)
		c.Next()
	})

	tenants := tenantpostgres.NewRepository(config.GetDB())
	api := router.Group("/api/v1")
	api.Use(tenanthttp.RequireTenant(tenants))
	opening.MountHTTPV1(api, config.GetDB())

	server := &http.Server{
		Addr:              config.GetEnv("HTTP_ADDR", ":8080"),
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		logger.Infof("HTTP server listening on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Errorf("HTTP server failed: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Errorf("graceful shutdown failed: %v", err)
	}
}
