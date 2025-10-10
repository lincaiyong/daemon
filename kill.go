package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
)

func doKill() {
	logPath := "daemon.log"
	b, err := os.ReadFile(logPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	content := string(b)
	results := regexp.MustCompile(`\[INFO ] pid: (\d+)`).FindAllStringSubmatch(content, -1)
	if len(results) == 0 {
		fmt.Printf("no pid found in %s\n", logPath)
		os.Exit(0)
	}
	pid := results[len(results)-1][1]
	out, err := exec.Command("kill", pid).CombinedOutput()
	if err != nil {
		fmt.Printf("fail to kill: %v, %s", err, string(out))
		os.Exit(1)
	}
	fmt.Println(string(out))
	fmt.Println("done")
}
