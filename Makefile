.PHONY: run build

run:
	docker compose up -d
	go run cmd/server/main.go
	cd frontend && npm run dev

build:
	go build -o rim cmd/server/main.go 