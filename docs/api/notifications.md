# Notifications

Module notifications berisi in-app notification yang dibuat oleh reminder worker.

Base path semua endpoint di dokumen ini: `/v1`.

Semua endpoint di sini protected.

## Notification Object

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "important_day_id": "550e8400-e29b-41d4-a716-446655440000",
  "type": "important_day_reminder",
  "title": "Mom birthday is in 7 days",
  "body": "Mom birthday is coming in 7 days.",
  "data": "{\"important_day_id\":\"550e8400-e29b-41d4-a716-446655440000\",\"reminder_job_id\":\"550e8400-e29b-41d4-a716-446655440000\",\"occurrence_date\":\"2026-05-13\"}",
  "read_at": null,
  "created_at": "2026-01-01T00:00:00Z"
}
```

Field:

| Field | Type | Notes |
| --- | --- | --- |
| `id` | string UUID | Notification ID. |
| `user_id` | string UUID | Owner. |
| `important_day_id` | string UUID or null | Parent important day jika notification berasal dari reminder. |
| `type` | string | Saat ini worker memakai `important_day_reminder`. |
| `title` | string | Copy untuk ditampilkan. |
| `body` | string | Copy detail. |
| `data` | string | JSON string, bukan object. FE parse manual jika perlu. |
| `read_at` | string RFC3339 or null | `null` berarti unread. |
| `created_at` | string RFC3339 | UTC. |

Data string untuk reminder biasanya berisi:

```json
{
  "important_day_id": "550e8400-e29b-41d4-a716-446655440000",
  "reminder_job_id": "550e8400-e29b-41d4-a716-446655440000",
  "occurrence_date": "2026-05-13"
}
```

## List Notifications

```http
GET /v1/notifications/?limit=20&offset=0&unread_only=false
Authorization: Bearer <access_token>
```

Query:

| Query | Required | Default | Rule |
| --- | --- | --- | --- |
| `unread_only` | no | `false` | Boolean query. |
| `limit` | no | `20` | `<=0` becomes `20`, max `100`. |
| `offset` | no | `0` | `<0` becomes `0`. |

Ordering:

```text
created_at DESC
```

Success `200`:

```json
{
  "notifications": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "user_id": "550e8400-e29b-41d4-a716-446655440000",
      "important_day_id": "550e8400-e29b-41d4-a716-446655440000",
      "type": "important_day_reminder",
      "title": "Mom birthday is in 7 days",
      "body": "Mom birthday is coming in 7 days.",
      "data": "{\"important_day_id\":\"550e8400-e29b-41d4-a716-446655440000\",\"reminder_job_id\":\"550e8400-e29b-41d4-a716-446655440000\",\"occurrence_date\":\"2026-05-13\"}",
      "read_at": null,
      "created_at": "2026-01-01T00:00:00Z"
    }
  ],
  "total": 1
}
```

Errors:

| Status | Body |
| --- | --- |
| `401` | Auth error. |
| `500` | `{ "error": "internal server error" }` |

## Unread Notification Count

```http
GET /v1/notifications/unread-count
Authorization: Bearer <access_token>
```

Success `200`:

```json
{
  "unread_count": 3
}
```

Use this for badge count. For initial app load, [Mobile Bootstrap](mobile.md) already includes `unread_notification_count`.

Errors:

| Status | Body |
| --- | --- |
| `401` | Auth error. |
| `500` | `{ "error": "internal server error" }` |

## Mark Notification Read

```http
PATCH /v1/notifications/:id/read
Authorization: Bearer <access_token>
```

Success `200`:

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "important_day_id": "550e8400-e29b-41d4-a716-446655440000",
  "type": "important_day_reminder",
  "title": "Mom birthday is in 7 days",
  "body": "Mom birthday is coming in 7 days.",
  "data": "{\"important_day_id\":\"550e8400-e29b-41d4-a716-446655440000\",\"reminder_job_id\":\"550e8400-e29b-41d4-a716-446655440000\",\"occurrence_date\":\"2026-05-13\"}",
  "read_at": "2026-01-01T00:00:00Z",
  "created_at": "2026-01-01T00:00:00Z"
}
```

Behavior:

- Backend set `read_at` ke waktu sekarang.
- Jika notification sudah read, endpoint tetap update `read_at`.

Errors:

| Status | Body |
| --- | --- |
| `401` | Auth error. |
| `404` | `{ "error": "notification not found" }` |
| `500` | `{ "error": "internal server error" }` |

## Mark All Notifications Read

```http
PATCH /v1/notifications/read-all
Authorization: Bearer <access_token>
```

Success:

```http
204 No Content
```

Behavior:

- Semua notification unread milik user diset read.
- Jika tidak ada unread notification, tetap `204`.

Errors:

| Status | Body |
| --- | --- |
| `401` | Auth error. |
| `500` | `{ "error": "internal server error" }` |

## FE Notes

- Treat `read_at == null` sebagai unread.
- Parse `data` defensively. Field ini string JSON dan bisa saja berubah untuk type notification lain di masa depan.
- Setelah mark read, update local list item dari response dan decrement badge jika sebelumnya unread.
- Setelah mark all read, set semua local notifications menjadi read atau refetch list/count.
