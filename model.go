package main

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type App struct {
	Name         string
	Pid          int
	Port         int
	ModifiedTime time.Time
}

func (a *App) String() string {
	return fmt.Sprintf("%s[pid:%d,port:%d,%s]", a.Name, a.Pid, a.Port, a.ModifiedTime.Format(time.TimeOnly))
}

type RunningApp struct {
	Newest App
	Others []App
}

func (r *RunningApp) String() string {
	if len(r.Others) > 0 {
		items := make([]string, 0, len(r.Others))
		for _, app := range r.Others {
			items = append(items, app.String())
		}
		sort.Strings(items)
		return fmt.Sprintf("%s[%s]", r.Newest.String(), strings.Join(items, ","))
	}
	return r.Newest.String()
}
