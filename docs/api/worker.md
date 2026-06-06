# Reminder Worker

Dokumen ini menjelaskan perilaku worker yang memproses reminder jobs. Ini berguna untuk FE supaya behavior email, in-app notification, dan push tidak dianggap magic.

Worker bukan endpoint REST. Worker dijalankan sebagai process terpisah.

## Run Command

```bash
make run-worker
```

Docker image juga memiliki binary:

```bash
/worker
```

## Relevant Environment

```env
RESEND_API_KEY=
RESEND_FROM_EMAIL=
ONESIGNAL_APP_ID=
ONESIGNAL_REST_API_KEY=
REMINDER_WORKER_BATCH_SIZE=50
REMINDER_WORKER_POLL_INTERVAL=1m
```

Meaning:

| Env | Required | Notes |
| --- | --- | --- |
| `RESEND_API_KEY` | for email delivery | Resend API key. |
| `RESEND_FROM_EMAIL` | for email delivery | Sender email from verified Resend domain. |
| `ONESIGNAL_APP_ID` | for push delivery | OneSignal app ID. |
| `ONESIGNAL_REST_API_KEY` | for push delivery | OneSignal REST API key. |
| `REMINDER_WORKER_BATCH_SIZE` | no | Default `50`. |
| `REMINDER_WORKER_POLL_INTERVAL` | no | Default `1m`. |

Resend config kosong tidak membuat app crash. Email channel yang tidak configured akan di-skip.

OneSignal config kosong tidak membuat app crash. Push channel yang tidak configured akan di-skip.

## Job Lifecycle

1. Important day dibuat atau diupdate.
2. Reminder rules dibuat atau diganti.
3. Backend membuat pending reminder jobs berdasarkan occurrence berikutnya.
4. Worker claim due jobs sesuai batch size.
5. Worker filter channel job berdasarkan user settings terbaru.
6. Worker mencoba delivery untuk satu channel job itu.
7. Jika job `sent` atau `skipped`, worker finish job dan schedule occurrence tahun berikutnya dalam satu transaksi.
8. Jika job gagal transient, worker mark failed dan retry sampai attempt ketiga.

## Channel Filtering

Rule menyimpan `channels`, tetapi backend membuat satu pending job per channel. Sebelum delivery, worker tetap melihat user settings terbaru.

Contoh:

```json
{
  "job_channel": "email",
  "user_notification_channels": ["in_app", "push"]
}
```

Worker skip job `email` karena channel itu tidak aktif di user settings.

Jika `notification_channels` user settings adalah `[]`, tidak ada channel yang dikirim.

## Delivery Behavior

Worker memproses channel sesuai `reminder_jobs.channel`:

| Channel | Behavior |
| --- | --- |
| `email` | Kirim HTML email via Resend dengan deterministic idempotency key. Jika Resend config kosong, channel email di-skip tanpa retry. |
| `in_app` | Simpan notification ke database dengan `dedupe_key`. Ini yang muncul di `/v1/notifications/`. |
| `push` | Kirim OneSignal push ke semua active OneSignal subscription IDs user dengan per-device idempotency key. Jika tidak ada active device, channel di-skip tanpa retry. |

Job dianggap `sent` jika:

- Channel provider berhasil, atau
- In-app notification berhasil disimpan.

Job dianggap `skipped` jika:

- Channel disabled di user settings.
- Provider belum configured.
- Push tidak punya target aktif.
- Push target hanya berisi device yang sudah tidak terdaftar.

Job dianggap `failed` dan retry-able jika ada failure transient provider/database. Karena job sudah per-channel, email gagal tidak menghilangkan retry push atau in-app, dan sebaliknya.

## Push DeviceNotRegistered

Jika OneSignal mengembalikan subscription ID dalam `invalid_player_ids`:

- Backend deactivate token itu.
- Worker lanjut memproses token lain.
- Error ini menjadi delivery `skipped`, bukan transient failure.

Untuk endpoint test push, kondisi ini dikembalikan sebagai:

```json
{
  "error": "push device not registered"
}
```

dengan status `410`.

## Reminder Copy

Untuk `offset_days = 0`:

```text
Title: <important day title> is today
Body:  <important day title> is today.
```

Untuk `offset_days > 0`:

```text
Title: <important day title> is in <offset_days> days
Body:  <important day title> is coming in <offset_days> days.
```

Email body HTML:

```html
<p>Hi {username},</p>
<p>{body}</p>
<p>Date: {occurrence_date}</p>
<p>Event: {important_day_title}</p>
```

## In-App Notification Data

Notification dari worker memakai:

```json
{
  "type": "important_day_reminder",
  "important_day_id": "550e8400-e29b-41d4-a716-446655440000",
  "data": "{\"important_day_id\":\"550e8400-e29b-41d4-a716-446655440000\",\"reminder_job_id\":\"550e8400-e29b-41d4-a716-446655440000\",\"occurrence_date\":\"2026-05-13\"}"
}
```

`data` di REST notification adalah string JSON.

## FE Implications

- Jika user mematikan `push`, push reminder berhenti walau device token masih aktif.
- Jika user mematikan `in_app`, notification badge tidak akan bertambah dari reminder baru.
- Jika email provider belum configured, user tidak akan menerima email tetapi reminder job tidak otomatis dianggap error.
- Untuk debugging push, pakai `POST /v1/devices/:id/test-push`.
