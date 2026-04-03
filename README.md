# 🛒 E-Commerce Microservices — Go + gRPC + Kafka + Docker

A production-style microservices backend built in Go, demonstrating distributed systems patterns including synchronous gRPC communication, asynchronous Kafka messaging, JWT auth, and Docker containerisation.

---

## Architecture

```
           ┌─────────────────────────────────────────────────┐
           │                  CLIENT (curl / Postman)         │
           └───────────────────────┬─────────────────────────┘
                                   │  HTTP/REST
                    ┌──────────────▼──────────────┐
                    │         API Gateway          │
                    │     (Gin • Port 8080)        │
                    └──┬──────┬──────┬──────┬──────┘
                       │gRPC  │gRPC  │gRPC  │gRPC
              ┌────────▼─┐ ┌──▼────┐ ┌▼────────┐ ┌▼──────────┐
              │  User    │ │Product│ │  Order  │ │  Payment  │
              │ Service  │ │Service│ │ Service │ │  Service  │
              │ :50051   │ │:50052 │ │ :50053  │ │  :50054   │
              └────┬─────┘ └──┬────┘ └──┬──────┘ └──────┬────┘
                   │          │         │ publish        │ publish
                   │          │  ┌──────▼────────────────▼────┐
                   │          │  │         Apache Kafka        │
                   │          │  │  • order.created           │
                   │          │  │  • payment.processed       │
                   │          │  │  • payment.failed          │
                   │          │  └──────────────┬─────────────┘
                   │          │                 │ consume
                   │          │        ┌────────▼────────┐
                   │          │        │  Notification   │
                   │          │        │    Service      │
                   │          │        └─────────────────┘
                   │          │
              ┌────▼──────────▼─────────────┐
              │        PostgreSQL            │
              │  (shared; one schema/svc)    │
              └─────────────────────────────┘
```

### Service responsibilities

| Service | Port | Role |
|---|---|---|
| **api-gateway** | 8080 | REST → gRPC translation, JWT validation, CORS |
| **user-service** | 50051 | Registration, login, JWT issuance, user profiles |
| **product-service** | 50052 | Product catalog, stock management (atomic deduction) |
| **order-service** | 50053 | Order creation, publishes `order.created` to Kafka |
| **payment-service** | 50054 | Consumes `order.created`, processes payments, publishes result |
| **notification-service** | — | Consumes payment events, sends email/SMS (simulated) |

### Infrastructure

| Component | Purpose |
|---|---|
| **PostgreSQL 16** | Persistent storage for all services |
| **Apache Kafka** | Async event bus (order → payment → notification) |
| **Kafka UI** | Browse topics and messages at `localhost:8090` |

---

## Prerequisites

- [Docker Desktop](https://www.docker.com/products/docker-desktop/) (or Docker + Docker Compose)
- [GoLand](https://www.jetbrains.com/go/) (JetBrains IDE — you're using this)
- `make` (built into macOS/Linux; Windows: use Git Bash or WSL)
- `curl` + [`jq`](https://jqlang.github.io/jq/) for the seed script (optional)

---

## Quick Start

```bash
# 1. Clone / open the project in GoLand
cd ecommerce-microservices

# 2. Start everything
make up

# 3. Wait ~30 seconds for Kafka to be ready, then seed sample data
make seed

# 4. Browse Kafka topics
open http://localhost:8090
```

---

## Running in GoLand (JetBrains)

### Option A — Run via Docker Compose (recommended)

1. Open the project folder in GoLand.
2. In the **Project** panel, right-click `docker-compose.yml` → **Run 'docker-compose.yml'**.
3. GoLand will show all containers in the **Services** tab.

### Option B — Run individual services locally (for debugging)

Set these environment variables in each **Run Configuration**:

```
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=ecommerce
KAFKA_BROKERS=localhost:29092
JWT_SECRET=change-me-in-production
USER_SERVICE_ADDR=localhost:50051
PRODUCT_SERVICE_ADDR=localhost:50052
ORDER_SERVICE_ADDR=localhost:50053
PAYMENT_SERVICE_ADDR=localhost:50054
```

Start infrastructure only first:
```bash
docker compose up postgres kafka zookeeper -d
```

Then run each service from GoLand using its `main.go`.

---

## API Reference

All protected routes require `Authorization: Bearer <token>`.

### Auth

```bash
# Register
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","email":"alice@example.com","password":"secret123"}'

# Login → returns token
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"alice@example.com","password":"secret123"}'
```

### Products

```bash
# Create product
curl -X POST http://localhost:8080/api/v1/products \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"MacBook Pro","description":"M3 chip","price":2499.99,"stock":50}'

# List products
curl http://localhost:8080/api/v1/products?page=1&limit=10 \
  -H "Authorization: Bearer $TOKEN"

# Get product by ID
curl http://localhost:8080/api/v1/products/<product_id> \
  -H "Authorization: Bearer $TOKEN"
```

### Orders

```bash
# Place order (triggers Kafka → async payment)
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"product_id":"<product_id>","quantity":2}'

# List my orders
curl http://localhost:8080/api/v1/orders \
  -H "Authorization: Bearer $TOKEN"
```

### Payments

```bash
# Get payment by ID (returned from order or from DB)
curl http://localhost:8080/api/v1/payments/<payment_id> \
  -H "Authorization: Bearer $TOKEN"
```

---

## How the async flow works

```
POST /orders
  │
  ├─► order-service (gRPC)
  │     ├─ validates product (gRPC → product-service)
  │     ├─ deducts stock    (gRPC → product-service)
  │     ├─ persists order   (PostgreSQL)
  │     └─ publishes ──────────────────────────────┐
  │                                                 │ Kafka: order.created
  │                                            ┌────▼───────────────────┐
  │                                            │    payment-service     │
  │                                            │  (Kafka consumer)      │
  │                                            │  ├─ calls mock gateway │
  │                                            │  ├─ persists payment   │
  │                                            │  └─ publishes ─────────┤
  │                                            └────────────────────────┘
  │                                       Kafka: payment.processed / failed
  │                                            ┌────▼───────────────────┐
  │                                            │ notification-service   │
  │                                            │  └─ logs email/SMS     │
  │                                            └────────────────────────┘
  │
  └─► 201 Created (order ID returned immediately — payment is async)
```

---

## Project Structure

```
ecommerce-microservices/
├── docker-compose.yml          # Full system orchestration
├── Makefile                    # Dev workflow shortcuts
├── go.mod                      # Single Go module for all services
│
├── proto/                      # gRPC service contracts (hand-written; no protoc needed)
│   ├── user/                   # types.go + service.go (client, server, ServiceDesc)
│   ├── product/
│   ├── order/
│   └── payment/
│
├── shared/                     # Cross-cutting packages
│   ├── codec/                  # JSON gRPC codec (replaces protobuf wire format)
│   ├── config/                 # Env-based config loader
│   ├── events/                 # Kafka event payload types
│   ├── kafka/                  # Producer & consumer helpers
│   └── middleware/             # JWT auth middleware + token generator
│
└── services/
    ├── api-gateway/            # HTTP → gRPC translation layer
    ├── user-service/           # Register, login, JWT
    ├── product-service/        # Catalog + atomic stock
    ├── order-service/          # Order orchestration + Kafka publisher
    ├── payment-service/        # Kafka consumer + mock gateway + gRPC
    └── notification-service/   # Kafka consumer → email/SMS simulation
```

---

## Key design decisions worth explaining in interviews

| Decision | Why |
|---|---|
| **gRPC for synchronous calls** | Strongly typed contracts, low latency, bi-directional streaming support |
| **Kafka for async events** | Decouples order creation from payment — order returns instantly; payment happens async |
| **JSON gRPC codec** | Eliminates protoc toolchain requirement; swap for protobuf in production |
| **Atomic stock deduction** | `SELECT FOR UPDATE` inside a DB transaction prevents overselling |
| **Idempotent payments** | Checks for existing payment before processing — safe to retry |
| **Outbox pattern note** | If Kafka publish fails after DB write, order is persisted but payment won't trigger. Production fix: write event to DB in same transaction, publish via outbox worker |
| **JWT in gateway only** | Services trust the gateway; internal gRPC calls are unauthenticated (mTLS in production) |

---

## Makefile Commands

| Command | Description |
|---|---|
| `make up` | Build and start all services |
| `make down` | Stop all services |
| `make logs` | Stream all logs |
| `make logs-svc SVC=order-service` | Stream logs for one service |
| `make restart SVC=payment-service` | Restart one service |
| `make seed` | Seed sample user, product, and order |
| `make clean` | Remove containers, volumes, and images |
| `make tidy` | `go mod tidy` |
| `make test` | Run all unit tests |
