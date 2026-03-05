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

### 1. Build & Start (Pertama Kali)
Gunakan perintah ini untuk mem-build *image* aplikasi Spring Boot dan menjalankan semua container (PostgreSQL, Redis, dll) di latar belakang:
```bash
docker-compose up -d --build
```
> [!NOTE]
> *Tunggu beberapa saat pada run pertama. Backend API baru akan menyala **setelah** PostgreSQL dan Redis berstatus `healthy`.*

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
