# Common Contract

Dokumen ini berlaku untuk semua REST endpoint Memora.

## Base Path

```text
Base URL local: http://127.0.0.1:8080
Base path API: /v1
```

Contoh full URL:

```http
GET http://127.0.0.1:8080/v1/user/profile
```

## Headers

Untuk request JSON:

```http
Content-Type: application/json
Accept: application/json
```

Untuk protected endpoint:

```http
Authorization: Bearer <access_token>
```

Public endpoint:

- `POST /v1/auth/register`
- `POST /v1/auth/login`
- `POST /v1/auth/refresh`
- `POST /v1/auth/logout`
- `GET /healthz`

Semua endpoint lain protected.

## Auth Error Messages

Middleware auth mengembalikan `401` dengan salah satu message ini:

```json
{
  "error": "missing_authorization_header",
  "message": "missing authorization header"
}
```

```json
{
  "error": "invalid_authorization_header",
  "message": "invalid authorization header format"
}
```

```json
{
  "error": "invalid_or_expired_token",
  "message": "invalid or expired token"
}
```

Format header harus persis `Bearer <token>`. `bearer`, token tanpa prefix, atau header kosong ditolak.

## Error Shape

Semua error REST memakai shape:

```json
{
  "error": "machine_readable_code",
  "message": "human readable message",
  "fields": {
    "field_name": "field error message"
  }
}
```

Rules:

- `error` adalah machine-readable code dalam `snake_case`.
- `message` adalah pesan singkat untuk debugging atau fallback UI.
- `fields` hanya ada pada validation error.
- FE sebaiknya branch logic berdasarkan status code dan `error`.

Contoh non-validation error:

```json
{
  "error": "invalid_credentials",
  "message": "invalid credentials"
}
```

Contoh validation error:

```json
{
  "error": "validation_error",
  "message": "validation failed",
  "fields": {
    "email": "must be a valid email",
    "password": "is required"
  }
}
```

## Common Status Codes

| Code | Arti untuk FE |
| --- | --- |
| `200` | Request sukses dengan JSON body. |
| `201` | Resource berhasil dibuat. |
| `204` | Request sukses tanpa body. Jangan coba parse JSON. |
| `400` | Body/query invalid, field tidak valid, tanggal invalid, token Expo invalid. |
| `401` | Access token tidak ada, format salah, expired, atau invalid. |
| `403` | Aksi ditolak oleh authorization eksplisit. ID milik user lain diperlakukan sebagai `404`. |
| `404` | Resource tidak ditemukan, tidak aktif, atau bukan milik user ini. |
| `409` | Conflict, saat ini dipakai untuk duplicate user email. |
| `410` | Expo menyatakan device token tidak terdaftar lagi. |
| `429` | Terkena rate limit. |
| `502` | Backend gagal mengirim ke provider push. |
| `503` | Push sender tidak tersedia atau tidak terkonfigurasi pada runtime tertentu. |
| `500` | Error backend tidak terduga. |

## JSON And Nulls

- Field dengan pointer kosong dikirim sebagai `null`, contoh `event_year: null`, `anniversary: null`, `read_at: null`.
- Field string kosong tetap dikirim sebagai `""` jika memang tersimpan kosong.
- Array kosong dikirim sebagai `[]`.
- `204 No Content` tidak punya body.

## ID Format

ID resource memakai UUID string:

```json
"550e8400-e29b-41d4-a716-446655440000"
```

Jangan generate ID resource di FE. Backend yang membuat ID.

## Time Format

Timestamp response memakai RFC3339 UTC:

```json
"2026-01-01T00:00:00Z"
```

Field date-only untuk important day:

```json
{
  "event_year": 1970,
  "event_month": 5,
  "event_day": 13
}
```

Field time-only untuk reminder:

```json
"reminder_time": "09:00"
```

`reminder_time` harus format `HH:mm` 24 jam.

## Timezone

`timezone` harus IANA timezone valid:

```json
"Asia/Jakarta"
```

Contoh lain:

- `Asia/Makassar`
- `Asia/Jayapura`
- `UTC`

Timezone dipakai untuk menghitung occurrence date dan waktu reminder lokal user/event.

## Pagination

Endpoint list memakai `limit` dan `offset`.

Rules umum:

- `offset < 0` dianggap `0`.
- `limit <= 0` dianggap default.
- `limit > 100` dicap ke `100`.
- Query yang bukan angka dianggap default oleh endpoint terkait.

Default per endpoint:

| Endpoint | Default limit | Max limit | Default offset |
| --- | ---: | ---: | ---: |
| `GET /important-days/` | `10` | `100` | `0` |
| `GET /important-days/upcoming` | `10` | `100` | `0` |
| `GET /notifications/` | `20` | `100` | `0` |
| `GET /mobile/bootstrap` upcoming list | `5` | `100` | `0` |

List response selalu membawa `total`, yaitu jumlah semua item yang match filter sebelum pagination.

## Collection Route Slash

Dokumen memakai path collection dengan trailing slash karena route saat ini didaftarkan seperti itu:

```http
GET /v1/important-days/
POST /v1/devices/
GET /v1/notifications/
```

Fiber biasanya tidak strict terhadap slash, tetapi FE sebaiknya mengikuti path di dokumen agar konsisten dengan integration test.

## Validation Messages

Validation error memakai body:

```json
{
  "error": "validation_error",
  "message": "validation failed",
  "fields": {
    "username": "must be at least 3 characters",
    "email": "must be a valid email",
    "password": "is required"
  }
}
```

Field names mengikuti JSON field, termasuk nested array path:

```json
{
  "fields": {
    "reminder_rules[0].offset_days": "must be at least 0",
    "reminder_rules[0].channels[0]": "must be one of: email, in_app, push"
  }
}
```

JSON parse error tetap memakai error umum:

```json
{
  "error": "invalid_request_body",
  "message": "invalid request body"
}
```

FE tetap disarankan melakukan validasi client-side agar user mendapat feedback sebelum submit.
