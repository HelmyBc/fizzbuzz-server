.PHONY: run build test lint docker

run:
	go run ./cmd/server

build:
	go build -o bin/server ./cmd/server

test:
	go test ./... -race -cover

lint:
	go vet ./...
	@staticcheck ./... 2>/dev/null || echo "staticcheck not installed — run: go install honnef.co/go/tools/cmd/staticcheck@latest"

docker:
	docker build -t fizzbuzz-server .
