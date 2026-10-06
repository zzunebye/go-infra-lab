# Go URL Shortener

A small URL Shortener built with Go's standard library. Milestone 1 keeps the application local and uses an in-memory repository so the HTTP, service, and storage boundaries are clear before adding infrastructure.

## Requirements

- Go 1.22 or later

## Run

```bash
go run ./cmd/api
```

The server listens on `:8080`. Set `PORT` to change the listening port and `BASE_URL` to change the short URL returned by the create endpoint.

```bash
PORT=9090 BASE_URL=http://localhost:9090 go run ./cmd/api
```

## API

Create a link:

```bash
curl -X POST http://localhost:8080/links \
  -H 'Content-Type: application/json' \
  -d '{"url":"https://example.com/some/very/long/path"}'
```

The response contains `code`, `shortUrl`, and `originalUrl`. Codes are seven-character Base62 values generated with `crypto/rand`.

Retrieve a link:

```bash
curl http://localhost:8080/links/aB3kP9x
```

Follow a short link:

```bash
curl -i http://localhost:8080/aB3kP9x
```

The redirect returns `302 Found` with the original URL in the `Location` header. Invalid URLs return `400`, unknown codes return `404`, and API errors use a JSON `error` field.

## Structure

```text
cmd/api/             application wiring and HTTP server
internal/http/       HTTP handlers and routes
internal/link/       Link model, repository contract, and service
internal/storage/    concurrent in-memory repository
```

The memory repository is process-local and clears on restart. PostgreSQL and cloud infrastructure are planned for later milestones.

## Verify

```bash
go test ./...
go test -race ./...
```
