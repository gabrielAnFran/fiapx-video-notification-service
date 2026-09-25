.PHONY: build test test-integration lint cover docker helm-lint run

build:
	go build ./...

test:
	go test ./...

test-integration:
	go test -tags integration ./tests/integration/...

lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not found, falling back to go vet + gofmt"; \
		go vet ./...; \
		test -z "$$(gofmt -l .)" || (gofmt -l . && exit 1); \
	fi

cover:
	go test ./... -coverprofile=coverage.out
	go tool cover -func=coverage.out | tail -1

docker:
	docker build -t fiapx-video-notification-service:server --build-arg TARGET=server .
	docker build -t fiapx-video-notification-service:worker --build-arg TARGET=worker .

helm-lint:
	@if command -v helm >/dev/null 2>&1; then \
		helm lint charts/notification-service; \
	else \
		echo "helm not installed, skipping lint"; \
	fi

run:
	NOTIFICATION_PORT=$${NOTIFICATION_PORT:-8083} \
	NOTIFICATION_DB_DSN=$${NOTIFICATION_DB_DSN:-"host=localhost user=postgres password=postgres dbname=notification_service port=5432 sslmode=disable"} \
	NOTIFICATION_AMQP_URL=$${NOTIFICATION_AMQP_URL:-"amqp://guest:guest@localhost:5672/"} \
	SMTP_HOST=$${SMTP_HOST:-} \
	SMTP_PORT=$${SMTP_PORT:-} \
	SMTP_USER=$${SMTP_USER:-} \
	SMTP_PASSWORD=$${SMTP_PASSWORD:-} \
	SMTP_FROM_EMAIL=$${SMTP_FROM_EMAIL:-"no-reply@fiapx.local"} \
	go run ./cmd/server
