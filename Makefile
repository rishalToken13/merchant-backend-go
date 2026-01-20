# =========================
# Variables
# =========================
DB_DSN ?= postgres://app:app@localhost:5432/paydb?sslmode=disable
MIGRATIONS_PATH = internal/repository/postgres/migrations

# =========================
# App
# =========================
run-api:
	go run ./cmd/api

run-worker:
	go run ./cmd/worker

test:
	go test ./...

# =========================
# Migrations
# =========================
migrate-up:
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_DSN)" up

migrate-down:
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_DSN)" down 1

migrate-force:
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_DSN)" force $(v)

# =========================
# Docker
# =========================
docker-up:
	docker compose -f deployments/docker-compose.yml up -d

docker-down:
	docker compose -f deployments/docker-compose.yml down

