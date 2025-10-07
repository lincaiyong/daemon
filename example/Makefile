

.PHONY: all
all: bin/demo

bin/demo: $(shell find app/demo -name '*.go')
	cd app/demo && go build -o ../../bin/demo .

.PHONY: ps
ps:
	@ps aux | grep /example/bin/

.PHONY: kill
kill:
	@ps aux | grep /example/bin/ | grep -v grep | awk '{print $$2}'
	@ps aux | grep /example/bin/ | grep -v grep | awk '{print $$2}' | xargs -r kill
