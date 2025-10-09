package internal

import (
	"bufio"
	"fmt"
	"github.com/lincaiyong/log"
	"os/exec"
	"strings"
	"sync"
)

func RunCommand(workDir, cmdName string, cmdArgs ...string) error {
	log.InfoLog("run command: %s %s", cmdName, strings.Join(cmdArgs, " "))
	cmd := exec.Command(cmdName, cmdArgs...)
	cmd.Dir = workDir
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.ErrorLog("fail to get stdout pipe: %v", err)
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.ErrorLog("fail to get stderr pipe: %v", err)
		return err
	}
	if err = cmd.Start(); err != nil {
		log.ErrorLog("fail to start command: %v", err)
		return err
	}

	var wg sync.WaitGroup
	//var stdoutBuf, stderrBuf bytes.Buffer
	errChan := make(chan error, 2)

	wg.Add(1)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			//stdoutBuf.WriteString(line + "\n")
			log.InfoLog("[STDOUT] %s", line)
		}
		if err = scanner.Err(); err != nil {
			errChan <- err
			log.ErrorLog("error reading stdout: %v", err)
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			//stderrBuf.WriteString(line + "\n")
			log.ErrorLog("[STDERR] %s", line)
		}
		if err = scanner.Err(); err != nil {
			errChan <- err
			log.ErrorLog("error reading stderr: %v", err)
		}
	}()
	wg.Wait()

	close(errChan)
	for err = range errChan {
		if err != nil {
			return err
		}
	}

	if err = cmd.Wait(); err != nil {
		log.ErrorLog("fail to wait command: %v", err)
		return err
	}
	exitCode := cmd.ProcessState.ExitCode()
	if exitCode != 0 {
		log.ErrorLog("command failed with exit code: %d", exitCode)
		return fmt.Errorf("command failed with exit code: %d", exitCode)
	}
	log.InfoLog("command completed successfully")
	return nil
}
