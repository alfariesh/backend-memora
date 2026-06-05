# Auth

Module auth mengatur user registration, login, refresh token rotation, logout, dan access token untuk protected endpoint.

Base path semua endpoint di dokumen ini: `/v1`.

## Token Model

Backend mengeluarkan dua token:

| Field | Fungsi |
| --- | --- |
| `access_token` | JWT untuk `Authorization: Bearer <access_token>`. |
| `token` | Alias dari `access_token` untuk kompatibilitas client lama. |
| `refresh_token` | Opaque token untuk meminta access token baru. |
| `expires_at` | Expiry access token dalam RFC3339 UTC. |

Default expiry dari env:

| Env | Default |
| --- | --- |
| `JWT_TOKEN_EXPIRY` | `15m` |
| `JWT_REFRESH_TOKEN_EXPIRY` | `720h` |

FE baru sebaiknya memakai `access_token`, bukan `token`.

## Recommended FE Flow

1. User login.
2. Simpan `access_token`, `refresh_token`, dan `expires_at` di secure storage.
3. Untuk protected request, kirim `Authorization: Bearer <access_token>`.
4. Saat access token expired atau menerima `401 invalid or expired token`, panggil refresh.
5. Jika refresh sukses, ganti access token dan refresh token secara atomik.
6. Jika refresh gagal `401`, hapus local session dan arahkan user ke login.

Jangan memakai refresh token lama setelah refresh sukses. Backend revoke refresh token lama dan mengeluarkan token baru. FE wajib serialize refresh request; double-submit refresh token lama akan dianggap reuse dan semua session aktif user akan di-revoke.

## Register

```http
POST /v1/auth/register
Content-Type: application/json
```

Request:

```json
{
  "username": "andi",
  "email": "andi@example.com",
  "password": "secret123"
}
```

Validation:

| Field | Required | Rule |
| --- | --- | --- |
| `username` | yes | min `3`, max `255` chars |
| `email` | yes | valid email |
| `password` | yes | min `8`, max `72` chars |

Success `201`:

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "username": "andi",
  "email": "andi@example.com",
  "created_at": "2026-01-01T00:00:00Z",
  "updated_at": "2026-01-01T00:00:00Z"
}
```

Register tidak otomatis mengembalikan token. Setelah register, FE perlu login.

Errors:

| Status | Body |
| --- | --- |
| `400` | `validation_error` atau `invalid_request_body` |
| `429` | `too_many_requests` |
| `409` | `user_already_exists` |
| `500` | `internal_server_error` |

## Login

```http
POST /v1/auth/login
Content-Type: application/json
```

Request:

```json
{
  "email": "andi@example.com",
  "password": "secret123"
}
```

Validation:

| Field | Required | Rule |
| --- | --- | --- |
| `email` | yes | valid email |
| `password` | yes | non-empty, max `72` chars |

Success `200`:

```json
{
  "token": "<jwt>",
  "access_token": "<jwt>",
  "refresh_token": "<opaque-refresh-token>",
  "expires_at": "2026-01-01T00:00:00Z"
}
```

Errors:

| Status | Body |
| --- | --- |
| `400` | `validation_error` atau `invalid_request_body` |
| `401` | `invalid_credentials` |
| `429` | `too_many_requests` |
| `500` | `internal_server_error` |

## Refresh Token

```http
POST /v1/auth/refresh
Content-Type: application/json
```

Request:

```json
{
  "refresh_token": "<opaque-refresh-token>"
}
```

Success `200`:

```json
{
  "token": "<new-jwt>",
  "access_token": "<new-jwt>",
  "refresh_token": "<new-opaque-refresh-token>",
  "expires_at": "2026-01-01T00:00:00Z"
}
```

Behavior:

- Refresh token di-rotate.
- Refresh token lama langsung revoked setelah refresh sukses.
- Jika FE retry request refresh yang sama setelah sukses, retry itu akan menerima `401` dan semua session aktif user akan di-revoke sebagai reuse protection.
- Simpan token baru dulu, lalu retry request protected yang sebelumnya gagal.

Errors:

| Status | Body |
| --- | --- |
| `400` | `validation_error` atau `invalid_request_body` |
| `401` | `invalid_refresh_token` |
| `429` | `too_many_requests` |
| `500` | `internal_server_error` |

## Logout

```http
POST /v1/auth/logout
Content-Type: application/json
```

Request:

```json
{
  "refresh_token": "<opaque-refresh-token>"
}
```

Success:

```http
204 No Content
```

Behavior:

- Backend revoke refresh token.
- Access token yang sudah terbit tetap valid sampai JWT expiry.
- Logout dengan refresh token yang sudah invalid/revoked tetap diperlakukan sukses oleh usecase.
- Logout tidak memicu reuse detection.
- FE tetap harus hapus local token setelah logout sukses.

Errors:

| Status | Body |
| --- | --- |
| `400` | `validation_error` atau `invalid_request_body` |
| `500` | `internal_server_error` |

## List Sessions

```http
GET /v1/auth/sessions
Authorization: Bearer <access_token>
```

Success `200`:

```json
{
  "sessions": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "expires_at": "2026-01-31T00:00:00Z",
      "created_ip": "127.0.0.1",
      "created_user_agent": "Memora/1.0",
      "last_used_at": "2026-01-02T00:00:00Z",
      "last_used_ip": "127.0.0.1",
      "last_used_user_agent": "Memora/1.0",
      "created_at": "2026-01-01T00:00:00Z",
      "updated_at": "2026-01-02T00:00:00Z"
    }
  ],
  "total": 1
}
```

Only active, non-expired sessions are returned. Refresh token hashes are never returned.

Errors:

| Status | Body |
| --- | --- |
| `401` | Auth error. |
| `500` | `internal_server_error` |

## Revoke Session

```http
DELETE /v1/auth/sessions/{id}
Authorization: Bearer <access_token>
```

Success:

```http
204 No Content
```

Behavior:

- Only current user's sessions can be revoked.
- Missing or cross-user session IDs return `404`.
- Access token for that device remains valid until normal JWT expiry.

Errors:

| Status | Body |
| --- | --- |
| `401` | Auth error. |
| `404` | `session_not_found` |
| `500` | `internal_server_error` |

## Logout All

```http
POST /v1/auth/logout-all
Authorization: Bearer <access_token>
```

Success:

```http
204 No Content
```

Behavior:

- Revokes all active refresh sessions for current user.
- Current access token remains valid until normal JWT expiry.
- FE should clear local tokens after success.

Errors:

| Status | Body |
| --- | --- |
| `401` | Auth error. |
| `500` | `internal_server_error` |

## Auth Header For Protected Calls

```http
Authorization: Bearer <access_token>
```

Contoh:

```http
GET /v1/user/profile
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

Common auth errors dijelaskan di [Common Contract](common.md#auth-error-messages).
