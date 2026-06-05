# Devices And Push

Module ini mengatur Expo push token milik user.

Base path semua endpoint di dokumen ini: `/v1`.

Semua endpoint di sini protected.

## Expo Token Flow For FE

1. App mobile meminta permission notification.
2. App mendapatkan Expo push token lewat Expo Notifications.
3. App mengirim token ke `POST /v1/devices/`.
4. Backend menyimpan token aktif untuk user.
5. Worker memakai token aktif saat mengirim reminder channel `push`.

Accepted token format:

```text
ExpoPushToken[...]
ExponentPushToken[...]
```

Backend hanya validasi envelope token. Validitas final tetap dari Expo saat push dikirim.

## Device Object

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "token": "ExpoPushToken[xxxxxxxxxxxxxxxxxxxxxx]",
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
| `token` | string | Expo push token. |
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
      "token": "ExpoPushToken[xxxxxxxxxxxxxxxxxxxxxx]",
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
| `500` | `{ "error": "internal server error" }` |

## Register Device

```http
POST /v1/devices/
Authorization: Bearer <access_token>
Content-Type: application/json
```

Request:

```json
{
  "token": "ExpoPushToken[xxxxxxxxxxxxxxxxxxxxxx]",
  "platform": "android",
  "name": "Pixel 8"
}
```

Validation:

| Field | Required | Rule |
| --- | --- | --- |
| `token` | yes | Must start with `ExpoPushToken[` or `ExponentPushToken[` and end with `]`. |
| `platform` | yes | max `40`. |
| `name` | no | max `255`. |

Success `201`:

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "token": "ExpoPushToken[xxxxxxxxxxxxxxxxxxxxxx]",
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
| `400` | `{ "error": "invalid request body" }` atau `{ "error": "invalid device token" }` |
| `401` | Auth error. |
| `500` | `{ "error": "internal server error" }` |

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
  "ticket_id": "expo-ticket-id",
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
| `400` | `{ "error": "invalid request body" }` |
| `401` | Auth error. |
| `404` | `{ "error": "device not found" }` |
| `410` | `{ "error": "push device not registered" }` |
| `502` | `{ "error": "push send failed" }` |
| `503` | `{ "error": "push sender not configured" }` |
| `500` | `{ "error": "internal server error" }` |

If Expo returns `DeviceNotRegistered`, backend deactivates token and returns `410`.

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
| `404` | `{ "error": "device not found" }` |
| `500` | `{ "error": "internal server error" }` |

## Expo Provider Notes

- Backend mengirim push ke `https://exp.host/--/api/v2/push/send`.
- Jika `EXPO_PUSH_ACCESS_TOKEN` diisi, backend mengirim header `Authorization: Bearer <token>`.
- Jika env itu kosong, backend tetap bisa mengirim request Expo tanpa Authorization.
- Firebase/APNs credential tidak dibutuhkan backend, tetapi tetap dibutuhkan di project Expo/EAS mobile agar device menerima push.

## FE Notes

- Simpan device ID dari response jika ingin menyediakan tombol "test push" atau "remove device".
- Jangan tampilkan raw push token kecuali untuk debug.
- Register ulang token saat user login, saat token berubah, atau saat app reinstall.
