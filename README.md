# Memora Backend

Backend API for Memora: authentication, important days, reminder rules, notifications, device tokens, and reminder delivery workers.

## Stack

- Go 1.26
- Fiber REST API
- gRPC
- RabbitMQ RPC
- NATS RPC
- PostgreSQL
- Cloudflare Email Service
- Expo push notifications

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

Email reminders use Cloudflare Email Service:

```env
CLOUDFLARE_EMAIL_ACCOUNT_ID=
CLOUDFLARE_EMAIL_API_TOKEN=
CLOUDFLARE_EMAIL_FROM_EMAIL=
```

Push reminders use Expo:

```env
EXPO_PUSH_ACCESS_TOKEN=
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
