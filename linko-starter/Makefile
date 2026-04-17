BINARY_NAME=linko

GIT_SHA := $(shell git rev-parse HEAD)
BUILD_TIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
BIN_PATH := ./bin/$(BINARY_NAME)

build:
	go build \
		-ldflags "-X xonoxc/linko/internal/build.GIT_SHA=$(GIT_SHA) -X xonoxc/linko/internal/build.BUILD_TIME=$(BUILD_TIME)"\
		-o $(BIN_PATH)

run: build
	./$(BIN_PATH)

clean:
	rm -f $(BIN_PATH)
