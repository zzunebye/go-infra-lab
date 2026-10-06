# Go URL Shortener

A small URL Shortener built with Go. The API uses an in-memory repository when `DATABASE_URL` is unset, or PostgreSQL when it is configured. PostgreSQL schema migrations run automatically at startup.

## Requirements

- Go 1.26 or later
- Docker with the Compose plugin (for PostgreSQL mode)

## Run

```bash
go run ./cmd/api
```

The server listens on `:8080`. With no `DATABASE_URL`, it uses in-memory storage. Set `PORT` to change the listening port and `BASE_URL` to change the short URL returned by the create endpoint.

```bash
PORT=9090 BASE_URL=http://localhost:9090 go run ./cmd/api
```

## Run with PostgreSQL

Start the API and PostgreSQL together:

```bash
docker compose up --build
```

The API waits for PostgreSQL to become healthy, applies pending migrations, then listens on `http://localhost:18080`. PostgreSQL is available to host tools on port `5433`. PostgreSQL data persists in the `postgres_data` Docker volume across container restarts.

Stop the services with:

```bash
docker compose down
```

For local development without Docker, start PostgreSQL separately and set `DATABASE_URL` before running `go run ./cmd/api`. The migration runs automatically on startup.

## API

The commands below use the Compose API address, `http://localhost:18080`. When running with `go run ./cmd/api`, use `http://localhost:8080` instead.

Create a link:

```bash
curl -X POST http://localhost:18080/links \
  -H 'Content-Type: application/json' \
  -d '{"url":"https://example.com/some/very/long/path"}'
```

The response contains `code`, `shortUrl`, and `originalUrl`. Codes are seven-character Base62 values generated with `crypto/rand`.

Retrieve a link:

```bash
curl http://localhost:18080/links/aB3kP9x
```

Follow a short link:

```bash
curl -i http://localhost:18080/aB3kP9x
```

The redirect returns `302 Found` with the original URL in the `Location` header. Invalid URLs return `400`, unknown codes return `404`, and API errors use a JSON `error` field.

## Structure

```text
cmd/api/             application wiring and HTTP server
internal/http/       HTTP handlers and routes
internal/link/       Link model, repository contract, and service
internal/storage/    in-memory and PostgreSQL repositories, queries, and migrations
```

The memory repository is process-local and clears on restart. PostgreSQL mode keeps links across API restarts. Cloud infrastructure remains planned for later milestones.

## Verify

```bash
go test ./...
go test -race ./...
```
