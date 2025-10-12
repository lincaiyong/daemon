package app

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/lincaiyong/log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func StartServer(
	name, version, requiredEnvs string,
	initFunc func([]string, *gin.RouterGroup) error,
) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		start := time.Now()
		log.InfoLog(" %s | %s", c.Request.URL.Path, c.ClientIP())
		c.Next()
		log.InfoLog(" %s | %s | %v | %d", c.Request.URL.Path, c.ClientIP(), time.Since(start), c.Writer.Status())
	})
	_, port := startup(name, version, requiredEnvs, initFunc, &router.RouterGroup)
	server := &http.Server{Addr: fmt.Sprintf("127.0.0.1:%s", port), Handler: router}
	go func() {
		log.InfoLog("start to run server on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.ErrorLog("fail to start: %v", err)
		}
	}()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	log.InfoLog("receive shutdown signal")

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		// <-c.Request.Context().Done()
		log.ErrorLog("shutdown with error: %v", err)
	} else {
		log.InfoLog("gracefully shutdown")
	}
}
