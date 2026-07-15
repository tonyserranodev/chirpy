# Chirpy

<p align="center">
  <img src="assets/logo.png" alt="Chirpy logo" width="200">
</p>

Chirpy is a small, learning-focused Twitter-like micro-blogging API written in Go. It lets users sign up, post short messages ("chirps"), browse chirps, and optionally upgrade to a premium tier ("Chirpy Red").

## Why Chirpy?

This project is a self-contained Go backend for anyone who wants to see how a modern HTTP API fits together without heavy frameworks:

- **Plain Go standard library** routing (`net/http`)
- **JWT access tokens** and **refresh tokens** for authenticated sessions
- **Argon2id** password hashing
- **PostgreSQL** persistence via `sqlc` generated queries
- **Database migrations** with `goose`
- **Webhook endpoint** for premium upgrades

It’s a great reference for learning how to wire authentication, middleware, JSON responses, and a Postgres database into a single deployable service.

## Quick Start

### Prerequisites

- Go 1.25+
- PostgreSQL 14+ (local server or container)
- `goose` for migrations: `go install github.com/pressly/goose/v3/cmd/goose@latest`
- `sqlc` for code generation: `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest`
- A `.env` file with the following variables:

```env
DB_URL=postgres://user:password@localhost:5432/chirpy?sslmode=disable
JWT_SECRET=your-secret-key-here
API_KEY=your-polka-webhook-api-key
PLATFORM=dev
```

### Install and run

```bash
# Clone the repository
git clone https://github.com/tonyserranodev/chirpy.git
cd chirpy

# Install dependencies
go mod download

# Create the database (run in psql or similar)
CREATE DATABASE chirpy;

# Run migrations
just mu
# Or manually: cd sql/schema && goose postgres "$DB_URL" up

# Generate database code (optional, if you change SQL files)
sqlc generate

# Start the server
go run .
```

The server will listen on `http://localhost:8080`.

You can verify the API is up with:

```bash
curl http://localhost:8080/api/healthz
```

For a complete list of available endpoints and their request/response formats, see [`docs/endpoints.md`](docs/endpoints.md).

## Project Structure

```
.
├── main.go              # Entry point and route registration
├── users.go             # User registration, login, update, refresh/revoke handlers
├── chirps.go            # Chirp CRUD and listing handlers
├── admin.go             # Metrics, reset, readiness, and middleware
├── webhooks.go          # Polka webhook handler for Chirpy Red upgrades
├── json.go              # JSON response helpers
├── internal/auth        # Password hashing, JWT, and token extraction
├── internal/database    # sqlc-generated database queries
├── sql/schema           # Goose migration files
└── sql/queries          # sqlc query files
```

## License

This project is for educational purposes. Use it as a starting point for your own Go APIs.
