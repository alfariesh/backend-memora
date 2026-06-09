# Memora — Identifikasi Improvement (Bug, Performa, Enhancement)

Dokumen ini memetakan peluang perbaikan pada backend Memora hasil telaah kode
(REST/gRPC/RPC, worker reminder, lapisan repo PostgreSQL). Setiap temuan
menyertakan lokasi (`file:line`), dampak, severity, dan rekomendasi singkat.

> Status: **dokumen identifikasi** — belum ada perubahan kode aplikasi. Dipakai
> sebagai backlog perbaikan terprioritas.

## Ringkasan Prioritas

| #  | Temuan                                            | Kategori      | Severity  |
|----|---------------------------------------------------|---------------|-----------|
| 1  | Double-send reminder saat `scheduleNext` gagal     | Bug           | High      |
| 6  | Tidak ada rate limiting pada endpoint auth         | Keamanan      | High      |
| 20 | `context.Background()` di handler RPC (tanpa timeout) | Observability | Med-High  |
| 2  | `attempts` terbuang saat klaim, bukan saat kirim   | Bug           | Medium    |
| 4  | Partial-failure pengiriman tertelan                | Bug           | Medium    |
| 11 | Validasi konfigurasi (JWT secret) minim            | Keamanan      | Medium    |
| 12 | `Upcoming` memuat hingga 1000 baris ke memori      | Performa      | Medium    |
| 19 | Worker tanpa structured logging per-job            | Observability | Medium    |
| 23 | Cakupan test jalur inti rendah                     | Kualitas      | Medium    |
| 7  | CORS tidak dikonfigurasi                           | Keamanan      | Medium    |
| 3  | Read-after-write pada `MarkRead`                   | Bug           | Low       |
| 5  | Cleanup error diabaikan (jalur non-atomik Create)  | Bug           | Low       |
| 8  | Tidak ada batas panjang field string               | Keamanan      | Low-Med   |
| 9  | Batas atas query param belum dipaksakan di edge    | Keamanan      | Low-Med   |
| 10 | User enumeration via timing pada login             | Keamanan      | Low       |
| 13 | COUNT + SELECT terpisah pada List                  | Performa      | Low       |
| 14 | Insert dalam loop (bukan batch)                    | Performa      | Low-Med   |
| 15 | `Deactivate` device dalam loop saat push           | Performa      | Low       |
| 16 | Lookup device tunggal memuat semua device          | Performa      | Low       |
| 17 | `ListActiveByUser` tanpa LIMIT                      | Performa      | Low       |
| 18 | Indeks `reminder_rules(user_id)` hilang            | Performa      | Low-Med   |
| 21 | Error `resp.Body.Close()` sebagian tertelan        | Observability | Low       |
| 22 | Tidak ada `/readyz` / health check DB              | Observability | Low-Med   |
| 24 | Context key bertipe string (bukan tipe khusus)     | Kualitas      | Low       |

---

## A. Bug Korektnes

### 1. Double-send reminder saat `scheduleNext` gagal — **High**
`internal/usecase/reminder/reminder.go:117-123`

`deliverJob` memanggil `MarkSent` (di-commit) lebih dulu, lalu `scheduleNext`.
Bila `scheduleNext` gagal, `deliverJob` mengembalikan error sehingga `RunOnce`
(`reminder.go:64-69`) memanggil `MarkFailed(retry=true)` yang membalik status job
dari `sent` kembali ke `pending`. Pada putaran worker berikutnya job dikirim ulang
→ **email/push/notifikasi ganda** ke pengguna.

**Rekomendasi:** pisahkan kegagalan "penjadwalan occurrence berikutnya" dari
kegagalan "pengiriman". Job yang sudah `sent` tidak boleh di-`retry`. Opsi: jadikan
`MarkSent` + `scheduleNext` satu transaksi, atau telan/anti-error `scheduleNext`
(log saja) tanpa memengaruhi status job yang sudah terkirim.

### 2. `attempts` terbuang saat klaim, bukan saat kirim — **Medium**
`internal/repo/persistent/reminder_job_postgres.go:102-107`

`ClaimDue` menaikkan `attempts = attempts + 1` pada UPDATE klaim. Bila worker crash
setelah klaim namun sebelum benar-benar mengirim, jatah retry (`maxAttempts = 3`,
`reminder.go:17`) ikut habis tanpa percobaan kirim nyata. Akibatnya sebuah reminder
bisa berakhir `failed` tanpa pernah benar-benar dicoba dikirim.

**Rekomendasi:** naikkan `attempts` saat `MarkFailed` (percobaan yang benar-benar
gagal), bukan saat klaim. Klaim cukup men-set `locked_until`.

### 3. Read-after-write pada `MarkRead` — **Low**
`internal/usecase/notification/notification.go:64-76`

UPDATE diikuti `GetByID` terpisah (dua query). Selain boros, ada celah race bila
notifikasi terhapus di antara dua panggilan.

**Rekomendasi:** gunakan `UPDATE ... RETURNING` untuk mengembalikan baris dalam satu
operasi.

### 4. Partial-failure pengiriman tertelan — **Medium**
`internal/usecase/reminder/reminder.go:93-119`

Logika `delivered = delivered || ok` menjadikan job `sent` selama **minimal satu**
kanal sukses. Jika email sukses tapi push gagal, job ditandai `sent` dan kegagalan
push hilang tanpa retry maupun log (`reminder.go:113` hanya return error bila tidak
ada kanal yang berhasil).

**Rekomendasi:** catat kegagalan per-kanal (minimal log terstruktur). Pertimbangkan
status partial atau pelacakan delivery per-kanal agar kanal yang gagal bisa di-retry.

### 5. Cleanup error diabaikan pada jalur non-atomik Create — **Low**
`internal/usecase/importantday/important_day.go:96-100`

Pada fallback non-atomik, `_ = uc.dayRepo.Delete(...)` membuang error kompensasi.
Bila penghapusan kompensasi gagal, akan tertinggal `important_day` yatim (tanpa rule)
dan kegagalan ini tidak terlihat. (Jalur atomik `StoreWithReminderRulesAndJobs` aman;
ini hanya fallback.)

**Rekomendasi:** log error kompensasi; pertimbangkan menjadikan jalur atomik wajib.

---

## B. Keamanan

### 6. Tidak ada rate limiting pada endpoint auth — **High**
`internal/controller/restapi/router.go`, `internal/controller/restapi/v1/user.go`

Endpoint `register`/`login`/`refresh` tanpa pembatasan laju → rentan brute-force
kredensial dan abuse.

**Rekomendasi:** tambahkan middleware limiter Fiber (`github.com/gofiber/fiber/v2/middleware/limiter`),
dengan batas lebih ketat khusus grup `/v1/auth/*`.

### 7. Tidak ada konfigurasi CORS — **Medium**
`internal/controller/restapi/router.go:40-49`

Tidak ada middleware CORS eksplisit. Untuk klien web lintas-origin perlu kebijakan
allowlist yang jelas agar tidak terlalu permisif maupun memblokir klien sah.

**Rekomendasi:** tambahkan `middleware/cors` dengan allowlist origin dari konfigurasi.

### 8. Tidak ada batas panjang field string — **Low-Medium**
`internal/controller/restapi/v1/request/*`

Field seperti `title`, `description`, `person_name` tidak memiliki batas panjang di
lapisan API → memungkinkan payload sangat besar.

**Rekomendasi:** tambahkan tag validasi `max=` pada DTO request.

### 9. Batas atas query param belum dipaksakan di edge — **Low-Medium**
`internal/controller/restapi/v1/mobile.go` (`upcoming_days`),
`internal/controller/restapi/v1/important_day.go:348-360` (pagination)

Usecase memang menormalkan (`maxUpcomingLookaheadDays = 3660`, `maxListLimit = 100`),
namun `Upcoming` tetap memuat hingga 1000 baris di memori (lihat C-12), sehingga
parameter besar tetap menjadi vektor beban.

**Rekomendasi:** validasi/clamp di edge dan dorong batas ke query DB.

### 10. User enumeration via timing pada login — **Low**
`internal/usecase/user/user.go:88-96`

Saat email tidak ditemukan, `Login` langsung return `ErrInvalidCredentials` tanpa
menjalankan `bcrypt.CompareHashAndPassword`, sehingga waktu respons berbeda dengan
kasus password salah → memungkinkan enumerasi akun.

**Rekomendasi:** jalankan dummy bcrypt compare untuk menyeragamkan waktu respons.

### 11. Validasi konfigurasi minim — **Medium**
`config/config.go`, `.env.example`

Tidak ada validasi kekuatan/panjang `JWT_SECRET`, dan tidak ada deteksi penggunaan
nilai default contoh (`your-secret-key-change-in-production`).

**Rekomendasi:** tambahkan validasi pasca-parse di `NewConfig()` (panjang minimum
secret, tolak nilai default contoh, validasi format URL).

---

## C. Performa

### 12. `Upcoming` memuat hingga 1000 important day ke memori — **Medium**
`internal/usecase/importantday/important_day.go:140-184`

Method memuat seluruh important day user (limit hardcoded 1000), memanggil
`NextOccurrence` per-baris, lalu memfilter rentang tanggal & memaginasi di aplikasi.
Berat secara memori & CPU untuk user dengan banyak entri.

**Rekomendasi:** dorong perhitungan occurrence/filter rentang & paginasi ke query DB
(atau pra-hitung occurrence berikutnya yang ter-index).

### 13. COUNT + SELECT terpisah pada List — **Low**
`internal/repo/persistent/important_day_postgres.go:91-150`,
`internal/repo/persistent/notification_postgres.go:56-116`

List menjalankan query COUNT dan SELECT terpisah → dua round-trip per permintaan.

**Rekomendasi:** gunakan window function `COUNT(*) OVER ()` untuk mengambil total +
data dalam satu query.

### 14. Insert dalam loop (bukan batch) — **Low-Medium**
`internal/repo/persistent/reminder_job_postgres.go` (`replacePendingReminderJobsTx`),
`internal/repo/persistent/important_day_postgres.go` (replace rules)

Mengganti rules/jobs dilakukan dengan menghapus lalu meng-INSERT satu per satu (N
query untuk N baris).

**Rekomendasi:** gunakan multi-value INSERT (satu query untuk seluruh baris).

### 15. `Deactivate` device dalam loop saat push — **Low**
`internal/usecase/reminder/reminder.go:199-211`

Token yang `DeviceNotRegistered` di-nonaktifkan satu per satu di dalam loop → N
UPDATE untuk N token mati.

**Rekomendasi:** kumpulkan ID token gagal, lakukan batch deactivate dalam satu query.

### 16. Lookup device tunggal memuat semua device — **Low**
`internal/usecase/device/device.go:129-142`

`getActiveDevice` memanggil `ListActiveByUser` lalu mencari satu ID via iterasi
(O(n)). Dipakai oleh `TestPush`.

**Rekomendasi:** tambahkan `GetActiveDeviceByID` di repo untuk query langsung.

### 17. `ListActiveByUser` tanpa LIMIT — **Low**
`internal/repo/persistent/device_token_postgres.go:90-109`

Tidak ada batas jumlah baris; user dengan banyak device akan mengembalikan set besar.

**Rekomendasi:** tambahkan LIMIT wajar atau paginasi.

### 18. Indeks DB kurang — **Low-Medium**
`migrations/20260513000001_create_memora.up.sql`

Tidak ada indeks pada `reminder_rules(user_id)` sehingga query per-user berpotensi
full table scan seiring data bertumbuh.

**Rekomendasi:** tambahkan migrasi `CREATE INDEX idx_reminder_rules_user_id ON reminder_rules(user_id)`.

---

## D. Observability & Kualitas

### 19. Worker/job tanpa structured logging per-job — **Medium**
`internal/usecase/reminder/reminder.go:54-73`

`RunOnce` hanya menghasilkan jumlah `processed`; kegagalan per-job (job ID, attempt,
kanal, error) tidak tercatat secara terstruktur → sulit men-debug reminder gagal.

**Rekomendasi:** injeksikan logger dan catat tiap percobaan job (ID, attempt, hasil
per-kanal, error).

### 20. Context tidak dipropagasi di handler RPC — **Medium-High**
`internal/controller/amqp_rpc/v1/*.go`, `internal/controller/nats_rpc/v1/*.go`
(mis. `amqp_rpc/v1/important_day.go:32,57`)

Seluruh handler RPC memakai `context.Background()` saat memanggil usecase → tidak ada
timeout/cancellation, operasi DB bisa menggantung tanpa batas.

**Rekomendasi:** buat `context.WithTimeout` per-handler (atau propagasi konteks dari
metadata pesan) sebelum memanggil usecase.

### 21. Error `resp.Body.Close()` sebagian tertelan — **Low**
`internal/repo/webapi/expo.go:66-70`, `internal/repo/webapi/cloudflare_email.go:69-73`

Pola `defer` hanya memperhatikan error close bila `err == nil`. Bila ada error utama,
kegagalan close diam-diam diabaikan.

**Rekomendasi:** selalu tangani/log error close (mis. variabel terpisah).

### 22. Health/readiness & DB health check — **Low-Medium**
`internal/controller/restapi/router.go:57`

Tersedia `/healthz` statis, tetapi tidak ada `/readyz` yang memverifikasi koneksi DB,
dan tidak ada pengecekan kesehatan pool secara periodik. Koneksi DB yang mati pasca
inisialisasi baru terdeteksi saat query gagal.

**Rekomendasi:** tambahkan `/readyz` yang melakukan `Ping` ke pool; pertimbangkan
health check periodik.

### 23. Cakupan test jalur inti rendah — **Medium**
~21 file test untuk ~125 file Go. Jalur kritis seperti pemrosesan reminder job dan
kalkulasi occurrence/timezone kurang ter-cover.

**Rekomendasi:** prioritaskan unit test untuk usecase `reminder` (deliver, retry,
schedule-next) dan `importantday` (occurrence, upcoming, timezone).

### 24. Context key bertipe string — **Low**
Seluruh handler REST mengakses `ctx.Locals("userID").(string)`. Key string rawan
tabrakan dan kurang aman-tipe.

**Rekomendasi:** gunakan tipe key khusus (mis. `type ctxKey string`) atau konstanta
terpusat.

---

## Catatan Implementasi (saran urutan)

1. **Korektnes dulu:** #1 (double-send), #2 (attempts), #4 (partial-failure).
2. **Keamanan cepat-menang:** #6 (rate limit auth), #11 (validasi config), #7 (CORS).
3. **Observability:** #20 (context RPC), #19 (logging worker), #22 (readyz).
4. **Performa:** #12 (Upcoming), #18 (indeks), #13/#14 (query batch).
5. **Sisanya (Low):** dikerjakan sebagai pembersihan bertahap.
