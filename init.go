package main

import (
	"fmt"
	"os"
)

var initConfigContent = `{
	"servers": ["demo"],
	"enable_nginx": true,
	"no_auth_servers": ["demo"],
	"secret_token": "",
	"enable_https": false,
	"domain": "localhost"
}`

var initMakefileContent = `
.PHONY: all
all: bin/demo
	cd app/demo && git pull

bin/demo: $(shell find app/demo -name '*.go')
	@if [ ! -d "app/demo" ]; then git clone https://github.com/lincaiyong/demo app/demo; fi
	cd app/demo && go build -o ../../bin/demo .

.PHONY: ps
ps:
	@ps aux | grep $(CURDIR)/bin/

.PHONY: kill
kill:
	@ps aux | grep $(CURDIR)/bin/ | grep -v grep | awk '{print $$2}'
	@ps aux | grep $(CURDIR)/bin/ | grep -v grep | awk '{print $$2}' | xargs -r kill
`

func doInit() {
	fmt.Println("creating ./app")
	if err := os.MkdirAll("app", os.ModePerm); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("creating ./bin")
	if err := os.MkdirAll("bin", os.ModePerm); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("creating ./log")
	if err := os.MkdirAll("log", os.ModePerm); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("creating ./daemon.json")
	if err := os.WriteFile("daemon.json", []byte(initConfigContent), 0644); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("creating ./Makefile")
	if err := os.WriteFile("Makefile", []byte(initMakefileContent), 0644); err != nil {
		fmt.Println(err)
	}
	fmt.Println("done")
}
