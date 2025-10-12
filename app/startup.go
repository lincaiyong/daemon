package app

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/lincaiyong/arg"
	"github.com/lincaiyong/log"
	"os"
	"strings"
)

func startup(
	name, version, requiredEnvs string,
	initFunc func([]string, *gin.RouterGroup) error,
	r *gin.RouterGroup,
) (logPath string, port string) {
	arg.Parse()
	if arg.BoolArg("version") {
		fmt.Println(version)
		os.Exit(0)
	}
	port = arg.KeyValueArg("port", "9123")
	logPath = arg.KeyValueArg("logpath", name+".log")
	if err := log.SetLogPath(logPath); err != nil {
		log.ErrorLog("fail to set log file path: %v", err)
		os.Exit(1)
	}
	log.InfoLog("version: %s", version)
	log.InfoLog("cmd line: %s", strings.Join(os.Args, " "))
	log.InfoLog("log path: %v", logPath)
	log.InfoLog("port: %s", port)
	log.InfoLog("pid: %d", os.Getpid())
	wd, _ := os.Getwd()
	log.InfoLog("work dir: %s", wd)

	keys := strings.Split(requiredEnvs, ",")
	envs := make([]string, len(keys))
	for i, key := range keys {
		value := os.Getenv(key)
		if value == "" {
			log.ErrorLog("env %s is required", key)
			os.Exit(1)
		}
		envs[i] = value
	}

	err := initFunc(envs, r)
	if err != nil {
		log.ErrorLog("fail to init: %v", err)
		os.Exit(1)
	}

	return
}
