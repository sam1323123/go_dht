# should only be run in the directory of this file
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOINSTALL=$(GOINSTALL) install
DIR_NAME=$(notdir $(shell pwd)) # does not work if make executed outside this dir
BIN_NAME=$(strip $(DIR_NAME)).out

all: test

test:
	$(GOTEST)

build:
	$(GOBUILD)

install:
	$(GOINSTALL)

clean:
	rm -f ./$(BIN_NAME)

