# Memora Backend

Backend API for Memora: authentication, important days, reminder rules, notifications, device tokens, and reminder delivery workers.

## Stack

- Go 1.26
- Fiber REST API
- gRPC
- RabbitMQ RPC
- NATS RPC
- PostgreSQL
- Resend email delivery
- OneSignal push notifications

## Quick Start

```sh
cp .env.example .env
make compose-up
make run
```

Run the reminder worker in a separate process:

```sh
make run-worker
```

Run the full Docker integration test stack:

```sh
make compose-up-integration-test
```

## Services

- REST API: `http://127.0.0.1:8080`
- Health check: `http://127.0.0.1:8080/healthz`
- Swagger: `http://127.0.0.1:8080/swagger`
- gRPC: `127.0.0.1:8081`
- PostgreSQL: `postgres://user:myAwEsOm3pa55@w0rd@127.0.0.1:5432/db`
- RabbitMQ management: `http://127.0.0.1:15672`
- NATS monitoring: `http://127.0.0.1:8222`

## Configuration

Configuration is loaded from environment variables. Start from [.env.example](.env.example).

Browser clients are controlled by exact-origin CORS defaults:

```env
HTTP_CORS_ENABLED=true
HTTP_ALLOWED_ORIGINS=http://localhost:3000,http://127.0.0.1:3000,http://localhost:5173,http://127.0.0.1:5173
HTTP_ALLOWED_METHODS=GET,POST,PUT,PATCH,DELETE,OPTIONS
HTTP_ALLOWED_HEADERS=Authorization,Content-Type,Accept,X-Request-ID,X-Correlation-ID
HTTP_EXPOSED_HEADERS=X-Request-ID,X-Correlation-ID
HTTP_CORS_ALLOW_CREDENTIALS=false
```

Security headers are enabled by default. Keep HSTS disabled locally, and enable it only when the public API is served exclusively over HTTPS:

```env
HTTP_SECURITY_HEADERS_ENABLED=true
HTTP_SECURITY_HSTS_ENABLED=false
HTTP_SECURITY_HSTS_MAX_AGE=31536000
HTTP_SECURITY_HSTS_INCLUDE_SUBDOMAINS=true
HTTP_SECURITY_HSTS_PRELOAD=false
```

Email reminders use Resend:

```env
RESEND_API_KEY=
RESEND_FROM_EMAIL=
```

Push reminders use OneSignal:

```env
ONESIGNAL_APP_ID=
ONESIGNAL_REST_API_KEY=
```

## API Docs

- FE-facing REST API guide: [docs/api/README.md](docs/api/README.md)
- OpenAPI generated docs: [docs/swagger.yaml](docs/swagger.yaml)
- gRPC proto files: [docs/proto/v1](docs/proto/v1)

## Development Commands

```sh
make test
make integration-test
make linter-golangci
make mock
make swag-v1
make proto-v1
```
