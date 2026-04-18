# Simple Makefile for Auth Service

APP_NAME=auth-service
PORT=8080

.PHONY: all build run clean test kill

all: build

build:
	go build -o bin/$(APP_NAME) cmd/api/main.go

run: kill
	go run cmd/api/main.go

kill:
	@echo "Stopping existing process on port $(PORT)..."
	@lsof -t -i:$(PORT) | xargs kill -9 2>/dev/null || true

clean:
	rm -rf bin/

test:
	go test ./... -v
