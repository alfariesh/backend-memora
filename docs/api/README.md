# Memora API Overview

Dokumentasi ini ditulis untuk frontend. Fokusnya REST API yang dipakai mobile/web client. gRPC, RabbitMQ RPC, dan NATS RPC tetap ada di backend, tetapi bukan kontrak utama untuk FE.

Base URL lokal:

```text
http://127.0.0.1:8080
```

Base path REST:

```text
/v1
```

Endpoint health check tidak memakai base path:

```http
GET /healthz
```

Swagger generated:

```text
/swagger
```

## Module Docs

- [Common Contract](common.md): header, auth, error shape, pagination, format tanggal/waktu, status code.
- [Auth](auth.md): register, login, refresh token, logout, token lifecycle.
- [User Profile And Settings](user.md): profile, reminder defaults, notification channel preferences.
- [Important Days](important-days.md): CRUD tanggal penting, enum type, upcoming calculation, date rules.
- [Reminder Rules](reminders.md): aturan reminder per important day, default offsets/channels, schedule behavior.
- [Devices And Push](devices.md): registrasi Expo token, test push, deactivate device.
- [Notifications](notifications.md): in-app notification list, unread badge, mark read.
- [Mobile Bootstrap](mobile.md): endpoint untuk hydrate home screen setelah login.
- [Reminder Worker](worker.md): perilaku delivery email/in-app/push dan environment terkait.

## Suggested FE Read Order

1. Baca [Common Contract](common.md) dulu.
2. Implement auth flow dari [Auth](auth.md).
3. Setelah login, panggil [Mobile Bootstrap](mobile.md).
4. Implement screen utama dari [Important Days](important-days.md), [Reminders](reminders.md), [Notifications](notifications.md), dan [Devices](devices.md).
5. Gunakan [User Settings](user.md) untuk preference timezone, reminder time, dan channel.

## High Level Flow

```text
Register or Login
  -> store access_token and refresh_token
  -> GET /v1/mobile/bootstrap
  -> render home screen, unread badge, upcoming days, device state
  -> register Expo push token when available
  -> CRUD important days and reminder rules
```

## Important FE Rules

- Semua protected endpoint wajib mengirim `Authorization: Bearer <access_token>`.
- `token` dan `access_token` di response auth berisi JWT yang sama. Pakai `access_token` untuk client baru.
- Refresh token di-rotate. Setelah `POST /auth/refresh` sukses, refresh token lama harus dibuang.
- Timestamp response memakai RFC3339 UTC, contoh `2026-01-01T00:00:00Z`.
- `event_month` dan `event_day` adalah date-only data, bukan timestamp.
- `reminder_time` memakai format 24 jam `HH:mm`, contoh `09:00`.
- `timezone` harus IANA timezone valid, contoh `Asia/Jakarta`.
- Error body selalu punya `error` machine code dan `message`; validation error menambahkan `fields`.
