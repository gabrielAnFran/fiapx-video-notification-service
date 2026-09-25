//go:build integration

package integration

// Placeholder for future testcontainers-go integration tests (real Postgres
// + RabbitMQ, and ideally Mailhog for a real SMTP round trip) covering
// NotificationRepository/ProcessedEventRepository and the messaging.Conn
// publish/consume/retry/DLQ path. Deferred as a stretch goal — not
// implemented yet, so it is intentionally left empty.
//
// It lives behind the `integration` build tag so `go test ./...` (this
// service's default, dependency-free test run) never tries to compile or
// run it.
