# Reminder Rules

Reminder rules menentukan kapan dan lewat channel apa user diingatkan untuk sebuah important day.

Base path semua endpoint di dokumen ini: `/v1`.

Semua endpoint di sini protected.

## Reminder Channel Enum

Valid `channels`:

```text
email
in_app
push
```

Arti channel:

| Channel | Delivery |
| --- | --- |
| `email` | Email ke email account user melalui Cloudflare Email Service. |
| `in_app` | Membuat record notification yang muncul di endpoint notifications. |
| `push` | Mengirim Expo push notification ke active device tokens user. |

## Reminder Rule Object

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "important_day_id": "550e8400-e29b-41d4-a716-446655440000",
  "offset_days": 7,
  "channels": ["email", "in_app", "push"],
  "created_at": "2026-01-01T00:00:00Z",
  "updated_at": "2026-01-01T00:00:00Z"
}
```

Field:

| Field | Type | Notes |
| --- | --- | --- |
| `id` | string UUID | Reminder rule ID. |
| `user_id` | string UUID | Owner. |
| `important_day_id` | string UUID | Parent important day. |
| `offset_days` | integer | `0` means on event day, `7` means 7 days before. Min `0`. |
| `channels` | string array | Valid values: `email`, `in_app`, `push`. |
| `created_at` | string RFC3339 | UTC. |
| `updated_at` | string RFC3339 | UTC. |

## Default Reminder Rules

Jika `reminder_rules` tidak dikirim saat create important day, atau `rules` kosong saat replace, backend membuat default:

```json
[
  { "offset_days": 7, "channels": ["email", "in_app", "push"] },
  { "offset_days": 1, "channels": ["email", "in_app", "push"] },
  { "offset_days": 0, "channels": ["email", "in_app", "push"] }
]
```

Channels default sebenarnya mengikuti `notification_channels` user settings. Jika user settings adalah `["in_app", "push"]`, rules default juga memakai channels itu.

Important:

- Mengirim array kosong tidak berarti "hapus semua reminder". Array kosong berarti backend memakai default rules.
- Jika user ingin mematikan delivery, update `notification_channels` user settings ke `[]`.
- Jika `channels` dalam satu rule dikirim kosong atau omit, backend mengisi dengan default channels dari user settings.

## Scheduling Behavior

Untuk setiap rule, backend membuat pending reminder job:

```text
scheduled_at = occurrence_date at reminder_time in timezone - offset_days
```

Contoh:

```json
{
  "event_month": 5,
  "event_day": 13,
  "timezone": "Asia/Jakarta",
  "reminder_time": "09:00",
  "offset_days": 7
}
```

Reminder dijadwalkan pada `6 Mei 09:00 Asia/Jakarta`, lalu disimpan sebagai UTC.

Jika hasil schedule sudah lewat saat create/update/replace, backend set `scheduled_at` ke waktu sekarang supaya job bisa segera diproses worker.

## Get Reminder Rules

```http
GET /v1/important-days/:id/reminders
Authorization: Bearer <access_token>
```

Success `200`:

```json
{
  "rules": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "user_id": "550e8400-e29b-41d4-a716-446655440000",
      "important_day_id": "550e8400-e29b-41d4-a716-446655440000",
      "offset_days": 7,
      "channels": ["email", "in_app", "push"],
      "created_at": "2026-01-01T00:00:00Z",
      "updated_at": "2026-01-01T00:00:00Z"
    }
  ]
}
```

Errors:

| Status | Body |
| --- | --- |
| `401` | Auth error. |
| `404` | `important_day_not_found` |
| `500` | `internal_server_error` |

## Replace Reminder Rules

```http
PUT /v1/important-days/:id/reminders
Authorization: Bearer <access_token>
Content-Type: application/json
```

Request:

```json
{
  "rules": [
    {
      "offset_days": 7,
      "channels": ["email", "in_app", "push"]
    },
    {
      "offset_days": 0,
      "channels": ["in_app", "push"]
    }
  ]
}
```

Validation:

| Field | Required | Rule |
| --- | --- | --- |
| `rules` | no | Array. Empty or omitted means default rules. |
| `rules[].offset_days` | no | Min `0`. Missing becomes `0`. |
| `rules[].channels` | no | Each value must be `email`, `in_app`, or `push`. Empty means default user channels. |

Success `200`:

```json
{
  "rules": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "user_id": "550e8400-e29b-41d4-a716-446655440000",
      "important_day_id": "550e8400-e29b-41d4-a716-446655440000",
      "offset_days": 7,
      "channels": ["email", "in_app", "push"],
      "created_at": "2026-01-01T00:00:00Z",
      "updated_at": "2026-01-01T00:00:00Z"
    }
  ]
}
```

Behavior:

- Semua existing rules untuk important day itu diganti.
- Pending reminder jobs ikut dibangun ulang.
- Sent/failed job history yang sudah lewat tidak diekspos ke FE lewat REST.

Errors:

| Status | Body |
| --- | --- |
| `400` | `validation_error` atau `invalid_request_body` |
| `401` | Auth error. |
| `404` | `important_day_not_found` |
| `500` | `internal_server_error` |

## Reminder Copy

Worker membuat copy otomatis:

Jika `offset_days = 0`:

```text
Title: <important day title> is today
Body:  <important day title> is today.
```

Jika `offset_days > 0`:

```text
Title: <important day title> is in <offset_days> days
Body:  <important day title> is coming in <offset_days> days.
```

## Delivery Filtering

Saat worker memproses job, channels pada job difilter dengan `notification_channels` user settings terbaru.

Contoh:

```json
{
  "job_channels": ["email", "in_app", "push"],
  "user_notification_channels": ["in_app"]
}
```

Yang dikirim hanya `in_app`.

## FE Notes

- Untuk UI checkbox channel, pakai enum `email`, `in_app`, `push`.
- Untuk same-day reminder, kirim `offset_days: 0`.
- Jangan kirim negative offset. Backend akan reject via validation.
- Untuk "disable all notifications", update settings ke `notification_channels: []`, bukan replace rules dengan `[]`.
