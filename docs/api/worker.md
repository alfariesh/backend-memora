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
CLOUDFLARE_EMAIL_ACCOUNT_ID=
CLOUDFLARE_EMAIL_API_TOKEN=
CLOUDFLARE_EMAIL_FROM_EMAIL=
EXPO_PUSH_ACCESS_TOKEN=
REMINDER_WORKER_BATCH_SIZE=50
REMINDER_WORKER_POLL_INTERVAL=1m
```

Meaning:

| Env | Required | Notes |
| --- | --- | --- |
| `CLOUDFLARE_EMAIL_ACCOUNT_ID` | for email delivery | Cloudflare account ID. |
| `CLOUDFLARE_EMAIL_API_TOKEN` | for email delivery | Bearer token for Cloudflare Email Service. |
| `CLOUDFLARE_EMAIL_FROM_EMAIL` | for email delivery | Sender email from onboarded Cloudflare sending domain. |
| `EXPO_PUSH_ACCESS_TOKEN` | optional | Sent as Expo Authorization header if present. |
| `REMINDER_WORKER_BATCH_SIZE` | no | Default `50`. |
| `REMINDER_WORKER_POLL_INTERVAL` | no | Default `1m`. |

Cloudflare email config kosong tidak membuat app crash. Email channel yang tidak configured akan di-skip.

Expo access token kosong juga tidak membuat app crash. Backend tetap mengirim request Expo tanpa Authorization.

## Job Lifecycle

1. Important day dibuat atau diupdate.
2. Reminder rules dibuat atau diganti.
3. Backend membuat pending reminder jobs berdasarkan occurrence berikutnya.
4. Worker claim due jobs sesuai batch size.
5. Worker filter channels berdasarkan user settings terbaru.
6. Worker mencoba delivery ke channel yang tersisa.
7. Jika job dianggap sukses, worker mark sent dan schedule occurrence tahun berikutnya.
8. Jika job gagal total, worker mark failed dan retry sampai attempt ketiga.

## Channel Filtering

Rule tersimpan pada job, tetapi sebelum delivery worker melihat user settings terbaru.

Contoh:

```json
{
  "job_channels": ["email", "in_app", "push"],
  "user_notification_channels": ["in_app", "push"]
}
```

Worker hanya mencoba `in_app` dan `push`.

Jika `notification_channels` user settings adalah `[]`, tidak ada channel yang dikirim.

## Delivery Behavior

Worker mencoba channel berikut jika ada dalam filtered channels:

| Channel | Behavior |
| --- | --- |
| `email` | Kirim HTML email via Cloudflare Email Service. Jika Cloudflare config kosong, channel email di-skip tanpa failure. |
| `in_app` | Simpan notification ke database. Ini yang muncul di `/v1/notifications/`. |
| `push` | Kirim Expo push ke semua active device tokens user. Jika tidak ada active device, ini tidak dianggap failure. |

Job dianggap sukses jika:

- Minimal satu channel berhasil, atau
- Tidak ada failure yang perlu di-retry.

Job dianggap gagal jika:

- Ada failure provider/database, dan
- Tidak ada channel lain yang berhasil.

Jika ada failure tetapi channel lain berhasil, job tetap ditandai sent. Contoh: in-app berhasil tetapi push provider gagal, user tetap punya in-app notification dan job tidak retry.

## Push DeviceNotRegistered

Jika Expo mengembalikan `DeviceNotRegistered`:

- Backend deactivate token itu.
- Worker lanjut memproses token lain.
- Error ini tidak otomatis membuat semua job gagal jika tidak ada failure lain.

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
