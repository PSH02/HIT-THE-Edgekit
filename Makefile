.PHONY: build run test lint proto migrate docker-up docker-down

BINARY := server
CONFIG ?= configs/config.local.yaml

build:
	go build -o bin/$(BINARY) ./cmd/server

run: build
	./bin/$(BINARY) -config $(CONFIG)

test:
	go test -v -race -count=1 ./...

lint:
	golangci-lint run ./...

proto:
	cd proto && buf generate

migrate:
	./scripts/migrate.sh

docker-up:
	docker compose -f deploy/docker-compose.yml up -d --build

docker-down:
	docker compose -f deploy/docker-compose.yml down

tidy:
	go mod tidy

fmt:
	gofumpt -l -w .

vet:
	go vet ./...

clean:
	rm -rf bin/
