.PHONY: up down build fresh logs tidy test seed clean help

# ── Lifecycle ─────────────────────────────────────────────────────────────────

## up: Start all services (uses cached images if available)
up:
	docker compose up -d
	@echo ""
	@echo "✅ System up!"
	@echo "   REST API  → http://localhost:8080/api/v1"
	@echo "   Kafka UI  → http://localhost:8090"

## fresh: Force-rebuild ALL images from scratch (use this first time or after code changes)
fresh:
	docker compose down --remove-orphans
	docker compose build --no-cache
	docker compose up -d
	@echo ""
	@echo "✅ Fresh build complete!"
	@echo "   REST API  → http://localhost:8080/api/v1"
	@echo "   Kafka UI  → http://localhost:8090"

## build: Rebuild images (uses cache where possible)
build:
	docker compose build

## down: Stop all services
down:
	docker compose down

## logs: Stream logs from all services
logs:
	docker compose logs -f

## logs-svc: Stream one service's logs (e.g. make logs-svc SVC=order-service)
logs-svc:
	docker compose logs -f $(SVC)

## restart: Restart one service (e.g. make restart SVC=payment-service)
restart:
	docker compose restart $(SVC)

# ── Development ───────────────────────────────────────────────────────────────

## tidy: Download and tidy Go modules locally
tidy:
	go mod tidy

## test: Run all unit tests
test:
	go test ./... -v -race -timeout 30s

## vet: Run go vet
vet:
	go vet ./...

## seed: Seed sample data (services must be running)
seed:
	@echo "── Register user ─────────────────────────────────────"
	@curl -s -X POST http://localhost:8080/api/v1/auth/register \
	  -H "Content-Type: application/json" \
	  -d '{"name":"Alice","email":"alice@example.com","password":"secret123"}' | python3 -m json.tool 2>/dev/null || true
	@echo ""
	@echo "── Login ─────────────────────────────────────────────"
	@TOKEN=$$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
	  -H "Content-Type: application/json" \
	  -d '{"email":"alice@example.com","password":"secret123"}' | python3 -c "import sys,json; print(json.load(sys.stdin)['token'])" 2>/dev/null); \
	echo "Token: $$TOKEN"; \
	echo ""; \
	echo "── Create product ────────────────────────────────────"; \
	PRODUCT_ID=$$(curl -s -X POST http://localhost:8080/api/v1/products \
	  -H "Authorization: Bearer $$TOKEN" \
	  -H "Content-Type: application/json" \
	  -d '{"name":"MacBook Pro","description":"M3 chip","price":2499.99,"stock":50}' | python3 -c "import sys,json; print(json.load(sys.stdin)['id'])" 2>/dev/null); \
	echo "Product ID: $$PRODUCT_ID"; \
	echo ""; \
	echo "── Place order ───────────────────────────────────────"; \
	curl -s -X POST http://localhost:8080/api/v1/orders \
	  -H "Authorization: Bearer $$TOKEN" \
	  -H "Content-Type: application/json" \
	  -d "{\"product_id\":\"$$PRODUCT_ID\",\"quantity\":1}" | python3 -m json.tool 2>/dev/null || true; \
	echo ""; \
	echo "✅ Seed done! Check Kafka UI → http://localhost:8090"

## clean: Remove containers, volumes, images
clean:
	docker compose down -v --rmi local

## help: Show this help
help:
	@grep -E '^##' Makefile | sed 's/## //'
