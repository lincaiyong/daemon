package internal

import (
	"bufio"
	"bytes"
	"github.com/lincaiyong/log"
	"os/exec"
	"strings"
	"sync"
)

func RunCommand(workDir, cmdName string, cmdArgs ...string) (string, error) {
	log.InfoLog("run command: %s %s", cmdName, strings.Join(cmdArgs, " "))
	cmd := exec.Command(cmdName, cmdArgs...)
	cmd.Dir = workDir
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.ErrorLog("fail to get stdout pipe: %v", err)
		return "", err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.ErrorLog("fail to get stderr pipe: %v", err)
		return "", err
	}
	if err = cmd.Start(); err != nil {
		log.ErrorLog("fail to start command: %v", err)
		return "", err
	}

	var wg sync.WaitGroup
	var stdoutBuf, stderrBuf bytes.Buffer

	wg.Add(1)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			stdoutBuf.WriteString(line + "\n")
			log.InfoLog("%s", line)
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			stderrBuf.WriteString(line + "\n")
			log.InfoLog("stderr: %s", line)
		}
	}()
	wg.Wait()
	if err = cmd.Wait(); err != nil {
		log.ErrorLog("fail to wait command: %v", err)
		return "", err
	}
	log.InfoLog("exit code: %d", cmd.ProcessState.ExitCode())
	return "", nil
}
