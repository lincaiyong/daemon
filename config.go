package main

import (
	"encoding/json"
	"fmt"
	"github.com/lincaiyong/log"
	"os"
	"runtime"
)

var config Config

type Config struct {
	Env           []string `json:"env"`
	LogPath       string
	SleepInterval int
	KillDelay     int
	RootDir       string
	LogDir        string
	BinDir        string
	AppDir        string
	Servers       []string `json:"servers"`
	ServerMap     map[string]bool
	NginxConfig
}

type NginxConfig struct {
	EnableNginx     bool `json:"enable_nginx"`
	NginxConfDDir   string
	NginxConfFile   string
	NoAuthServers   []string `json:"no_auth_servers"`
	NoAuthServerMap map[string]bool
	SecretToken     string `json:"secret_token"`
	EnableHttps     bool   `json:"enable_https"`
	Domain          string `json:"domain"`
	SSLDir          string
}

func loadConfig() error {
	cwd, _ := os.Getwd()
	log.InfoLog("cwd: %s", cwd)
	// daemon.json
	if b, err := os.ReadFile("daemon.json"); err != nil {
		return fmt.Errorf("fail to read %s/daemon.json: %v", cwd, err)
	} else {
		err = json.Unmarshal(b, &config)
		if err != nil {
			return fmt.Errorf("fail to parse daemon.json: %v", err)
		}
	}
	// default
	config.LogPath = "daemon.log"
	config.SleepInterval = 10
	config.KillDelay = 30
	config.RootDir = cwd
	// nginx default
	config.NginxConfDDir = "/etc/nginx/conf.d"
	if runtime.GOOS == "darwin" {
		config.NginxConfDDir = "/opt/homebrew/etc/nginx/conf.d"
	}
	config.NginxConfFile = "/etc/nginx/nginx.conf"
	if runtime.GOOS == "darwin" {
		config.NginxConfFile = "/opt/homebrew/etc/nginx/nginx.conf"
	}
	config.SSLDir = fmt.Sprintf("%s/ssl", config.RootDir)
	// compute
	config.LogDir = fmt.Sprintf("%s/log", config.RootDir)
	config.BinDir = fmt.Sprintf("%s/bin", config.RootDir)
	config.AppDir = fmt.Sprintf("%s/app", config.RootDir)
	config.ServerMap = make(map[string]bool)
	for _, server := range config.Servers {
		config.ServerMap[server] = true
	}
	config.NoAuthServerMap = make(map[string]bool)
	for _, server := range config.NoAuthServers {
		config.NoAuthServerMap[server] = true
	}
	return nil
}
