package main

import "time"

type App struct {
	Name         string
	Pid          int
	Port         int
	ModifiedTime time.Time
}

type RunningApp struct {
	Newest App
	Others []App
}
