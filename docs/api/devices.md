# Devices And Push

Module ini mengatur OneSignal subscription ID milik user.

Base path semua endpoint di dokumen ini: `/v1`.

Semua endpoint di sini protected.

## OneSignal Token Flow For FE

1. App mobile meminta permission notification.
2. App mendapatkan OneSignal subscription ID dari OneSignal SDK.
3. App mengirim subscription ID ke `POST /v1/devices/`.
4. Backend menyimpan token aktif untuk user.
5. Worker memakai token aktif saat mengirim reminder channel `push`.

Accepted token format adalah UUID subscription ID:

```text
1dd608f2-c6a1-11e3-851d-000c2940e62c
```

Backend hanya validasi format UUID. Validitas final tetap dari OneSignal saat push dikirim.

## Device Object

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "token": "1dd608f2-c6a1-11e3-851d-000c2940e62c",
  "platform": "android",
  "name": "Pixel 8",
  "active": true,
  "created_at": "2026-01-01T00:00:00Z",
  "updated_at": "2026-01-01T00:00:00Z"
}
```

Field:

| Field | Type | Notes |
| --- | --- | --- |
| `id` | string UUID | Device token row ID. |
| `user_id` | string UUID | Owner. |
| `token` | string UUID | OneSignal subscription ID. |
| `platform` | string | Required, max `40`. FE bebas mengirim `ios`, `android`, `web`, dll. |
| `name` | string | Optional display name, max `255`. |
| `active` | boolean | Hanya active token dipakai dan ditampilkan. |
| `created_at` | string RFC3339 | UTC. |
| `updated_at` | string RFC3339 | UTC. |

## List Devices

```http
GET /v1/devices/
Authorization: Bearer <access_token>
```

Success `200`:

```json
{
  "devices": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "user_id": "550e8400-e29b-41d4-a716-446655440000",
      "token": "1dd608f2-c6a1-11e3-851d-000c2940e62c",
      "platform": "android",
      "name": "Pixel 8",
      "active": true,
      "created_at": "2026-01-01T00:00:00Z",
      "updated_at": "2026-01-01T00:00:00Z"
    }
  ],
  "total": 1
}
```

Behavior:

- Hanya active device token yang dikembalikan.
- Ordering `updated_at DESC`.
- Tidak ada pagination saat ini.

Errors:

| Status | Body |
| --- | --- |
| `401` | Auth error. |
| `500` | `internal_server_error` |

## Register Device

```http
POST /v1/devices/
Authorization: Bearer <access_token>
Content-Type: application/json
```

Request:

```json
{
  "token": "1dd608f2-c6a1-11e3-851d-000c2940e62c",
  "platform": "android",
  "name": "Pixel 8"
}
```

Validation:

| Field | Required | Rule |
| --- | --- | --- |
| `token` | yes | Must be a valid OneSignal subscription ID UUID. |
| `platform` | yes | max `40`. |
| `name` | no | max `255`. |

Success `201`:

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "token": "1dd608f2-c6a1-11e3-851d-000c2940e62c",
  "platform": "android",
  "name": "Pixel 8",
  "active": true,
  "created_at": "2026-01-01T00:00:00Z",
  "updated_at": "2026-01-01T00:00:00Z"
}
```

Upsert behavior:

- Unique key adalah `(user_id, token)`.
- Jika token yang sama dikirim lagi, backend update `platform`, `name`, set `active = true`, dan update `updated_at`.
- Ini berarti FE aman memanggil register setiap app start atau setelah token berubah.

Errors:

| Status | Body |
| --- | --- |
| `400` | `validation_error`, `invalid_request_body`, atau `invalid_device_token` |
| `401` | Auth error. |
| `500` | `internal_server_error` |

## Send Test Push

```http
POST /v1/devices/:id/test-push
Authorization: Bearer <access_token>
Content-Type: application/json
```

Body optional:

```json
{
  "title": "Memora test",
  "body": "Push notifications are working."
}
```

Validation:

| Field | Required | Rule |
| --- | --- | --- |
| `title` | no | max `100`; default `Memora test`. |
| `body` | no | max `255`; default `Push notifications are working.` |

Success `200`:

```json
{
  "device_id": "550e8400-e29b-41d4-a716-446655440000",
  "ticket_id": "00000000-0000-0000-0000-000000000000",
  "sent_at": "2026-01-01T00:00:00Z"
}
```

Use case:

- Panggil setelah register token untuk memverifikasi device bisa menerima push.
- Cocok untuk screen debug/settings.
- Jangan panggil otomatis terlalu sering karena ini benar-benar mengirim push.

Errors:

| Status | Body |
| --- | --- |
| `400` | `validation_error` atau `invalid_request_body` |
| `401` | Auth error. |
| `404` | `device_not_found` |
| `410` | `push_device_not_registered` |
| `502` | `push_send_failed` |
| `503` | `push_sender_not_configured` |
| `500` | `internal_server_error` |

If OneSignal returns the subscription in `invalid_player_ids`, backend deactivates token and returns `410`.

## Delete Device

```http
DELETE /v1/devices/:id
Authorization: Bearer <access_token>
```

Success:

```http
204 No Content
```

Behavior:

- Backend soft-deactivates device token.
- Deleted/deactivated token tidak muncul di list.
- Jika token yang sama didaftarkan lagi, backend reactivate token itu.

Errors:

| Status | Body |
| --- | --- |
| `401` | Auth error. |
| `404` | `device_not_found` |
| `500` | `internal_server_error` |

## OneSignal Provider Notes

- Backend mengirim push ke OneSignal Create Push Notification API.
- Field `token` pada register device adalah OneSignal `subscription_id`.
- Jika `ONESIGNAL_APP_ID` atau `ONESIGNAL_REST_API_KEY` kosong, push reminder/test push dianggap provider belum configured.
- Firebase/APNs credential tetap dikonfigurasi di project mobile/OneSignal agar device menerima push.

## FE Notes

- Simpan device ID dari response jika ingin menyediakan tombol "test push" atau "remove device".
- Jangan tampilkan raw push token kecuali untuk debug.
- Register ulang token saat user login, saat token berubah, atau saat app reinstall.
