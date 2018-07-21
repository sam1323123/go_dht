# should only be run in the directory of this file
GOCMD=go
GOBUILD=$(GOCMD) build
DIR_NAME=$(notdir $(shell pwd)) # does not work if make executed outside this dir
BIN_NAME=$(strip $(DIR_NAME)).out

all: run

build:
	$(GOBUILD) -o $(BIN_NAME)

run: build
	./$(BIN_NAME)

clean:
	rm -f ./$(BIN_NAME)

