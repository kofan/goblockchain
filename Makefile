# https://gist.github.com/isaacs/62a2d1825d04437c6f08

MAKE_FILE_PATH := $(abspath $(lastword $(MAKEFILE_LIST)))
DIRECTORY_PATH := $(patsubst %/,%,$(dir $(MAKE_FILE_PATH)))
MAIN_GO = $(DIRECTORY_PATH)/src/cmd/main.go

.PHONY: all install run buildxw

all: install run

install:
	cd $(DIRECTORY_PATH) && dep ensure

run:
	go run $(MAIN_GO)

build:
	go build -o $(DIRECTORY_PATH)/bin/goblockchain $(MAIN_GO) 

