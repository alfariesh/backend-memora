# Important Days

Important day adalah tanggal tahunan yang ingin diingat user: birthday, wedding, document renewal, medical reminder, dan sejenisnya.

Base path semua endpoint di dokumen ini: `/v1`.

Semua endpoint di sini protected.

## Domain Rules

- Recurrence saat ini selalu `yearly`.
- `event_month` dan `event_day` wajib.
- `event_year` optional. Jika tidak ada, anniversary di upcoming response bernilai `null`.
- `timezone` menentukan tanggal lokal occurrence.
- `reminder_time` menentukan jam lokal untuk reminder.
- Tanggal `29 Februari` valid. Di tahun non-kabisat occurrence jatuh ke `28 Februari`.

## Important Day Type Enum

Valid `type`:

```text
birthday
wedding
memorial
graduation
first_day
document
subscription
medical
custom
```

Jika create/update mengirim `type: ""` atau omit, backend memakai `custom`.

## Important Day Object

```json
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
  "updated_at": "2026-01-01T00:00:00Z"
}
```

Field:

| Field | Type | Notes |
| --- | --- | --- |
| `id` | string UUID | Important day ID. |
| `user_id` | string UUID | Owner. FE biasanya tidak perlu menampilkan ini. |
| `title` | string | Required, max `255`. |
| `type` | string enum | Default `custom`. |
| `person_name` | string | Optional, max `255`. |
| `relationship` | string | Optional, max `100`. |
| `description` | string | Optional, max `1000`. |
| `event_year` | integer or null | Optional, min `1`. |
| `event_month` | integer | Required, `1..12`. |
| `event_day` | integer | Required, valid day for month. |
| `recurrence` | string | Saat ini selalu `yearly`. |
| `timezone` | string | IANA timezone. |
| `reminder_time` | string | `HH:mm`. |
| `created_at` | string RFC3339 | UTC. |
| `updated_at` | string RFC3339 | UTC. |

## Create Important Day

```http
POST /v1/important-days/
Authorization: Bearer <access_token>
Content-Type: application/json
```

Request:

```json
{
  "title": "Mom birthday",
  "type": "birthday",
  "person_name": "Mom",
  "relationship": "mother",
  "description": "Buy flowers",
  "event_year": 1970,
  "event_month": 5,
  "event_day": 13,
  "timezone": "Asia/Jakarta",
  "reminder_time": "09:00",
  "reminder_rules": [
    {
      "offset_days": 7,
      "channels": ["email", "in_app", "push"]
    },
    {
      "offset_days": 1,
      "channels": ["in_app", "push"]
    }
  ]
}
```

Request fields:

| Field | Required | Rule |
| --- | --- | --- |
| `title` | yes | max `255`. |
| `type` | no | enum, default `custom`. |
| `person_name` | no | max `255`. |
| `relationship` | no | max `100`. |
| `description` | no | max `1000`. |
| `event_year` | no | min `1`; `null` allowed. |
| `event_month` | yes | `1..12`. |
| `event_day` | yes | valid day for `event_month`; `29 Feb` allowed. |
| `timezone` | no | IANA timezone, max `64`; default dari user settings. |
| `reminder_time` | no | `HH:mm`; default dari user settings. |
| `reminder_rules` | no | Lihat [Reminder Rules](reminders.md). |

Defaults:

- `type` default `custom`.
- `timezone` default dari user settings, lalu `Asia/Jakarta`.
- `reminder_time` default dari user settings, lalu `09:00`.
- `reminder_rules` default offsets `7`, `1`, `0` dengan channels dari user settings.
- `recurrence` selalu `yearly`.

Success `201`:

```json
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
  "updated_at": "2026-01-01T00:00:00Z"
}
```

Errors:

| Status | Body |
| --- | --- |
| `400` | `{ "error": "invalid request body" }` atau `{ "error": "invalid important day date" }` |
| `401` | Auth error. |
| `500` | `{ "error": "internal server error" }` |

## List Important Days

```http
GET /v1/important-days/?limit=10&offset=0&type=birthday
Authorization: Bearer <access_token>
```

Query:

| Query | Required | Default | Rule |
| --- | --- | --- | --- |
| `type` | no | none | Filter by important day type enum. |
| `limit` | no | `10` | `<=0` becomes `10`, max `100`. |
| `offset` | no | `0` | `<0` becomes `0`. |

Ordering:

```text
created_at DESC
```

Success `200`:

```json
{
  "important_days": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "title": "Mom birthday",
      "type": "birthday",
      "event_month": 5,
      "event_day": 13,
      "recurrence": "yearly",
      "timezone": "Asia/Jakarta",
      "reminder_time": "09:00"
    }
  ],
  "total": 1
}
```

Errors:

| Status | Body |
| --- | --- |
| `400` | `{ "error": "invalid important day type" }` |
| `401` | Auth error. |
| `500` | `{ "error": "internal server error" }` |

## Upcoming Important Days

```http
GET /v1/important-days/upcoming?days=365&limit=10&offset=0
Authorization: Bearer <access_token>
```

Query:

| Query | Required | Default | Rule |
| --- | --- | --- | --- |
| `days` | no | `365` | Lookahead window. `<=0` becomes `365`, max `3660`. |
| `limit` | no | `10` | `<=0` becomes `10`, max `100`. |
| `offset` | no | `0` | `<0` becomes `0`. |

Ordering:

```text
occurrence_date ASC
```

Success `200`:

```json
{
  "important_days": [
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
  "total": 1
}
```

Additional fields:

| Field | Type | Notes |
| --- | --- | --- |
| `occurrence_date` | string `YYYY-MM-DD` | Next yearly occurrence in important day timezone. |
| `days_until` | integer | Days from current date in occurrence timezone. |
| `anniversary` | integer or null | `occurrence year - event_year`; null if `event_year` is null. |

Errors:

| Status | Body |
| --- | --- |
| `401` | Auth error. |
| `500` | `{ "error": "internal server error" }` |

## Get Important Day

```http
GET /v1/important-days/:id
Authorization: Bearer <access_token>
```

Success `200`: returns Important Day Object.

Errors:

| Status | Body |
| --- | --- |
| `401` | Auth error. |
| `403` | `{ "error": "forbidden" }` |
| `404` | `{ "error": "important day not found" }` |
| `500` | `{ "error": "internal server error" }` |

`403` dapat terjadi jika ID ada tetapi milik user lain.

## Update Important Day

```http
PUT /v1/important-days/:id
Authorization: Bearer <access_token>
Content-Type: application/json
```

Request body sama seperti create, tetapi `reminder_rules` tidak diterima di endpoint ini. Untuk mengganti reminder rules, pakai [Replace Reminder Rules](reminders.md#replace-reminder-rules).

Request:

```json
{
  "title": "Mom birthday",
  "type": "birthday",
  "person_name": "Mom",
  "relationship": "mother",
  "description": "Buy flowers",
  "event_year": 1970,
  "event_month": 5,
  "event_day": 13,
  "timezone": "Asia/Jakarta",
  "reminder_time": "09:00"
}
```

Success `200`: returns updated Important Day Object.

Behavior:

- Pending reminder jobs for this important day are rebuilt using existing rules.
- Existing rules are not changed by this endpoint.

Errors:

| Status | Body |
| --- | --- |
| `400` | `{ "error": "invalid request body" }` atau `{ "error": "invalid important day date" }` |
| `401` | Auth error. |
| `404` | `{ "error": "important day not found" }` |
| `500` | `{ "error": "internal server error" }` |

## Delete Important Day

```http
DELETE /v1/important-days/:id
Authorization: Bearer <access_token>
```

Success:

```http
204 No Content
```

Behavior:

- Important day dihapus.
- Reminder rules ikut terhapus via database cascade.
- Reminder jobs ikut terhapus via database cascade.
- Notifications lama tidak ikut terhapus, tetapi `important_day_id` di notification menjadi `null`.

Errors:

| Status | Body |
| --- | --- |
| `401` | Auth error. |
| `404` | `{ "error": "important day not found" }` |
| `500` | `{ "error": "internal server error" }` |

## FE Notes

- Untuk date picker, simpan month/day/year sebagai field terpisah, bukan timestamp.
- Untuk birthday tanpa tahun lahir, kirim `event_year: null` atau omit.
- Untuk tanggal `29 Feb`, FE boleh mengizinkan input itu. Backend akan handle non-leap year occurrence sebagai `28 Feb`.
- Gunakan `upcoming` untuk home list dan calendar preview, bukan menghitung sendiri dari list biasa.
