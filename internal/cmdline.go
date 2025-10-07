package internal

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

func CommandNameArgs(binaryFile string, port int, modTime time.Time) (string, []string) {
	return binaryFile, []string{
		fmt.Sprintf("--port=%d", port),
		fmt.Sprintf("--time=%s", timeToBeijingTimeStr(modTime)),
	}
}

var commandLineRegex *regexp.Regexp

func ParseCommandLine(binaryDir, line string) (name string, port int, modTime time.Time, err error) {
	if commandLineRegex == nil {
		pattern := fmt.Sprintf(`%s/([a-z]+) --port=(\d+) --time=(\d\d\d\d_\d\d_\d\d_\d\d_\d\d_\d\d)`, binaryDir)
		commandLineRegex = regexp.MustCompile(pattern)
	}
	ret := commandLineRegex.FindStringSubmatch(line)
	if len(ret) != 4 {
		err = fmt.Errorf("fail to match command line: %s", line)
		return
	}
	port, _ = strconv.Atoi(ret[2])
	return ret[1], port, beijingTimeStrToTime(ret[3]), nil
}
