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
{ "error": "missing authorization header" }
```

```json
{ "error": "invalid authorization header format" }
```

```json
{ "error": "invalid or expired token" }
```

Format header harus persis `Bearer <token>`. `bearer`, token tanpa prefix, atau header kosong ditolak.

## Error Shape

Semua error REST memakai shape:

```json
{
  "error": "message"
}
```

FE sebaiknya render berdasarkan status code dan `error`, bukan berdasarkan struktur lain.

## Common Status Codes

| Code | Arti untuk FE |
| --- | --- |
| `200` | Request sukses dengan JSON body. |
| `201` | Resource berhasil dibuat. |
| `204` | Request sukses tanpa body. Jangan coba parse JSON. |
| `400` | Body/query invalid, field tidak valid, tanggal invalid, token Expo invalid. |
| `401` | Access token tidak ada, format salah, expired, atau invalid. |
| `403` | Resource ada tetapi milik user lain. |
| `404` | Resource tidak ditemukan atau tidak aktif untuk user ini. |
| `409` | Conflict, saat ini dipakai untuk duplicate user email. |
| `410` | Expo menyatakan device token tidak terdaftar lagi. |
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

Backend saat ini mengembalikan message generic untuk banyak validation error, misalnya:

```json
{ "error": "invalid request body" }
```

Artinya FE perlu melakukan validasi client-side sendiri supaya user mendapat pesan field-level yang lebih jelas.
