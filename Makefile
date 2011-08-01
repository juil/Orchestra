GOPATH=$(shell pwd)
GOINSTALL_FLAGS=-dashboard=false -clean=true

all: build

build:
	GOPATH=$(GOPATH) goinstall $(GOINSTALL_FLAGS) conductor
	GOPATH=$(GOPATH) goinstall $(GOINSTALL_FLAGS) player
	GOPATH=$(GOPATH) goinstall $(GOINSTALL_FLAGS) submitjob
	GOPATH=$(GOPATH) goinstall $(GOINSTALL_FLAGS) getstatus
