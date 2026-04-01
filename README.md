# Go Fast Start Template

This project is a fast startup backend template for people learning Go.
It gives beginners a clean structure, real API flow, Docker support, and PostgreSQL wiring so you can learn by building, not by setting everything up from zero.

## What this template offers now

- Router setup with `chi`
- Basic middleware (request id, logging, recovery, timeout)
- PostgreSQL connection pool with `pgx`
- Health and readiness handlers (`/healthz`, `/readyz`)
- User CRUD module with layered architecture (handler/service/repository/model)
- Startup seed for default admin user (`admin@email.com`)
- Dockerfile + Docker Compose setup
- Environment-based configuration (`.env`)

## What we are planning to add next

- Strong authentication flow (login, token handling, secure password strategy)
- Better validation and standardized API error responses
- Database migrations and seed examples
- More test examples for beginners

These planned features are not fully implemented yet, but this repo is structured so they can be added step by step.

## Beginner quick start

### 1. Prepare environment file

```bash
cp .env.example .env
```

This creates your local config file from the template.

### 2. Install and clean dependencies

```bash
go mod tidy
```

What `go mod tidy` does:

- Adds missing modules required by your imports
- Removes modules you do not use anymore
- Updates `go.sum` checksums so builds are reproducible

### 3. Start PostgreSQL with Docker

```bash
docker compose up -d db redis
```

This runs only the Postgres service in background mode.

### 4. Run the Go application

```bash
go run ./cmd/api
```

By default, the server runs on `http://localhost:8080`.

At startup, the app also seeds a default admin user (idempotent):

- Name from `SEED_ADMIN_NAME` (default: `Admin`)
- Email from `SEED_ADMIN_EMAIL` (default: `admin@email.com`)

If the email already exists, the seed updates the name and keeps the same user record.

### 5. Database Migrations

The project uses `golang-migrate` for versioned schema management.

- **Startup Sync:** The application calls `postgres.EnsureSchema()` on startup. It automatically scans the `migrations/` folder and applies any pending `.up.sql` files.
- **Manual Control (CLI):**
  - Install: `brew install golang-migrate`
  - Create new migration: `migrate create -ext sql -dir migrations -seq <name>`
- **Check Status:** Query the `schema_migrations` table in your database to see which version you're on.

### 6. Verify everything is working

```bash
curl localhost:8080/healthz
curl localhost:8080/readyz
curl localhost:8080/api/v1/hello
```

### 6. Try User CRUD endpoints

Create a user:

```bash
curl -X POST localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice","email":"alice@example.com"}'
```

List users:

```bash
curl localhost:8080/api/v1/users
```

You should see at least the seeded admin user.

Get user by id:

```bash
curl localhost:8080/api/v1/users/1
```

Update user:

```bash
curl -X PUT localhost:8080/api/v1/users/1 \
  -H "Content-Type: application/json" \
  -d '{"name":"Alice Updated","email":"alice.updated@example.com"}'
```

Delete user:

```bash
curl -X DELETE localhost:8080/api/v1/users/1
```

## Run everything with Docker

```bash
cp .env.example .env
docker compose up --build
```

This starts both app + database together.

If you changed Go code, restart/rebuild the app container:

```bash
docker compose up -d --build app
```

## Useful commands for beginners

```bash
go test ./...              # run all tests
docker compose logs -f app # view app logs
docker compose down        # stop services
docker compose down -v     # stop + remove database volume
```

## Enterprise Standard Project Structure

This template uses a simplified structure for beginners, but it's designed to seamlessly scale into the **Go Standard Layout** used by enterprise teams. 

Here is how a fully-fledged Go enterprise application is typically structured, and where you should place your DTOs, API endpoints, and third-party libraries (like Redis):

```text
.
├── cmd/
│   ├── api/               # The main entrypoint for your HTTP server (moves main.go here)
│   └── worker/            # Entrypoint for background sync workers or CRONs
│
├── internal/              # Private application code (cannot be imported by other repos)
│   │
│   ├── config/            # Env config + DB credentials builder
│   ├── server/            # Router + middleware (The API endpoints mapping)
│   │
│   ├── user/              # A Domain Module (Feature-based grouping)
│   │   ├── handler.go     # HTTP handlers executing logic for endpoints
│   │   ├── service.go     # Core Business logic (Interfaces & Implementations)
│   │   ├── repository.go  # Database queries specific to the User domain
│   │   ├── model.go       # Core DB Models / Entities
│   │   └── dto.go         # Data Transfer Objects (Requests/Responses shape)
│   │
│   ├── infrastructure/    # Third-party implementations and external resources
│   │   ├── cache/         # Redis connection and shared caching logic
│   │   ├── storage/       # AWS S3 / Google Cloud Storage wrappers
│   │   ├── payment/       # Stripe / PayPal clients
│   │   └── postgres/      # Global DB pool / Schema bootstrap
│   │
│   └── pkg/               # Shared, private project-wide helpers
│       └── httpx/         # Shared HTTP helpers (JSON parse/respond)
│
├── pkg/                   # PUBLIC libraries you author that other repos CAN import (optional)
├── api/                   # OpenAPI/Swagger specs, Protocol Buffers definitions
├── Dockerfile
└── docker-compose.yml
```

### Where to store specific components?

1. **DTOs (Data Transfer Objects)**:
   - Place them close to the handlers that use them. Usually, inside the domain module folder as `dto.go` (e.g., `internal/user/dto.go`), or define request/response types at the top of the `handler.go` file.
   - For highly complex APIs, teams sometimes create an `api/types/` package, but domain-driven structures prefer keeping them tightly scoped.
2. **API Endpoints**: 
   - The actual route mapping (e.g., `router.Get("/users", ...)` or `router.Post("/users", ...)`) lives centrally in `internal/server/router.go`. 
   - The execution code for that endpoint sits in the Module's Handler (e.g., `internal/user/handler.go`).
3. **Third-Party Libraries (Redis, Stripe, AWS)**: 
   - Store these in an `internal/infrastructure/` folder. For example, your Redis initiator goes in `internal/infrastructure/cache`, and your email functionality inside `internal/infrastructure/email`.
   - Your `service.go` should only depend on an *interface* (e.g., `type Cache interface`), and the concrete Redis implementation from `/infrastructure` is injected into the service. 
4. **`cmd/` Directory (The Enterprise Standard)**:
   - Instead of a single `main.go` at the root, enterprise apps move it to `cmd/api/main.go`. This allows one repository to host multiple executables flawlessly (such as `cmd/migration/main.go` or `cmd/cron/main.go`).

