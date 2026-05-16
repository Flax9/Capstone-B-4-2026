# Capstone Backend - Project Boilerplate

Repositori ini menggunakan Docker Compose untuk menyimulasikan arsitektur backend perbankan (High Availability, Prometheus Monitoring, K6 Load Testing).

## 🗺️ Peta Wilayah & Pembagian Tugas
Agar tidak terjadi konflik (*merge conflicts*), kita telah membagi struktur folder berdasarkan tanggung jawab masing-masing. Silakan fokus bekerja pada direktori Anda sendiri:

```text
capstone-backend-b4/
├── docker-compose.yml           <-- (Infrastruktur Utama)
├── backend-api/                 <-- (Wilayah Nadia: Kode Spring Boot & Dockerfile)
│   ├── Dockerfile
│   └── application-dummy.yml
├── database-init/               <-- (Wilayah Seva: Migrasi & Skema DB)
│   └── 01-init-dummy.sql
├── k6-scripts/                  <-- (Wilayah Rafael: Skenario Load Testing)
│   └── dummy-load-test.js
└── monitoring/                  <-- (Wilayah Ego & Vaness## 🚀 Panduan Deployment (Sistem 3-Tier Workflow)

Bagian ini berisi langkah-langkah presisi tinggi untuk menjalankan, mengelola, dan mematikan environment arsitektur Master-Replica perbankan Anda secara terstruktur. Pastikan Anda sudah menginstall Docker Desktop.
```

### TIER 1: Native Local Development (Hot-Reload Mode)
Pada mode ini, *source code* (Golang) dieksekusi secara *native* pada Host OS untuk memfasilitasi proses *debugging* dan *hot-reloading* yang instan tanpa *overhead* proses *build* internal Docker. Klaster kontainer hanya dialokasikan untuk menyediakan dependensi persisten infrastruktur data (PostgreSQL Master, PostgreSQL Replica, dan Redis).

1. **Automated Bootstrapping (Recommended)**:
   Gunakan *batch script* otomatis yang telah disediakan untuk memicu proses agregasi *stateful containers* secara *detached* (latar belakang), yang kemudian diikuti dengan inisialisasi server kompilasi JIT (*Just-In-Time*) Golang pada terminal utama (*foreground*):
   ```cmd
   .\run-dev.bat
   ```
   *(Catatan Arsitektur: Repositori dikonfigurasi menyerupai pola monolitik dengan fitur workspace `go.work`. Meskipun kode secara logis diisolasi ke dalam sub-direktori `backend-api/`, kompilasi dapat dieksekusi langsung dari *root directory* tanpa konflik routing).*

2. **Manual Execution (Alternative)**:
   - Salin file `.env.development.template` menjadi sebuah file konfigurasi `.env`.
   - Inisialisasi *Stateful Containers*: `docker-compose up -d postgres-master postgres-replica redis-cache`
   - Eksekusi *Application Server*: `go run ./backend-api`

### TIER 2: Model Uji Coba Terintegrasi (Docker Cluster) 🐳
Apabila fitur *Backend* sudah rampung, Anda wajib mengetes interaksi aplikasi secara utuh layaknya lingkungan korporat yang kokoh menempel pada *Virtual Network* Docker lokal (Mencakup K6 *Load Tester* & Prometheus Grafana).
1. Pastikan seluruh relasi dan DDL termutakhir di `01-init.sql`.
2. Peluncuran Klaster Menyeluruh (API + DB + Redis + Tools):
   ```bash
   docker-compose up -d --build
   ```
3. **Akses Panel Komando Visual (Monitoring)**:
   - 🟢 **Status Inti Backend**: `http://localhost:9000/health`
   - 📡 **Portal Data Prometheus**: `http://localhost:9090`
   - 📊 **Dasbor Utama Grafana**: `http://localhost:3000` *(Kredensial Bawaan Login: `admin` / Password: `admin`)*

### TIER 3: Model Penyerahan / Production (Eksternal DB) 🌐
Dalam lintasan menuju *Go-Live* atau Production, lapisan tangki penyimpanan *database* lokal **wajib** pensiun dan digantikan sepenuhnya oleh penyedia *Managed External Database* (seperti AWS RDS / Cloud SQL / Server Fisik On-Premise Mitra). Strategi pemisahan ini mutlak guna menjamin IOPS (Input/Output Throughput) level tinggi tanpa *bottleneck*, keamanan enkripsi finansial, dan pemulihan *Point-In-Time-Recovery*.
1. **Siapkan Kredensial**: Salin draf suci `.env.partner.template` menjadi sebuah `.env` di peladen produksi / kantor Mitra. Buka dan patrikan/ubah alamat URL-nya agar menarget lempeng IP mesin Database Eksternal raksasa mereka.
2. **Penyalaan Khusus Produksi**: Jangan gunakan file *compose* standar! Cukup terbangkan klaster murni secara terpusat "API Only" melalui deklarasi ekstensi ini di konsol terminal Production:
   ```bash
   docker-compose -f docker-compose-external.yml up -d --build
   ```
Dalam sekejap, kontainer mandiri `backend-api` Anda secara gagah melebarkan pipa transaksinya (*JPA/GORM*) menerjang langsung ke DB Awan Mitra!

---

## 🛠️ Tata Kelola Infrastruktur Tambahan

### 1. Eksekusi Skema Struktur (Tabel / DDL)
Sebaiknya Anda membaca skrip `database-init/01-init.sql`. Skrip mutakhir ini terinjeksi saklar kebal overwriting (`CREATE TABLE IF NOT EXISTS`). Saat kontainer *postgres-master* lahir, file ini ditayangkan otomatis. Bilamana Anda beralih pada Tier 3 (External Postgres/DBeaver Pihak Ketiga), maka cukup *copy-paste* jajaran _query_ tersebut ke layar *admin* klien mitra.

### 2. Memantau Denyut Log Sistem & Stress Test
Membaca getar langkah kontainer Java/Go: `docker-compose logs -f backend-api`
Menekan pemicu letusan K6 (Simulasi 30k TRX/Jam): `docker-compose --profile testing run --rm k6-loadtester run /scripts/dummy-load-test.js`

### 3. Pemusnahan Ke Titik Nol (Clean Slate Mode)
> [!CAUTION]
> Taktik bumi hangus. Jika Anda ingin meratakan `postgres_master_data` dan menyapu bersih seluruh residu kontainer serta ketersediaan datanya hingga menguap 100%:
```bash
docker-compose down -v
```

---

## 🏗️ Keunggulan Infrastruktur Arsitektur Puncak
- **Isolasi Beban Transaksi Master-Replica**: Penempatan eksekusi perintah (Insert/Update) terkarantina pada **DB Master**, selagi perintah gelombang perambah massal (Select) dibelokkan instan ke **DB Replica**, mereduksi drastis hambatan latensi perpesanan antrean (*queue latency*).
- **Elastisitas Distribusi Hybrid (Java & Go)**: Sukses membuktikan penyadapan fungsional lintas instans dengan proxy asinkron (`LazyConnectionDataSourceProxy`) dan (`dbresolver`).
- **Idempoten DML & Skema Keamanan Tanpa Status (`Stateless`)**: Relasi tabel mumpuni berpadukan token jwt gembok basis `jwt_jti`, serta fitur `ON CONFLICT DO NOTHING` mengisolasi sistem dari kepanikan gempuran duplikasi peluncuran yang berulang-ulang.

---

## Fitur MVP Golang API (Fiber v2 + GORM + Redis)

Arsitektur logika bisnis murni telah selesai diintegrasikan ke dalam ekosistem `main.go` yang berbasiskan kerangka kerja (*Framework*) web **Fiber v2**. Entitas API ini telah berjalan stabil pada *port* TCP **`9000`** dan berhasil mengimplementasikan kapabilitas mekanik *Read/Write Separation* di pangkalan logikanya:

1. **`GET /api/accounts/:id` (Account Balance Retrieval)**
   - Mengimplementasikan pola arsitektur sentris **Cache-Aside**. Parameter *Request* mula-mula disadap oleh memori *Redis Cache* (L1). Apabila terdeteksi insiden *Cache Miss*, fungsi pendaur kueri *GORM* akan mem-proksinya secara absolut menuju instans PostgreSQL **Replica (15433)**, lalu status simpanan *Redis* langsung diperbarui.
2. **`POST /api/transactions/transfer` (Mutasi Finansial & Persistensi ACID)**
   - Algoritma divalidasi dengan pengawalan integritas ketat (Atomicity) menggunakan utilitas *Pessimistic Locking* tingkat baris database SQL (`SELECT ... FOR UPDATE`). Seluruh rentet kode modifikasi *(DML)* hanya akan mengeksekusi PostgreSQL **Master (15432)** guna menghalau deviasi rasio ganda (*Race Condition*). Sesegera pasca proses komit selesai, sistem mengeksekusi serbuan *Cache Invalidation* khusus untuk mencabut entri parameter Redis yang telah direkayasa.
3. **`POST /api/auth/login` (Autentikasi & Perekaman Audit Log)**
   - Penerapan pola disjungsi arsitektural (*Asynchronous-Like Subroutine*). Validasi tarikan kueri identitas *(SELECT)* dipropagasi mandiri menyusuri klaster Replica. Secara simultan di sisa milidetik yang berjalan, mekanisme fungsi GORM memaksa rekaman rute aktivitas *(INSERT Audit Logs / Alamat IP)* persisten menuju *storage Master* untuk menjamin rekam kepatuhan (Compliance Standard) di industri perbankan yang kokoh.

### 🚥 Prosedur Standar Uji Mutu Pengikatan Endpoint (Quality Assurance)
Guna menghindari anomali *string escaping* interpolasi JSON pada utilitas eksternal di ekosistem OS Windows, pengujian *(testing)* administratif diamanatkan untuk secara murni menggunakan _cmdlet_ bawaan `Invoke-RestMethod` pada sesi PowerShell lokal. Penarikan UUID dinamis dapat dilakukan melalui tabel `accounts` pada modul pangkalan data pengujian *(sandbox)*.
Berikut adalah parameter spesifikasi uji (*Test Payload*) baku yang telah tersertifikasi:

### 🚥 Prosedur Standar Uji Mutu Pengikatan Endpoint (OS Windows - PowerShell)

Karena Windows PowerShell memiliki struktur pembacaan karakter tanda kutip tunggal/ganda (JSON) yang ketat, pengujian API ini direkomendasikan menggunakan perakitan variabel native PowerShell.

Pastikan Docker Services (Tier 2) sudah aktif dan berjalan. Buka aplikasi Terminal PowerShell, arahkan (CD) ke folder tempat `docker-compose.yml` berada, lalu ikuti urutan perintah ini:

## TAHAP 0 : Mengumpulkan UUID (Universally Unique Identifier)
Dikarenakan Endpoint Transfer dan Cek Saldo membutuhkan ID unik (account_id) bawaan Database dan bukan hanya angka rekening biasa demi keamanan.
Tarik daftar UUID-nya secara paksa langsung dari instans kontainer PostgreSQL:
```
docker-compose exec -e PGPASSWORD=password postgres-master psql -U user -d capstonev2 -P pager=off -c "SELECT account_id, account_number, balance FROM accounts;"
```

Dicatat dua UUID yang keluar dari tabel pada Terminal Anda. Variabel Asumsi:
- UUID Pengirim (Akun 1): 924de2cf-e950-4f92-8e37-ae2eb7dda7e5
- UUID Penerima (Akun 2): e3acd2bc-94d1-475e-ac7a-12fe405ad426

## TAHAP 1 : Autentikasi Sistem & Mencetak Token Kriptografi (JWT)
Harap catat, Token JWT Server Golang ini memiliki batas hangus waktu (Expired) *hanya 15 menit*. Oleh karena itu eksekusi Tahap 2 dan 3 wajib dilakukan pada saat Token ini baru dicetak.
Jalankan deretan skrip PowerShell ini secara berurutan:

```powershell
$bodyLogin = @{ username="nasabah_01"; password="rahasia" } | ConvertTo-Json
$response = Invoke-RestMethod -Uri "http://localhost:9000/api/auth/login" -Method POST -Body $bodyLogin -ContentType "application/json"
$token = $response.token
$headers = @{ "Authorization" = "Bearer $token" }
Write-Host "Sesi JWT Aktif! Token Tersimpan: $token"
```

## TAHAP 2 : Ekstraksi Saldo Lapis Utama (Method GET)
Karena variabel `$headers` (yang berisi Token Kunci) masih ada di rekam jejak memori PowerShell Anda, Anda bebas menembus Gateway tanpa ditolak.

### Ganti dengan UUID yang didapat di Tahap 0
```powershell
$idAkunTarget = "924de2cf-e950-4f92-8e37-ae2eb7dda7e5" 
```

### Kirim Request GET nya (Hanya Memerlukan URL & Headers) Guna melihat status akun seperti sisa saldo
```powershell
Invoke-RestMethod -Uri "http://localhost:9000/api/accounts/$idAkunTarget" -Method GET -Headers $headers | ConvertTo-Json -Depth 5
```

## TAHAP 3 : Simulasi Transaksi Mutasi Rekening Terkunci (Method POST)
Lakukan instruksi POST dengan menyematkan formulir pengiriman saldo secara dinamis melalui parameter yang dikonversi menjadi format standar web murni (JSON):

### Rakit Payload Body Transaksi dengan aman fungsi transfer uang
### Payload konfig
```powershell
$bodyMutasi = @{ 
    from_account_id = "924de2cf-e950-4f92-8e37-ae2eb7dda7e5";    # UUID Pengirim
    to_account_id   = "e3acd2bc-94d1-475e-ac7a-12fe405ad426";    # UUID Penerima
    amount          = 50000                                      # Saldo Mutasi Rp. 50.000
} | ConvertTo-Json
```

### Kirim Request (POST) guna mentransfer uang
```powershell
Invoke-RestMethod -Uri "http://localhost:9000/api/transactions/transfer" -Method POST -Body $bodyMutasi -ContentType "application/json" -Headers $headers
```

## 🧨 Protokol Evaluasi Ketahanan (SRE Load Testing - K6)
Arsitektur API ini sejatinya secara gagah dilindungi oleh perisai anti DDoS (`Rate-Limiter` 60 request/menit). Jika pengujian beban *(Load Testing)* dilesatkan secara telanjang, API Golang akan menampar balik laju `k6` dengan pesan peringatan `HTTP 429 Too Many Requests`.

Guna mendobrak pembatas tersebut khusus di lingkungan akademis/pengujian, prosedur injeksi tim SRE telah digawangi oleh protokol **Manajemen Jalur Retas (Testing Bypass Route)**:
1. **Prediktabilitas UUID (Account Seeding):** Skema inisialisasi basis data (`01-init.sql`) telah dijahit paku *(hardcoded)* dengan dua kantong UUID identitas (*Account ID*) statis: `924de2cf...` dan `e3acd2bc...`. Rekayasa ini memustahilkan munculnya insiden rekaman hilang `404 Not Found` setiap kali purwarupa SRE direstart secara persisten.
2. **Sandi V.I.P Pelolosan Pertahanan:** Seluruh selongsong armada `k6-scripts/*.js` dipersenjatai dengan penyuntik rahasia *HTTP Header* bernilai `X-Test-Bypass: b7fc809a-super-secret-key-capstone`. Selama `k6` bersenandungkan kalimat sandi magis ini, sang *Rate-Limiter* Golang akan menunduk dan membiarkan puluhan ribu permohonan lewat tak terbatas; sambil murni mengunci serangan penyerang asing publik yang tidak tahu sandi tersebut!

### Skema Pemantik Senjata K6 (Native Windows Host)
Lakukan pendobrakan peluru skenario latensi secara asali dari Command Prompt atau PowerShell Anda (Dengan prasyarat instalasi murni `winget install k6`):

```powershell
# ✅ Uji Subsistem I/O Disk & Pembangkitan JWT (Auth Login)
k6 run k6-scripts\load_test_login.js

# ✅ Uji Pembebanan Subsistem Baca Cepat (Redis Cache Target)
k6 run k6-scripts\load_test_cek_saldo.js

# ✅ Uji Penyiksaan Subsistem Tulis Ekstrem (Postgres ACID Master)
k6 run k6-scripts\load_test_transfer.js
```
---

## 🛡️ Arsitektur Keamanan Lapis Baja (Prototyping Level-2)
Merespons standar akademik ketat untuk purwarupa layanan perbankan, infrastruktur mesin *Fiber* telah ditempa dengan 3 lapis perisai interseptor tambahan:
1. **Stateless JWT Cryptography (`HS256`)**: Memusnahkan ketergantungan validasi berbasis *database* untuk setiap rute Mutasi dan Cek Saldo. Eksekusi kini murni hanya mengandalkan perhitungan verifikasi Tanda Tangan Kriminologis *(Token Bearer 15-Menit)* dari memori RAM, mempercepat latensi transaksi hingga persekian milidetik.
2. **Global Sliding-Window Rate Limiter**: Anti *DDoS (L7)* aktif. Seluruh *endpoint* telah dipasangkan gembok kecepatan. Jika k6/JMeter memborbardir IP dengan lebih dari **60 Request / Menit**, *Fiber* akan otomatis membanting jaringan tersebut dengan respon `HTTP 429: Too Many Requests`.
3. **Penyadap Metrik Prometheus (`/metrics`)**: Terminal klandestin murni *(Target Scraping Grafana)* yang langsung dijahit ke nadi inti (Core) mesin Golang menggunakan *PromHTTP Native Adaptor*. Mengekspos jutaan angka performa mulai dari *Garbage Collector (GC)* hingga Konsumsi RAM (*Alloc Bytes*) mentah ke dunia luar untuk dibuktikan secara kasat mata pada layar penguji.

---

## 📝 Changelog (Pembaruan Terbaru)

**[06 Mei 2026] - Implementasi Envoy Proxy & Load Balancer**
- 🚀 **Penambahan Envoy Proxy**: Arsitektur kini menggunakan Envoy sebagai pintu masuk utama (*API Gateway* & *Load Balancer*).
- ⚖️ **Skalabilitas Backend**: Layanan `backend-api` tidak lagi diakses langsung secara publik (port 9000 dipindahkan ke Envoy) dan dikonfigurasi untuk secara asali berjalan sebanyak 3 instance (*replicas*). Envoy mendistribusikan lalu lintas secara merata dengan algoritma *Round Robin*.
- 📊 **Monitoring Lanjutan**: Konfigurasi Prometheus diperbarui menggunakan `dns_sd_configs` untuk secara presisi mendeteksi dan memantau 3 instance `backend-api` sekaligus, serta mengekstraksi metrik bawaan dari port admin Envoy (`9901`).
- 🚦 **K6 Load Testing Terpadu**: Penambahan skrip `k6-scripts/load_test_combined.js` yang memborbardir sistem dengan 1 juta iterasi request secara acak (Login, Cek Saldo, Transfer) untuk menguji ketahanan Envoy dan *cluster backend* secara serentak.

---

## 🎯 Target Pekerjaan Selanjutnya (To-Do List Tim)

Saat ini, *Boilerplate* infrastruktur sudah menyala sempurna menggunakan kontainer *dummy/placeholders*. Berikut adalah hal yang perlu dilakukan oleh masing-masing PIC agar sistem ini menjadi *Real Application*:

- **Nadia (Backend API)**: 
  - Ganti isi folder `backend-api/` dengan *source code* Spring Boot aslinya. 
  - Timpa konfigurasi *dummy* (seperti `pom.xml`, `Application.java`, dan `application.yml`) dengan *Logic* aplikasi perbankan yang riil. Pastikan koneksi ke `postgres-db` dan `redis-cache` tersambung dengan benar.

- **Seva (Database)**: 
  - Ganti file `database-init/01-init-dummy.sql` dengan skema DDL (Data Definition Language) PostgreSQL yang sebenarnya.
  - Setiap kali ada *table* atau relasi baru, silakan tambahkan file bereksistensi `.sql` di folder tersebut (contoh: `02-insert-master-data.sql`).

- ~~**Rafael (Load Testing)**~~ *(STATUS: SELESAI)*:
  - ~~Gubahan struktur ID persisten pada inisialisator basis data dan injeksi selubung saklar pintu pelolosan *Rate-Limiter* telah diimplementasikan.~~
  - ~~Tiga skenario pengujian paripurna K6 (I/O Login, Membaca Cepat Redis, dan Tulis Ekstrem Master) terbukti absolut berhasil memuntahkan iterasi tanpa celah kemacetan di target milidetik SRE SLO yang menawan.~~

- **Ego & Vanessa (Monitoring)**:
  - Buka *Dashboard* Grafana di `http://localhost:3000`.
  - Sambungkan *Data Source* ke Prometheus (`http://prometheus:9090`) / (`http://localhost:9090`).
  - Buat *Dashboard* kustomisasi untuk memantau metrik dari Spring Boot Actuator, JVM memory, dan Database connection pool.

  # Note Cek Health :
  - (`http://localhost:9000/health`)
  # Note Cek Metrics :
  - (`http://localhost:9000/metrics`)


# ⚠️ SOP PENGGUNAAN REPOSITORI GITHUB KELOMPOK 6 (WAJIB DIBACA & DIPATUHI) ⚠️

Repositori ini menggunakan arsitektur *container* (Docker Compose) yang saling terhubung. Agar infrastruktur tidak hancur atau *error* saat disatukan, seluruh anggota tim **WAJIB** mematuhi 4 aturan emas berikut:

### 1. DILARANG KERAS PUSH LANGSUNG KE BRANCH `main`
Branch `main` adalah **lingkungan produksi/stabil** yang harus selalu bisa dijalankan kapan saja. Tidak ada satu pun anggota yang diizinkan melakukan `git commit` atau `git push` langsung ke branch `main`.

### 2. WAJIB BEKERJA DI BRANCH MASING-MASING
Sebelum mulai menulis kode, kalian **wajib** menarik data terbaru dari `main` dan membuat *branch* baru dengan format `fitur/nama-tugas`.
* Perintah: `git pull origin main` lalu `git checkout -b fitur/nama-tugas kalian`
* Contoh Nadia: `git checkout -b feature/api-spring-boot`
* Contoh Seva: `git checkout -b feature/db-schema`
* Contoh Rafael: `git checkout -b testing/k6-load-test`

### 3. PATUHI BATASAN RUANG KERJA (FOLDER)
Kerjakan tugas kalian **HANYA** di dalam direktori yang sudah disediakan sesuai pembagian *role*. Dilarang keras mengedit *file* di luar *folder* kalian untuk mencegah konflik.
* **Nadia (Backend):** Hanya bekerja di folder `backend-api/`
* **Seva (Data Engineer):** Hanya bekerja di folder `database-init/`
* **Rafael (Load Tester):** Hanya bekerja di folder `k6-scripts/`
* **Ego & Vanessa (Monitoring):** Hanya bekerja di folder `monitoring/`
* ⛔ **DILARANG KERAS** menyentuh, mengedit, atau memindahkan file `docker-compose.yml` tanpa persetujuan **Furqon (DevOps/Infrastructure)**.

### 4. ALUR PENGUMPULAN KODE (PULL REQUEST)
Jika kode di *branch* kalian sudah selesai dan sukses dijalankan di komputer masing-masing, ikuti alur penyatuan kode berikut:
1. *Push* kode ke *branch* kalian sendiri (`git push origin nama-branch-kalian`).
2. Buka halaman GitHub dan buat **Pull Request (PR)** dari *branch* kalian ke arah `main`.
3. Beri tahu Furqon di grup WhatsApp bahwa PR sudah dibuat.
4. **Furqon (sebagai Gatekeeper)** akan melakukan *review* terhadap PR tersebut. Jika aman dan tidak merusak infrastruktur Docker, Furqon yang akan melakukan **Merge**.

**Catatan Tegas:** Jika ada kode yang merusak struktur Docker Compose atau melanggar SOP ini, Pull Request akan langsung di-*reject* dan dikembalikan untuk diperbaiki.

> **Status Saat Ini (DevOps)**: Infrastruktur Docker Compose telah stabil dan siap digunakan pengembangan paralel. Jika ada *request* penambahan *Environment Variable* atau *Dependency* spesifik pada Docker, silakan koordinasikan ke tim *Infra/DevOps*. Kodingan kalian ditunggu! 🚀
