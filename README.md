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

### TIER 1: Model Development Murni / IDE 💻
Di mode ini, *developer* (Backend Java/Go) men-debug kodingannya langsung melalui jendela *Run Button* IDE favorit (VSCode/IntelliJ) tanpa melempar Node API ke dalam Docker, namun mendelegasikan beban *databasenya* ke kontainer lokal.
1. Salin draf `.env.development.template` menjadi sebuah file utuh bernama `.env`.
2. Anda akan mendapati letak koneksi *Master* terkunci di `localhost:5432` dan *Replica* di `localhost:5433` pada laptop Anda.
3. Pancing mesin databasenya secara independen dengan eksekusi:
   ```bash
   docker-compose up -d postgres-master postgres-replica redis-cache
   ```

### TIER 2: Model Uji Coba Terintegrasi (Docker Cluster) 🐳
Apabila fitur *Backend* sudah rampung, Anda wajib mengetes interaksi aplikasi secara utuh layaknya lingkungan korporat yang kokoh menempel pada *Virtual Network* Docker lokal (Mencakup K6 *Load Tester* & Prometheus Grafana).
1. Pastikan seluruh relasi dan DDL termutakhir di `01-init.sql`.
2. Peluncuran Klaster Menyeluruh (API + DB + Redis + Tools):
   ```bash
   docker-compose up -d --build
   ```

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
