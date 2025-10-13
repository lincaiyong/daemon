package common

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/lincaiyong/log"
	"github.com/lincaiyong/processlock"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var interval = 10

func SetInterval(i int) {
	interval = i
}

func StartWorker(
	name, version, requiredEnvs string,
	initFunc func([]string) error,
	worker func(context.Context),
) {
	initFuncWrap := func(envs []string, _ *gin.RouterGroup) error {
		return initFunc(envs)
	}
	logPath, _ := startup(name, version, requiredEnvs, initFuncWrap, nil)
	if err := processlock.Lock(logPath); err != nil {
		log.ErrorLog("fail to acquire process lock: %v", err)
		os.Exit(1)
	}
	defer processlock.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{}, 1)
	go func() {
		log.InfoLog("start to run worker")
		defer func() { done <- struct{}{} }()
		first := true
		for {
			if first {
				first = false
			} else {
				for i := 0; i < interval; i++ {
					select {
					case <-ctx.Done():
						return
					case <-time.After(1 * time.Second):
					}
				}
			}
			log.InfoLog("------%s------", time.Now().Format(time.TimeOnly))
			worker(ctx)
		}
	}()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-sigs:
		log.InfoLog("receive shutdown signal")
		cancel()
		select {
		case <-done:
			log.InfoLog("gracefully shutdown")
		case <-time.After(time.Minute):
			log.ErrorLog("force shutdown")
			os.Exit(1)
		}
	case <-done:
		log.InfoLog("done")
	}
}
