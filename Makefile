-include .env

DEPENDENCIES = postgres postgres.init

.PHONY: default run build image dependencies test format

default: format run

run:
	@echo "==> Executing code.."
	@go run main.go

build:
	@echo "==> Compiling code.."
	go build main.go

image:
	@echo "==> Building image.."
	docker-compose build app

dependencies:
	@echo "==> Starting auxiliary containers.."
	docker-compose up -d ${DEPENDENCIES}

test:
	@echo "==> Running tests.."
	go test -v ./...

format:
	@echo "==> Formatting code.."
	go fmt ./...