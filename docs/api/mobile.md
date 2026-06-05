# Mobile Bootstrap

Endpoint bootstrap menggabungkan data kecil yang dibutuhkan mobile app setelah login atau saat resume.

Base path endpoint di dokumen ini: `/v1`.

Endpoint ini protected.

## Get Mobile Bootstrap

```http
GET /v1/mobile/bootstrap?upcoming_days=30&upcoming_limit=5&upcoming_offset=0
Authorization: Bearer <access_token>
```

Query:

| Query | Required | Default | Rule |
| --- | --- | --- | --- |
| `upcoming_days` | no | `30` | Lookahead upcoming important days. Invalid atau `<=0` menjadi `30`, max `3660`. |
| `upcoming_limit` | no | `5` | Invalid atau `<=0` menjadi `5`, max `100`. |
| `upcoming_offset` | no | `0` | Invalid atau `<0` menjadi `0`. |

Success `200`:

```json
{
  "settings": {
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "timezone": "Asia/Jakarta",
    "reminder_time": "09:00",
    "notification_channels": ["email", "in_app", "push"],
    "created_at": "2026-01-01T00:00:00Z",
    "updated_at": "2026-01-01T00:00:00Z"
  },
  "upcoming_important_days": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "user_id": "550e8400-e29b-41d4-a716-446655440000",
      "title": "Mom birthday",
      "type": "birthday",
      "person_name": "Mom",
      "relationship": "mother",
      "description": "Buy flowers",
      "event_year": 1970,
      "event_month": 5,
      "event_day": 13,
      "recurrence": "yearly",
      "timezone": "Asia/Jakarta",
      "reminder_time": "09:00",
      "created_at": "2026-01-01T00:00:00Z",
      "updated_at": "2026-01-01T00:00:00Z",
      "occurrence_date": "2026-05-13",
      "days_until": 7,
      "anniversary": 56
    }
  ],
  "upcoming_total": 1,
  "unread_notification_count": 3,
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
  "devices_total": 1
}
```

Response sections:

| Field | Source | Notes |
| --- | --- | --- |
| `settings` | User settings | Defaults returned if user has not saved settings. |
| `upcoming_important_days` | Upcoming important days | Same object as `GET /important-days/upcoming`. |
| `upcoming_total` | Upcoming total | Total before pagination. |
| `unread_notification_count` | Notifications | Use for badge. |
| `devices` | Active device tokens | Same object as `GET /devices/`. |
| `devices_total` | Device total | Count active devices returned. |

Errors:

| Status | Body |
| --- | --- |
| `401` | Auth error. |
| `500` | `internal_server_error` |

## Recommended FE Usage

Call bootstrap:

- After login/refresh success.
- On app cold start when token exists.
- On foreground resume if local cache may be stale.
- After large mutation if simpler than manually updating multiple caches.

Do not use bootstrap for infinite scrolling. For full lists:

- Important days: `GET /v1/important-days/`
- Upcoming: `GET /v1/important-days/upcoming`
- Notifications: `GET /v1/notifications/`
- Devices: `GET /v1/devices/`

## Cache Notes

- `settings` can update from settings screen.
- `unread_notification_count` changes after mark read/read all and after worker creates in-app notification.
- `devices` changes after register/delete/test push that deactivates token.
- `upcoming_important_days` changes after create/update/delete important day or replace reminder rules indirectly through schedule, but upcoming itself is based on important day date.
