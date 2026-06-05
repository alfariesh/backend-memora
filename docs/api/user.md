# User Profile And Settings

Module ini mencakup profile user saat ini dan setting default reminder.

Base path semua endpoint di dokumen ini: `/v1`.

Semua endpoint di sini protected.

## User Object

Response user:

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "username": "andi",
  "email": "andi@example.com",
  "created_at": "2026-01-01T00:00:00Z",
  "updated_at": "2026-01-01T00:00:00Z"
}
```

Field:

| Field | Type | Notes |
| --- | --- | --- |
| `id` | string UUID | User ID. |
| `username` | string | Nama display/login saat register. |
| `email` | string | Email user. |
| `created_at` | string RFC3339 | UTC. |
| `updated_at` | string RFC3339 | UTC. |

Password hash tidak pernah dikirim ke FE.

## Get Profile

```http
GET /v1/user/profile
Authorization: Bearer <access_token>
```

Success `200`:

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "username": "andi",
  "email": "andi@example.com",
  "created_at": "2026-01-01T00:00:00Z",
  "updated_at": "2026-01-01T00:00:00Z"
}
```

Errors:

| Status | Body |
| --- | --- |
| `401` | Auth error. |
| `404` | `user_not_found` |
| `500` | `internal_server_error` |

## User Settings Object

Settings mengatur default saat membuat important day baru dan preference channel saat worker mengirim reminder.

```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "timezone": "Asia/Jakarta",
  "reminder_time": "09:00",
  "notification_channels": ["email", "in_app", "push"],
  "created_at": "2026-01-01T00:00:00Z",
  "updated_at": "2026-01-01T00:00:00Z"
}
```

Field:

| Field | Type | Notes |
| --- | --- | --- |
| `user_id` | string UUID | User pemilik settings. |
| `timezone` | string | IANA timezone valid. Default `Asia/Jakarta`. |
| `reminder_time` | string | Format `HH:mm`. Default `09:00`. |
| `notification_channels` | string array | Channel aktif: `email`, `in_app`, `push`. |
| `created_at` | string RFC3339 | UTC. |
| `updated_at` | string RFC3339 | UTC. |

## Get User Settings

```http
GET /v1/user/settings
Authorization: Bearer <access_token>
```

Success `200`:

```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "timezone": "Asia/Jakarta",
  "reminder_time": "09:00",
  "notification_channels": ["email", "in_app", "push"],
  "created_at": "2026-01-01T00:00:00Z",
  "updated_at": "2026-01-01T00:00:00Z"
}
```

Jika user belum pernah menyimpan settings, backend tetap mengembalikan default virtual dengan shape yang sama:

```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "timezone": "Asia/Jakarta",
  "reminder_time": "09:00",
  "notification_channels": ["email", "in_app", "push"],
  "created_at": "2026-01-01T00:00:00Z",
  "updated_at": "2026-01-01T00:00:00Z"
}
```

Errors:

| Status | Body |
| --- | --- |
| `401` | Auth error. |
| `500` | `internal_server_error` |

## Update User Settings

```http
PUT /v1/user/settings
Authorization: Bearer <access_token>
Content-Type: application/json
```

Request:

```json
{
  "timezone": "Asia/Makassar",
  "reminder_time": "08:30",
  "notification_channels": ["in_app", "push"]
}
```

Validation:

| Field | Required | Rule |
| --- | --- | --- |
| `timezone` | no | IANA timezone valid, max `64` chars. Empty string becomes default. |
| `reminder_time` | no | Format `HH:mm`. Empty string becomes default. |
| `notification_channels` | no | Each value one of `email`, `in_app`, `push`. |

Success `200`:

```json
{
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "timezone": "Asia/Makassar",
  "reminder_time": "08:30",
  "notification_channels": ["in_app", "push"],
  "created_at": "2026-01-01T00:00:00Z",
  "updated_at": "2026-01-01T00:00:00Z"
}
```

Important behavior:

- Jika `timezone` tidak dikirim atau `""`, backend memakai `Asia/Jakarta`.
- Jika `reminder_time` tidak dikirim atau `""`, backend memakai `09:00`.
- Jika `notification_channels` tidak dikirim atau `null`, backend memakai default `["email", "in_app", "push"]`.
- Jika `notification_channels` dikirim `[]`, semua channel reminder dinonaktifkan secara preference.
- Duplicate channel akan di-deduplicate dengan urutan pertama dipertahankan.
- Settings baru mempengaruhi default important day yang dibuat setelah settings berubah.
- Saat worker mengirim reminder, channels pada job tetap difilter lagi dengan `notification_channels` terbaru.

Errors:

| Status | Body |
| --- | --- |
| `400` | `validation_error`, `invalid_request_body`, atau `invalid_user_settings` |
| `401` | Auth error. |
| `500` | `internal_server_error` |

## FE Notes

- Untuk screen settings, validasi timezone dan format `HH:mm` di client sebelum submit.
- Jika user mematikan semua channel dengan `[]`, reminder job tetap bisa terschedule, tetapi worker tidak akan mengirim channel apa pun setelah difilter.
- Untuk default create important day, FE boleh omit `timezone` dan `reminder_time` agar backend memakai settings user.
