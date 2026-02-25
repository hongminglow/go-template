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
docker compose up -d db
```

This runs only the Postgres service in background mode.

### 4. Run the Go application

```bash
go run .
```

By default, the server runs on `http://localhost:8080`.

At startup, the app also seeds a default admin user (idempotent):
- Name from `SEED_ADMIN_NAME` (default: `Admin`)
- Email from `SEED_ADMIN_EMAIL` (default: `admin@email.com`)

If the email already exists, the seed updates the name and keeps the same user record.

### 5. Verify everything is working

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

## Project structure

```text
.
├── internal/
│   ├── config/      # env config + DSN builder
│   ├── database/    # postgres pool + schema bootstrap
│   ├── httpx/       # shared HTTP helpers (JSON request/response)
│   ├── server/      # router + middleware
│   ├── system/      # root/health/readiness handlers
│   └── user/        # user module (handler, service, repository, model)
├── main.go          # app bootstrap and graceful shutdown
├── Dockerfile
└── docker-compose.yml
```
