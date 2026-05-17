# Makefile для URL Shortener

.PHONY: build run clean test help

VERSION := v1.0.0
COMMIT := manual
BUILD_DATE := $(shell powershell -Command "Get-Date -Format 'yyyy-MM-dd_HH:mm:ss'")

LDFLAGS := -X 'main.buildVersion=$(VERSION)' -X 'main.buildDate=$(BUILD_DATE)' -X 'main.buildCommit=$(COMMIT)'

build:
	go build -ldflags "$(LDFLAGS)" -o bin/shortener.exe ./cmd/shortener

build-simple:
	go build -o bin/shortener.exe ./cmd/shortener

run: build
	./bin/shortener.exe

clean:
	if exist bin rmdir /s /q bin

test:
	go test -v ./...

version:
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT)"
	@echo "Build Date: $(BUILD_DATE)"

help:
	@echo "Доступные команды:"
	@echo "  make build    - Собрать приложение с версией"
	@echo "  make run      - Собрать и запустить"
	@echo "  make clean    - Очистить бинарники"
	@echo "  make test     - Запустить тесты"
	@echo "  make version  - Показать информацию о версии"