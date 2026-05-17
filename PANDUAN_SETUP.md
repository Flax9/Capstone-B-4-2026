# 🚀 Panduan Setup & Pengujian — Hyper-Scale gRPC Banking System

Panduan ini menjelaskan langkah-langkah **dari awal hingga load testing** untuk proyek arsitektur microservices perbankan berbasis gRPC. Ikuti setiap tahap secara berurutan.

---

## 📋 Daftar Isi

1. [Prasyarat](#1--prasyarat)
2. [Clone Repository](#2--clone-repository)
3. [Konfigurasi Environment](#3--konfigurasi-environment)
4. [Menjalankan Seluruh Infrastruktur](#4--menjalankan-seluruh-infrastruktur)
5. [Verifikasi Sistem](#5--verifikasi-sistem)
6. [Mengakses Dashboard Monitoring](#6--mengakses-dashboard-monitoring)
7. [Menjalankan Load Testing](#7--menjalankan-load-testing)
8. [Membaca Hasil Load Test](#8--membaca-hasil-load-test)
9. [Menghentikan Sistem](#9--menghentikan-sistem)
10. [Troubleshooting](#10--troubleshooting)
11. [Arsitektur Sistem](#11--arsitektur-sistem)

---

## 1. 📦 Prasyarat

Pastikan perangkat Anda sudah terinstal software berikut **sebelum memulai**:

| Software | Versi Minimum | Fungsi | Link Download |
|----------|---------------|--------|---------------|
| **Docker Desktop** | v4.25+ | Menjalankan seluruh infrastruktur dalam container | [docker.com/products/docker-desktop](https://www.docker.com/products/docker-desktop/) |
| **Git** | v2.40+ | Meng-clone repository proyek | [git-scm.com](https://git-scm.com/) |

> ⚠️ **PENTING:** Pastikan **Docker Desktop sudah berjalan (running)** sebelum melanjutkan ke langkah berikutnya. Anda bisa mengeceknya dengan membuka terminal dan menjalankan:
> ```bash
> docker --version
> docker-compose --version
> ```

### Rekomendasi Spesifikasi Hardware

| Komponen | Minimum | Rekomendasi |
|----------|---------|-------------|
| **RAM** | 8 GB | 32 GB |
| **CPU** | 6 Core | 16 Core |
| **Disk** | 10 GB kosong | 20 GB kosong |
| **OS** | Windows 10/11, macOS, Linux | Windows 11 + WSL2 |

> 💡 **Tips Docker Desktop (Windows):** Buka *Settings > Resources* dan alokasikan minimal **6 GB RAM** dan **4 CPU** untuk Docker agar sistem berjalan optimal.

---

## 2. 📥 Clone Repository

Buka terminal (PowerShell / Command Prompt / Terminal) dan jalankan:

```bash
# Clone repository
git clone https://github.com/Flax9/Capstone-B-4-2026.git capstone-grpc

# Masuk ke direktori proyek
cd capstone-grpc
```

### Struktur Direktori Proyek

Setelah clone, Anda akan melihat struktur berikut:

```
capstone-grpc/
├── auth-service/           # Layanan Autentikasi (gRPC, Port 9001)
├── balance-service/        # Layanan Saldo (gRPC, Port 9002)
├── transaction-service/    # Layanan Transaksi (gRPC, Port 9003)
├── transaction-worker/     # Worker Pemroses Transaksi (Kafka Consumer)
├── audit-worker/           # Worker Pencatat Audit Log (Kafka Consumer)
├── proto/                  # Definisi Protocol Buffers (.proto)
├── haproxy/                # Konfigurasi Load Balancer L4
│   └── haproxy.cfg
├── monitoring/             # Konfigurasi Prometheus & Dashboard Grafana
│   ├── prometheus.yml
│   └── custom_dashboard_v3_grpc.json
├── k6-scripts/             # Skrip Load Testing
│   ├── load_test_combined.js    # Load test utama (semua fitur)
│   ├── load_test_combined2.js   # Load test dengan data acak
│   ├── load_test_login.js       # Load test khusus login
│   ├── load_test_cek_saldo.js   # Load test khusus cek saldo
│   └── load_test_transfer.js    # Load test khusus transfer
├── database-init/          # SQL inisialisasi database
│   └── 01-init.sql
├── docker-compose.yml      # Orkestrasi seluruh infrastruktur
├── .env.development.template  # Template environment variables
└── PANDUAN_SETUP.md        # 📖 File ini
```

---

## 3. ⚙️ Konfigurasi Environment

Salin file template environment ke `.env`:

```bash
# Windows (PowerShell)
Copy-Item .env.development.template .env

# Linux / macOS
cp .env.development.template .env
```

> 📌 File `.env` berisi konfigurasi koneksi database untuk pengembangan lokal. Untuk menjalankan **seluruh sistem di Docker** (yang kita lakukan di panduan ini), file ini bersifat opsional karena semua variabel sudah didefinisikan di dalam `docker-compose.yml`.

---

## 4. 🏗️ Menjalankan Seluruh Infrastruktur

Ini adalah langkah utama. **Satu perintah ini akan membangun dan menjalankan 20+ container** secara otomatis:

```bash
docker-compose up -d --build
```

### Apa yang terjadi di balik layar?

Perintah tersebut akan:
1. **Membangun (Build)** image Docker untuk: `auth-service`, `balance-service`, `transaction-service`, `transaction-worker`, dan `audit-worker`.
2. **Mengunduh (Pull)** image publik untuk: PostgreSQL, Redis, Kafka, Zookeeper, PgBouncer, HAProxy, Prometheus, Grafana, dan InfluxDB.
3. **Menyalakan** seluruh container dengan urutan dependensi yang benar.
4. **Menginisialisasi** skema database secara otomatis melalui `database-init/01-init.sql`.

### Estimasi Waktu

| Kondisi | Waktu |
|---------|-------|
| **Pertama kali** (download semua image) | 5 - 15 menit |
| **Selanjutnya** (image sudah ada) | 1 - 3 menit |

> ⏳ Setelah perintah selesai, tunggu **±30 detik** tambahan agar semua health check selesai (terutama Kafka dan PostgreSQL).

---

## 5. ✅ Verifikasi Sistem

Pastikan semua container berjalan dengan baik:

```bash
docker-compose ps
```

Anda harus melihat **semua container** berstatus `Up` atau `Running`. Berikut daftar container yang diharapkan:

| Container | Status yang Diharapkan | Fungsi |
|-----------|----------------------|--------|
| `capstone-postgres-master-1` | Up (healthy) | Database utama (Write) |
| `capstone-postgres-replica-1` | Up (healthy) | Database replika (Read) |
| `capstone-redis-cache-1` | Up (healthy) | In-memory cache |
| `capstone-zookeeper-1` | Up | Koordinator Kafka |
| `capstone-kafka-1` | Up (healthy) | Message broker |
| `capstone-pgbouncer-master-1` | Up (healthy) | Connection pooler (Master) |
| `capstone-pgbouncer-replica-1` | Up (healthy) | Connection pooler (Replica) |
| `capstone-auth-service-1` | Up | Layanan autentikasi (Replica 1) |
| `capstone-auth-service-2` | Up | Layanan autentikasi (Replica 2) |
| `capstone-balance-service-1` | Up | Layanan saldo (Replica 1) |
| `capstone-balance-service-2` | Up | Layanan saldo (Replica 2) |
| `capstone-balance-service-3` | Up | Layanan saldo (Replica 3) |
| `capstone-transaction-service-1` | Up | Layanan transaksi (Replica 1) |
| `capstone-transaction-service-2` | Up | Layanan transaksi (Replica 2) |
| `capstone-transaction-service-3` | Up | Layanan transaksi (Replica 3) |
| `capstone-transaction-worker-1` | Up | Worker transaksi (Instance 1) |
| `capstone-transaction-worker-2` | Up | Worker transaksi (Instance 2) |
| `capstone-audit-worker-1` | Up | Worker audit log |
| `capstone-haproxy-1` | Up | Load balancer L4 |
| `capstone-prometheus-1` | Up | Metrics scraper |
| `capstone-grafana-1` | Up | Dashboard monitoring |
| `capstone-influxdb-1` | Up | K6 metrics storage |

### Verifikasi Cepat via Log

Jika ada container yang tidak muncul atau statusnya `Restarting`, cek log-nya:

```bash
# Contoh: cek log auth-service
docker logs capstone-auth-service-1 --tail 20

# Contoh: cek log haproxy
docker logs capstone-haproxy-1 --tail 10
```

Log yang sehat untuk HAProxy akan menampilkan:
```
Server auth_back/auth1 ('auth-service') is UP/READY (resolves again).
Server auth_back/auth2 ('auth-service') is UP/READY (resolves again).
Server balance_back/balance1 ('balance-service') is UP/READY (resolves again).
...
```

---

## 6. 📊 Mengakses Dashboard Monitoring

Setelah sistem berjalan, Anda bisa mengakses dashboard berikut melalui browser:

| Dashboard | URL | Kredensial |
|-----------|-----|-----------|
| **Grafana** | [http://localhost:3000](http://localhost:3000) | `admin` / `admin` (ubah saat login pertama) |
| **Prometheus** | [http://localhost:9090](http://localhost:9090) | Tidak ada |

### Setup Grafana (Pertama Kali)

#### Langkah 1: Tambahkan Data Source Prometheus

1. Buka Grafana di [http://localhost:3000](http://localhost:3000).
2. Login dengan `admin` / `admin`.
3. Pergi ke **Connections > Data sources > Add data source**.
4. Pilih **Prometheus**.
5. Pada field **Prometheus server URL**, isi: `http://prometheus:9090`
6. Scroll ke bawah dan klik **Save & Test**.
7. Pastikan muncul pesan hijau: *"Successfully queried the Prometheus API."*

#### Langkah 2: Import Dashboard gRPC

1. Pergi ke **Dashboards > New > Import**.
2. Klik **Upload JSON file**.
3. Pilih file: `monitoring/custom_dashboard_v3_grpc.json`
4. Pada dropdown **Prometheus**, pilih data source yang baru saja ditambahkan.
5. Klik **Import**.

Anda akan melihat dashboard dengan panel-panel:
- **gRPC Request Rate** (request/detik)
- **gRPC P95 Latency** (histogram latensi persentil ke-95)
- **gRPC Error Rate** (persentase error)

> 💡 Data akan mulai muncul di dashboard setelah Anda menjalankan load test di langkah berikutnya.

---

## 7. 🔥 Menjalankan Load Testing

Proyek ini menyediakan beberapa skrip pengujian beban menggunakan **K6 (Grafana Labs)**. Semua skrip dijalankan di dalam container Docker, sehingga Anda **tidak perlu menginstal K6 secara lokal**.

### Skrip yang Tersedia

| Skrip | Deskripsi | Target |
|-------|-----------|--------|
| `load_test_combined.js` | Menguji **semua fitur** (Login, Cek Saldo, Transfer) secara bersamaan | 1 juta req/menit |
| `load_test_combined2.js` | Sama seperti di atas, tetapi dengan **data acak** (menguji tanpa cache) | 1 juta req/menit |
| `load_test_login.js` | Pengujian khusus **Login / Autentikasi** | Tergantung konfigurasi |
| `load_test_cek_saldo.js` | Pengujian khusus **Cek Saldo** | Tergantung konfigurasi |
| `load_test_transfer.js` | Pengujian khusus **Transfer Dana** | Tergantung konfigurasi |

### Menjalankan Load Test

```bash
# Load Test Gabungan (Semua Fitur) — REKOMENDASI UTAMA
docker-compose --profile testing run --rm k6-loadtester run /scripts/load_test_combined.js

# Load Test dengan Data Acak (Tanpa Cache)
docker-compose --profile testing run --rm k6-loadtester run /scripts/load_test_combined2.js

# Load Test Khusus Login
docker-compose --profile testing run --rm k6-loadtester run /scripts/load_test_login.js

# Load Test Khusus Cek Saldo
docker-compose --profile testing run --rm k6-loadtester run /scripts/load_test_cek_saldo.js

# Load Test Khusus Transfer
docker-compose --profile testing run --rm k6-loadtester run /scripts/load_test_transfer.js
```

> 📌 **Catatan:** Perintah di atas akan otomatis menunggu 5 detik (*warm-up delay*) sebelum memulai test agar jaringan Docker sudah stabil.

### Menghentikan Load Test (Jika Diperlukan)

Tekan `Ctrl + C` untuk menghentikan load test kapanpun. K6 akan menampilkan hasil akhir meskipun test dihentikan lebih awal.

---

## 8. 📈 Membaca Hasil Load Test

Setelah test selesai, K6 akan menampilkan laporan seperti ini:

```
█ TOTAL RESULTS

  checks_total.......: 1009693  3359.342412/s
  checks_succeeded...: 100.00%  1009693 out of 1009693
  checks_failed......: 0.00%    0 out of 1009693

  ✓ transfer OK
  ✓ login OK
  ✓ cek_saldo OK

  GRPC
  grpc_req_duration....: avg=429.61ms min=284.82µs med=381.42ms max=3.68s p(90)=789.8ms p(95)=974.18ms
```

### Cara Membaca Metrik Penting

| Metrik | Penjelasan | Target Ideal |
|--------|-----------|--------------|
| **checks_succeeded** | Persentase request yang berhasil divalidasi. **Ini metrik terpenting.** | ≥ 99% |
| **iterations** | Total jumlah request yang benar-benar dieksekusi oleh sistem. | Semakin tinggi semakin baik |
| **dropped_iterations** | Request yang tidak sempat dieksekusi karena keterbatasan VU/hardware. Bukan kegagalan sistem. | Wajar jika ada di laptop |
| **grpc_req_duration (avg)** | Rata-rata waktu respons per request. | < 1 detik |
| **grpc_req_duration (p95)** | 95% request selesai dalam waktu ini. **Metrik standar industri (SLA).** | < 2 detik |
| **vus_max** | Jumlah Virtual Users (koneksi simultan) maksimal yang terpakai. | Tergantung `maxVUs` |

### Memahami "Dropped Iterations"

Jika Anda melihat angka `dropped_iterations` yang besar, **jangan khawatir** — ini bukan kegagalan. Ini berarti:
- K6 diminta menembak 1 juta request/menit (target ambisius).
- Laptop Anda hanya mampu memproses sebagian dari target tersebut.
- Request yang tidak sempat dikirim dimasukkan ke kategori "dropped".
- Sistem backend tetap **100% sukses** untuk setiap request yang berhasil dikirim.

> 💡 Untuk mencapai target 1 juta request/menit secara penuh, diperlukan infrastruktur cloud dengan spesifikasi produksi (Distributed Load Testing).

---

## 9. 🛑 Menghentikan Sistem

```bash
# Matikan semua container (data tetap tersimpan di volume)
docker-compose down

# Matikan dan HAPUS semua data (database, cache, metrics) — fresh start
docker-compose down -v
```

> ⚠️ Opsi `-v` akan **menghapus semua data** termasuk isi database dan metrics Grafana. Gunakan hanya jika Anda ingin memulai dari awal.

---

## 10. 🔧 Troubleshooting

### Masalah Umum & Solusi

| Masalah | Penyebab | Solusi |
|---------|----------|--------|
| Container terus *Restarting* | Database belum siap saat service mulai | Tunggu 30-60 detik, lalu jalankan `docker-compose up -d` lagi |
| `lookup haproxy: no such host` | DNS Docker belum stabil | Sudah diatasi dengan Static IP. Pastikan Anda menggunakan versi terbaru `docker-compose.yml` |
| `Port already in use` | Port sudah dipakai aplikasi lain | Matikan aplikasi yang menggunakan port tersebut, atau ubah port mapping di `docker-compose.yml` |
| K6 error `context deadline exceeded` | VU terlalu banyak untuk kapasitas laptop | Turunkan `maxVUs` di file skrip K6 (misal: dari 10000 ke 1500) |
| Grafana tidak menampilkan data | Data source belum dikonfigurasi | Ikuti langkah setup di [Bagian 6](#6--mengakses-dashboard-monitoring) |
| Build gagal / image error | Cache Docker korup | Jalankan `docker-compose build --no-cache` lalu `docker-compose up -d` |

### Port yang Digunakan

Pastikan port-port berikut **tidak digunakan** oleh aplikasi lain di komputer Anda:

| Port | Service |
|------|---------|
| 3000 | Grafana |
| 8086 | InfluxDB |
| 9001 | HAProxy → Auth Service |
| 9002 | HAProxy → Balance Service |
| 9003 | HAProxy → Transaction Service |
| 9090 | Prometheus |
| 15432 | PostgreSQL Master |

---

## 11. 🏛️ Arsitektur Sistem

```
┌─────────────────────────────────────────────────────────────┐
│                    EXTERNAL LAYER                           │
│   ┌──────────────┐      ┌──────────────────────────┐        │
│   │  Mobile App  │      │  K6 Load Tester          │        │
│   │  (gRPC)      │      │  (1M req/min target)     │        │
│   └──────┬───────┘      └────────────┬─────────────┘        │
└──────────┼───────────────────────────┼──────────────────────┘
           │         gRPC (HTTP/2)     │
           ▼                           ▼
┌─────────────────────────────────────────────────────────────┐
│              GATEWAY LAYER — HAProxy (L4 TCP)               │
│                    IP: 172.25.0.100                          │
│              Round-Robin Load Balancing                      │
└────────┬──────────────────┬─────────────────┬───────────────┘
         │                  │                 │
         ▼                  ▼                 ▼
┌────────────────┐ ┌────────────────┐ ┌────────────────┐
│  Auth Service  │ │Balance Service │ │ Trx Service    │
│  (2 Replicas)  │ │ (3 Replicas)   │ │ (3 Replicas)   │
│  Port 9001     │ │  Port 9002     │ │  Port 9003     │
└───┬────┬───────┘ └───┬────┬───────┘ └───────┬────────┘
    │    │             │    │                  │
    │    ▼             │    ▼                  ▼
    │  ┌──────┐        │  ┌──────┐     ┌─────────────┐
    │  │Redis │◄───────┘  │Redis │     │    Kafka     │
    │  │Cache │           │Cache │     │   (Broker)   │
    │  └──────┘           └──────┘     └──┬──────┬────┘
    │                                     │      │
    ▼                                     ▼      ▼
┌──────────┐                    ┌──────────┐ ┌────────┐
│PgBouncer │                    │Trx Worker│ │ Audit  │
│ Replica  │                    │(Consumer)│ │ Worker │
└────┬─────┘                    └────┬─────┘ └───┬────┘
     ▼                               │           │
┌──────────┐                         ▼           ▼
│ PG Read  │◄──── Streaming ───┌──────────┐◄─────┘
│ (Replica)│    Replication    │ PG Write │
└──────────┘                   │ (Master) │
                               └──────────┘
```

### Alur Data

1. **Client → HAProxy:** Request gRPC masuk dan didistribusikan ke replika service.
2. **Service → Redis:** Operasi baca (Login, Saldo) dicek di cache terlebih dahulu.
3. **Service → PgBouncer → PostgreSQL:** Jika cache kosong, query diteruskan ke database.
4. **Service → Kafka:** Operasi tulis (Transfer) dipublikasikan sebagai event asinkron.
5. **Worker → PostgreSQL:** Worker mengonsumsi event dari Kafka dan menulis ke database.
6. **Prometheus → Service:** Metrics discrape setiap 15 detik dari port 2112.
7. **Grafana → Prometheus:** Dashboard memvisualisasikan metrics secara real-time.

---

## 📝 Catatan Akhir

- Seluruh infrastruktur berjalan **100% di dalam Docker**. Tidak perlu menginstal Go, PostgreSQL, Redis, atau Kafka secara lokal.
- Sistem menggunakan **Static IP** (`172.25.0.0/16` subnet) untuk menghindari masalah DNS Docker.
- Jika mengalami masalah, cek bagian [Troubleshooting](#10--troubleshooting) atau hubungi tim pengembang.

---

*Dokumen ini terakhir diperbarui pada: Mei 2026*
