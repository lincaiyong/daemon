package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/lincaiyong/arg"
	"github.com/lincaiyong/log"
	"os"
	"path"
	"runtime"
)

var config Config

type Config struct {
	LogPath       string `json:"log_path"`
	SleepInterval int    `json:"sleep_interval"`
	KillDelay     int    `json:"kill_delay"`
	RootDir       string `json:"root_dir"`
	BinDir        string
	AppDir        string
	SSLDir        string
	LogDir        string
	Workers       []string `json:"workers"`
	WorkerMap     map[string]bool
	NginxConfDDir string   `json:"nginx_conf_d_dir"`
	NginxConfFile string   `json:"nginx_conf_file"`
	EnableHttps   bool     `json:"enable_https"`
	Domain        string   `json:"domain"`
	SecretToken   string   `json:"secret_token"`
	AuthApps      []string `json:"auth_apps"`
	AuthAppMap    map[string]bool
}

func loadConfig() error {
	arg.Parse()
	// config.json
	if configFile := arg.KeyValueArg("config", "daemon.json"); configFile != "" {
		if b, err := os.ReadFile(configFile); err != nil {
			return fmt.Errorf("fail to read config file: %v", err)
		} else {
			err = json.Unmarshal(b, &config)
			if err != nil {
				return fmt.Errorf("fail to parse config file: %v", err)
			}
		}
	}
	// default
	if config.LogPath == "" {
		config.LogPath = "/tmp/daemon.log"
	}
	if config.KillDelay == 0 {
		config.KillDelay = 30
	}
	if config.NginxConfDDir == "" {
		config.NginxConfDDir = "/etc/nginx/conf.d"
		if runtime.GOOS == "darwin" {
			config.NginxConfDDir = "/opt/homebrew/etc/nginx/conf.d"
		}
	}
	if config.NginxConfFile == "" {
		config.NginxConfFile = "/etc/nginx/nginx.conf"
		if runtime.GOOS == "darwin" {
			config.NginxConfFile = "/opt/homebrew/etc/nginx/nginx.conf"
		}
	}
	if config.Domain == "" {
		config.Domain = "localhost"
	}
	// check
	if config.RootDir == "" {
		config.RootDir, _ = os.Getwd()
	}
	config.BinDir = fmt.Sprintf("%s/bin", config.RootDir)
	config.AppDir = fmt.Sprintf("%s/app", config.RootDir)
	config.SSLDir = fmt.Sprintf("%s/ssl", config.RootDir)
	config.LogDir = fmt.Sprintf("%s/log", config.RootDir)
	_ = os.MkdirAll(config.LogDir, os.ModePerm)
	config.WorkerMap = make(map[string]bool)
	for _, worker := range config.Workers {
		config.WorkerMap[worker] = true
	}
	config.AuthAppMap = make(map[string]bool)
	for _, app := range config.AuthApps {
		config.AuthAppMap[app] = true
	}
	var err error
	hasFileErr := false
	for _, dir := range []string{
		config.RootDir, config.BinDir, config.AppDir, config.LogDir, config.NginxConfDDir, config.NginxConfFile,
		path.Join(config.RootDir, "Makefile"),
	} {
		if _, err = os.Stat(dir); err != nil {
			hasFileErr = true
			log.ErrorLog("%s does not exist", dir)
		}
	}
	if hasFileErr {
		return errors.New("some file or directory are required")
	}
	if config.EnableHttps {
		if _, err = os.Stat(config.SSLDir); err != nil {
			return fmt.Errorf("%s does not exist", config.SSLDir)
		}
	}
	return nil
}
