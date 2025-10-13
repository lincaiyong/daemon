
.PHONY: all
all: git_pull bin/demo

.PHONY: git_pull
git_pull:
	@if [ -d "app/demo" ]; then cd app/demo && git pull; fi

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
