package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/lincaiyong/arg"
	"github.com/lincaiyong/daemon/common"
	"github.com/lincaiyong/daemon/internal"
	"github.com/lincaiyong/log"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"syscall"
	"time"
)

//go:embed version
var version string

func main() {
	runningApps := map[string]*RunningApp{}
	binaryApps := map[string]*App{}
	toBeKilled := map[int]time.Time{}
	common.StartWorker(
		"daemon",
		version,
		"",
		func(i []string) error {
			if arg.BoolArg("init") {
				doInit()
				os.Exit(0)
			}
			if arg.BoolArg("kill") {
				doKill(context.Background())
				os.Exit(0)
			}
			if err := loadConfig(); err != nil {
				return err
			}
			return nil
		},
		func(ctx context.Context) {
			var err error
			if err = loadConfig(); err != nil {
				log.ErrorLog("fail to load config: %v", err)
				return
			}
			if err = runMakeCommand(ctx); err != nil {
				log.ErrorLog("fail to run make command: %v", err)
				return
			}
			if binaryApps, err = collectBinaryApps(); err != nil {
				log.ErrorLog("fail to collect binary apps: %v", err)
				return
			}
			if runningApps, err = collectRunningApps(ctx); err != nil {
				log.ErrorLog("fail to collect running apps: %v", err)
				return
			}
			if err = launchNewApps(ctx, binaryApps, runningApps); err != nil {
				log.ErrorLog("fail to launch new apps: %v", err)
				return
			}
			if err = updateToBeKilled(runningApps, toBeKilled); err != nil {
				log.ErrorLog("fail to clean old apps: %v", err)
				return
			}
			if err = runKillCommand(ctx, toBeKilled); err != nil {
				log.ErrorLog("fail to run kill: %v", err)
				return
			}
			if config.EnableNginx {
				if err = reloadNginx(ctx, runningApps); err != nil {
					log.ErrorLog("fail to reload nginx: %v", err)
					return
				}
			}
		},
	)
}

func updateRunningApps(runningApps map[string]*RunningApp, name string, pid, port int, modifiedTime time.Time) {
	app := App{
		Name:         name,
		Pid:          pid,
		Port:         port,
		ModifiedTime: modifiedTime,
	}
	if runningApps[name] == nil {
		runningApps[name] = &RunningApp{Newest: app}
	} else if runningApps[name].Newest.ModifiedTime.Before(modifiedTime) {
		runningApps[name].Others = append(runningApps[name].Others, runningApps[name].Newest)
		runningApps[name].Newest = app
	} else {
		runningApps[name].Others = append(runningApps[name].Others, app)
	}
}

func collectRunningApps(ctx context.Context) (map[string]*RunningApp, error) {
	output, err := exec.CommandContext(ctx, "ps", "aux").Output()
	if err != nil {
		return nil, err
	}
	result := make(map[string]*RunningApp)
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, config.BinDir) {
			fields := strings.Fields(line)
			if len(fields) < 11 {
				log.WarnLog("fields count is less than 11: %d, %s", len(fields), line)
				continue
			}
			var pid int
			pid, err = strconv.Atoi(fields[1])
			if err != nil {
				log.WarnLog("fail to convert pid %s to int: %s", fields[1], err)
				continue
			}
			name, port, modTime, parseErr := internal.ParseCommandLine(config.BinDir, line)
			if parseErr != nil {
				log.WarnLog("fail to parse command: %s, %v", line, parseErr)
				continue
			}
			updateRunningApps(result, name, pid, port, modTime)
		}
	}
	m := map[string]string{}
	for name, v := range result {
		m[name] = v.String()
	}
	b, _ := json.Marshal(m)
	log.InfoLog("collected running apps: %s", string(b))
	return result, nil
}

func collectBinaryApps() (map[string]*App, error) {
	items, err := os.ReadDir(config.BinDir)
	if err != nil {
		return nil, err
	}
	result := make(map[string]*App)
	for _, item := range items {
		if !item.IsDir() && !strings.HasPrefix(item.Name(), ".") {
			var mod time.Time
			mod, err = internal.LastModifiedTime(path.Join(config.BinDir, item.Name()))
			if err != nil {
				return nil, err
			}
			result[item.Name()] = &App{
				Name:         item.Name(),
				ModifiedTime: mod,
			}
		}
	}
	m := map[string]string{}
	for name, v := range result {
		m[name] = v.String()
	}
	b, _ := json.Marshal(m)
	log.InfoLog("collected binary apps: %s", string(b))
	return result, nil
}

func runMakeCommand(ctx context.Context) error {
	err := internal.RunCommand(ctx, config.RootDir, "make")
	if err != nil {
		return err
	}
	return nil
}

func launchNewApps(ctx context.Context, binaryApps map[string]*App, runningApps map[string]*RunningApp) error {
	for name, binaryApp := range binaryApps {
		runningApp, ok := runningApps[name]
		if !ok || runningApp.Newest.ModifiedTime.Before(binaryApp.ModifiedTime) {
			var port int
			var err error
			if config.ServerMap[name] {
				port, err = internal.PickUnusedPort()
				if err != nil {
					log.ErrorLog("fail to pick unused port: %v", err)
					return err
				}
			}
			cmdName, cmdArgs := internal.CommandNameArgs(
				fmt.Sprintf("%s/%s", config.BinDir, name),
				port,
				binaryApp.ModifiedTime,
				path.Join(config.LogDir, name+".log"),
			)
			cmd := exec.CommandContext(ctx, cmdName, cmdArgs...)
			cmd.Dir = fmt.Sprintf("%s/%s", config.AppDir, name)
			cmd.Env = append(os.Environ(), config.Env...)
			cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
			log.InfoLog("launch new app \"%s\" on port %d", name, port)
			if err = cmd.Start(); err != nil {
				log.ErrorLog("fail to run command: %v", err)
				continue
			}
			updateRunningApps(runningApps, name, cmd.Process.Pid, port, binaryApp.ModifiedTime)
			go func() {
				_ = cmd.Wait()
			}()
		}
	}
	return nil
}

func updateToBeKilled(runningApps map[string]*RunningApp, toBeKilled map[int]time.Time) error {
	allPids := map[int]bool{}
	for _, runningApp := range runningApps {
		allPids[runningApp.Newest.Pid] = true
		for _, app := range runningApp.Others {
			allPids[app.Pid] = true
			if _, ok := toBeKilled[app.Pid]; !ok {
				if config.ServerMap[app.Name] {
					toBeKilled[app.Pid] = time.Now().Add(time.Duration(config.KillDelay) * time.Second)
				} else {
					toBeKilled[app.Pid] = time.Now()
				}
			}
		}
	}
	for pid := range toBeKilled {
		if !allPids[pid] {
			delete(toBeKilled, pid)
		}
	}
	m := map[int]string{}
	for port, t := range toBeKilled {
		m[port] = t.Format(time.TimeOnly)
	}
	b, _ := json.Marshal(m)
	log.InfoLog("current toBeKilled: %s", string(b))
	return nil
}

func runKillCommand(ctx context.Context, toBeKilled map[int]time.Time) error {
	if len(toBeKilled) == 0 {
		return nil
	}
	now := time.Now()
	for pid, t := range toBeKilled {
		log.InfoLog("check pid=%d time=%s", pid, t.Format(time.TimeOnly))
		if t.Before(now) {
			log.InfoLog("kill app pid=%d", pid)
			err := exec.CommandContext(ctx, "kill", strconv.Itoa(pid)).Run()
			if err != nil {
				log.WarnLog("fail to exec command pid=%d: %v", pid, err)
			}
		}
	}
	return nil
}

func reloadNginx(ctx context.Context, runningApps map[string]*RunningApp) error {
	nginxApps, err := getNginxApps()
	if err != nil {
		return err
	}
	needReload := false
	toReload := make(map[string]int)
	for name, runningApp := range runningApps {
		if !config.ServerMap[name] {
			continue
		}
		if nginxApps[name] != runningApp.Newest.Port {
			needReload = true
		}
		toReload[name] = runningApp.Newest.Port
	}
	if needReload {
		err = doReloadNginx(ctx, toReload)
		if err != nil {
			return err
		}
	}
	return nil
}
