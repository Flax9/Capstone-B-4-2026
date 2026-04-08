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
└── monitoring/                  <-- (Wilayah Ego & Vanessa: Konfigurasi Observabilitas)
    └── prometheus.yml
```

## � Panduan Deployment & Cara Penggunaan (Resep)

Bagian ini berisi langkah-langkah detail untuk menjalankan, mengelola, dan mematikan environment Docker secara lokal. Pastikan Anda sudah menginstall Docker Desktop.

### 1. Persiapan File Environment (.env)
Karena arsitektur kita sekarang mendukung pemisahan dinamis *Master-Replica*, kredensial tidak lagi ditanam atau (*hardcoded*).
- Salin/Rename file `.env.template` menjadi `.env`.
- Buka `.env` dan atur propertinya:
  - Gunakan `host.docker.internal` jika komputer tersebut menggunakan/memiliki instalasi PostgreSQL eksternal (contohnya DBeaver lokal).
  - Atau biarkan saja isi bawaannya bila Anda mengandalkan kontainer DB virtual di `docker-compose`.

### 2. Pembuatan Struktur Skema (Khusus PG Eksternal)
- Bila Anda memilih menyambung ke PostgreSQL lokal di PC baru tersebut, pastikan pembuatan struktur relasi dilakukan terlebih dahulu dengan mengeksekusi isi skrip pada `schema/init.sql` pada terminal SQL/DBeaver.

### 3. Build & Start Seluruh Komponen Engine
Setelah fondasi kunci (env & skema) rampung, *compile* serta nyalakan lingkungan *backend* menggunakan perintah ajaib ini:
```bash
docker-compose up -d --build
```
> [!NOTE]
> *Tunggu beberapa saat hingga mesin Docker berhasil melakukan resolusi dan agregasi image. Pastikan layanan Docker Desktop Anda berstatus On.*

### 2. Melihat Log Aplikasi (Debugging)
Untuk melihat secara *real-time* apa yang sedang terjadi di *backend-api*:
```bash
docker-compose logs -f backend-api
```
*(Gunakan `CTRL+C` pada terminal untuk berhenti melihat log).*

### 3. Menjalankan Load Test (Menggunakan k6)
Container load-tester dirancang untuk **tidak menyala otomatis**. Untuk men-trigger *test plan* (`dummy-load-test.js`), eksekusi perintah ini:
```bash
docker-compose --profile testing run --rm k6-loadtester run /scripts/dummy-load-test.js
```

### 4. Menghentikan Environment (Stop & Down)
Jika Anda selesai bekerja dan ingin mematikan container **tanpa menghilangkan apapun** (Volume dipertahankan):
```bash
docker-compose stop
```
Jika Anda ingin mematikan container dan menghapus jaringannya (Volume tetap utuh & aman):
```bash
docker-compose down
```

### 5. Memusnahkan Total Lingkungan (Clean Slate Mode)
> [!CAUTION]
> Perintah ini akan menghapus container, network, DAN SELURUH ISI DATABASE (Volume). Eksekusi hanya jika Anda ingin mereset state sistem 100% dari nol.
```bash
docker-compose down -v
```

*(Catatan Tambahan untuk tim: Tolong diskusikan atau review bersama jika Anda berniat memodifikasi infrastruktur inti di `docker-compose.yml`!).*

---

## 🏗️ Pembaruan Arsitektur & Progres Utama (Master-Replica)

Berdasarkan pemutakhiran infrastruktur terbaru, proyek ini telah mengadopsi standar rekayasa tingkat lanjut:

### 1. Pola Arsitektur Read/Write Separation (Master-Replica)
Dirancang khusus untuk melunasi target *Service Level Objective* (SLO) memangkas waktu respons hingga 30%.
- **Mekanisme**: Perintah modifikasi data (Insert/Update/Delete) ditangani eksklusif oleh **DB Utama (Master)**, sementara seluruh perintah pembacaan (Select) disalurkan otomatis ke **DB Cadangan (Replica)** untuk mengurangi latensi antrean operasi dasar.

### 2. Implementasi Database Routing Ganda (Java & Go)
Repositori ini memiliki dua mode pendekatan *backend* yang diisolasi dengan cermat:
- **Sisi Spring Boot (Java)**: Penyadapan otomatis di tingkat kode terjamin oleh adaptasi `AbstractRoutingDataSource` serta proteksi deteksi koneksi via `LazyConnectionDataSourceProxy` (`@Transactional(readOnly=true)`).
- **Sisi Golang (GORM)**: Dilengkapi distribusi `dbresolver` dengan kebijakan rute (*RandomPolicy*) untuk efisiensi instan berbasis Golang lokal.

### 3. Integrasi Endpoint JWT & Skema Independen
- File *bootstrap* relasional SQL (DDL) dirancang mumpuni mencakup *entity* fungsional inti: `users`, `accounts`, dan `transactions` dengan penyiapan optimasi Indeksasi.
- Konversi arsitektural pada tabel `sessions` untuk bergerak ringan (*Stateless*) dengan validasi gembok basis `jwt_jti`.

### 4. Jaringan Lingkungan Docker Terdistribusi Bebas
Struktur orkestrasi `docker-compose` sudah dikonfigurasikan agar adaptif menggantungkan interkonektivitas file `.env` di atas topologi `host.docker.internal`, memungkinkan para *developer* menyuntik langsung *Instance* PostgreSQL mandiri mereka dari sistem *Host* tanpa membongkar bongkahan *container* internal.

### 5. Panduan Transisi Sistem ke Database Eksternal (Fase Production)
Meskipun rancangan *Master-Replica* saat ini berjalan sangat memanjakan *developer* di dalam lingkup Docker untuk fase *Development* dan Uji Coba, arsitektur ini **WAJIB** memboyong (*export*) *layer* databasenya keluar menuju *Managed External Database* (seperti AWS RDS / Server Spesifik) bila menemui 3 pemicu berikut:
- **Menjelang Peluncuran (Go-Live)**: Membutuhkan jaminan keamanan regulasi (*Encryption at Rest*) & fitur keamanan data dari kemusnahan (*Point-in-Time Recovery*) yang sulit didapat di *Docker Volume*.
- **Kendala Botol Leher (I/O Throughput)**: Volume baca/tulis aplikasi (seperti target 30k TRX/jam) mulai membentur batas atas performa kecepatan virtualisasi disk bawaan Docker.
- **Tuntutan Keselamatan (Uptime 99.9%)**: Insulasi sistem. Bila server mesin Docker API hancur/padam, tumpukan *database* raksasa nasabah di server eksternal dipastikan tetap hidup bernapas tanpa terganggu rontoknya *container*.
**Langkah-Langkah Eksekusi Migrasi Menuju DB Eksternal:**
1. **Siapkan Kredensial (Environment)**: Salin file `.env.partner.template` menjadi sebuah file bernama `.env` di server.
2. **Kunci Target IP Publik**: Buka `.env`, lalu ganti *placeholder* IP `198.51.100.x` yang tertera pada `DB_MASTER_URL` dan `DB_REPLICA_URL` agar menarget lurus ke alamat IP/Domain mesin database raksasa spesifik milik Anda/mitra.
3. **Penyalaan Khusus Produksi**: Jangan pakai perintah compose yang biasa! Tanpa perlu membongkar-bongkar konfigurasi kode infrastruktur asli, langsung terbangkan ekosistem secara murni *API only* menggunakan file deklrasi yang kami pisahkan:
```bash
docker-compose -f docker-compose-external.yml up -d --build
```
Kontainer `backend-api` otomatis bangkit sendirian tanpa beban dan meluncurkan semua tembakan transaksi basis datanya eksklusif keluar menuju pusaran Awan (Cloud/External) yang Anda tuju!

---

## 🎯 Target Pekerjaan Selanjutnya (To-Do List Tim)

Saat ini, *Boilerplate* infrastruktur sudah menyala sempurna menggunakan kontainer *dummy/placeholders*. Berikut adalah hal yang perlu dilakukan oleh masing-masing PIC agar sistem ini menjadi *Real Application*:

- **Nadia (Backend API)**: 
  - Ganti isi folder `backend-api/` dengan *source code* Spring Boot aslinya. 
  - Timpa konfigurasi *dummy* (seperti `pom.xml`, `Application.java`, dan `application.yml`) dengan *Logic* aplikasi perbankan yang riil. Pastikan koneksi ke `postgres-db` dan `redis-cache` tersambung dengan benar.

- **Seva (Database)**: 
  - Ganti file `database-init/01-init-dummy.sql` dengan skema DDL (Data Definition Language) PostgreSQL yang sebenarnya.
  - Setiap kali ada *table* atau relasi baru, silakan tambahkan file bereksistensi `.sql` di folder tersebut (contoh: `02-insert-master-data.sql`).

- **Rafael (Load Testing)**:
  - Eksekusi skrip K6. Ubah isi `k6-scripts/dummy-load-test.js` dengan berbagai skenario stress testing/load testing (misal: simulasi 30.000 transaksi login/transfer per jam).
  - Anda bisa membuat file test `.js` baru jika skenarionya banyak, lalu sesuaikan argumen pemanggilannya di terminal.

- **Ego & Vanessa (Monitoring)**:
  - Buka *Dashboard* Grafana di `http://localhost:3000`.
  - Sambungkan *Data Source* ke Prometheus (`http://prometheus:9090`).
  - Buat *Dashboard* kustomisasi untuk memantau metrik dari Spring Boot Actuator, JVM memory, dan Database connection pool.


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
