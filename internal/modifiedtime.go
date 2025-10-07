package internal

import (
	"os"
	"time"
)

func LastModifiedTime(filePath string) (time.Time, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return time.Time{}, err
	}
	modified := fileInfo.ModTime()
	modified = modified.Truncate(time.Second)
	return modified, nil
}
