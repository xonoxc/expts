# durable-queue — durable job queue backed by SQLite (Go)
# Module: durqueue
# Source files: main.go, job/job.go, que/queue.go, store/store.go

BINARY_NAME := durable-queue
BUILD_DIR   := bin

GO_FILES := main.go \
		job/job.go \
		que/queue.go \
		store/store.go

.PHONY: all build run test clean

all: build

build:
	go build -o $(BUILD_DIR)/$(BINARY_NAME) .

run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

test:
	go test ./...

clean:
	rm -rf $(BUILD_DIR)
