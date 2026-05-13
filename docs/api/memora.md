# Memora Mobile API Contract

Base path: `/v1`

All protected endpoints require:

```http
Authorization: Bearer <jwt>
Content-Type: application/json
```

Error responses use:

```json
{
  "error": "message"
}
```

## Auth

### Register

`POST /auth/register`

```json
{
  "username": "andi",
  "email": "andi@example.com",
  "password": "secret123"
}
```

Success: `201`

### Login

`POST /auth/login`

```json
{
  "email": "andi@example.com",
  "password": "secret123"
}
```

Success: `200`

```json
{
  "token": "<jwt>"
}
```

### Profile

`GET /user/profile`

Success: `200`

## Important Days

Supported `type` values:

`birthday`, `wedding`, `memorial`, `graduation`, `first_day`, `document`, `subscription`, `medical`, `custom`

Supported reminder `channels`:

`email`, `in_app`, `push`

### Create Important Day

`POST /important-days/`

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

Success: `201`

Notes:

- `event_year` is optional. If present, backend returns anniversary count in upcoming responses.
- `timezone` defaults to `Asia/Jakarta`.
- `reminder_time` defaults to `09:00`.
- `reminder_rules` defaults to `7`, `1`, and `0` days before with `email`, `in_app`, and `push`.
- Yearly recurrence is currently the only recurrence mode.

### List Important Days

`GET /important-days/?limit=10&offset=0&type=birthday`

Success: `200`

```json
{
  "important_days": [],
  "total": 0
}
```

### Upcoming Important Days

`GET /important-days/upcoming?days=365&limit=10&offset=0`

Success: `200`

```json
{
  "important_days": [
    {
      "id": "uuid",
      "title": "Mom birthday",
      "type": "birthday",
      "event_month": 5,
      "event_day": 13,
      "occurrence_date": "2026-05-13",
      "days_until": 7,
      "anniversary": 56
    }
  ],
  "total": 1
}
```

### Get Important Day

`GET /important-days/:id`

Success: `200`

### Update Important Day

`PUT /important-days/:id`

Request body is the same as create, except `reminder_rules` is not accepted here.

Success: `200`

### Replace Reminder Rules

`PUT /important-days/:id/reminders`

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

Success: `200`

```json
{
  "rules": []
}
```

### Delete Important Day

`DELETE /important-days/:id`

Success: `204`

## Devices And Push

The Expo app should call `Notifications.getExpoPushTokenAsync(...)`, then register the returned token.

Accepted token formats:

- `ExpoPushToken[...]`
- `ExponentPushToken[...]`

Backend sends push notifications through Expo Push Service. Firebase/APNs credentials are not needed by the backend. They are still needed in the Expo/EAS mobile project so devices can receive push notifications.

### Register Device

`POST /devices/`

```json
{
  "token": "ExpoPushToken[xxxxxxxxxxxxxxxxxxxxxx]",
  "platform": "android",
  "name": "Pixel 8"
}
```

Success: `201`

If the same token is registered again, the backend updates platform/name and reactivates it.

Invalid token response: `400`

```json
{
  "error": "invalid device token"
}
```

### Delete Device

`DELETE /devices/:id`

Success: `204`

The backend soft-deactivates the token. If Expo later returns `DeviceNotRegistered`, the worker also deactivates that token automatically.

## Notifications

### List Notifications

`GET /notifications/?limit=10&offset=0&unread_only=true`

Success: `200`

```json
{
  "notifications": [],
  "total": 0
}
```

### Mark Notification Read

`PATCH /notifications/:id/read`

Success: `200`

### Mark All Notifications Read

`PATCH /notifications/read-all`

Success: `204`

## Common Status Codes

- `400`: invalid request body, invalid date, invalid type, invalid Expo push token
- `401`: missing or invalid JWT
- `403`: resource belongs to another user
- `404`: resource not found
- `409`: duplicate user email
- `500`: unexpected backend error

## Worker And Environment

Run the reminder worker with:

```bash
make run-worker
```

Relevant environment variables:

```env
RESEND_API_KEY=
RESEND_FROM_EMAIL=
EXPO_PUSH_ACCESS_TOKEN=
REMINDER_WORKER_BATCH_SIZE=50
REMINDER_WORKER_POLL_INTERVAL=1m
```

`EXPO_PUSH_ACCESS_TOKEN` is optional unless Expo push security is enabled.
