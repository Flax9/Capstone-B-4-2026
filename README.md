# 🏦 Hyper-Scale gRPC Banking System

Arsitektur *microservices* perbankan berkinerja tinggi berbasis **gRPC** dan **Protocol Buffers**, dirancang untuk menangani ratusan ribu *request* per detik dengan latensi sub-detik. Seluruh infrastruktur berjalan sepenuhnya di dalam ekosistem **Docker**.

---

## 🏛️ Arsitektur Sistem

```
                          ┌─────────────────────┐
                          │   Client / K6 Load  │
                          │   Tester (gRPC)     │
                          └─────────┬───────────┘
                                    │
                          ┌─────────▼───────────┐
                          │  HAProxy (L4 TCP)   │
                          │  Load Balancer      │
                          │  IP: 172.25.0.100   │
                          └──┬──────┬──────┬────┘
                             │      │      │
               ┌─────────────┘      │      └─────────────┐
               ▼                    ▼                    ▼
      ┌────────────────┐  ┌────────────────┐  ┌────────────────┐
      │  Auth Service  │  │Balance Service │  │  Transaction   │
      │  (2 Replicas)  │  │ (3 Replicas)   │  │   Service      │
      │  Port 9001     │  │  Port 9002     │  │ (3 Replicas)   │
      └───┬────────────┘  └───┬────────────┘  │  Port 9003     │
          │                   │               └───┬────────────┘
          ▼                   ▼                   ▼
      ┌────────┐          ┌────────┐         ┌─────────┐
      │ Redis  │◄─────────│ Redis  │         │  Kafka  │
      │ Cache  │          │ Cache  │         │ Broker  │
      └────────┘          └────────┘         └──┬───┬──┘
                                                │   │
          ┌─────────────────────────────────────┘   │
          ▼                                         ▼
   ┌──────────────┐                          ┌────────────┐
   │  Transaction │                          │   Audit    │
   │    Worker    │                          │   Worker   │
   └──────┬───────┘                          └─────┬──────┘
          │                                        │
          ▼                                        ▼
   ┌─────────────┐    Streaming       ┌─────────────────┐
   │  PgBouncer  │───Replication─────▶│   PgBouncer     │
   │   Master    │                    │    Replica      │
   └──────┬──────┘                    └───────┬─────────┘
          ▼                                   ▼
   ┌─────────────┐                    ┌─────────────────┐
   │  PostgreSQL │                    │   PostgreSQL    │
   │   Master    │                    │    Replica      │
   └─────────────┘                    └─────────────────┘
```

### Alur Data Utama

1. **Client → HAProxy** — Request gRPC masuk dan didistribusikan ke replika service via *round-robin*.
2. **Auth/Balance → Redis** — Operasi baca menggunakan pola *Cache-Aside* (cek cache dulu, fallback ke DB).
3. **Transaction → Kafka** — Operasi tulis dipublikasikan sebagai event asinkron (*Fire & Forget*).
4. **Worker → PostgreSQL** — Worker mengonsumsi event dari Kafka dan menulis ke database Master.
5. **Prometheus → Service** — Metrik di-*scrape* setiap 15 detik dari port sidecar `:2112`.

---

## ⚡ Quick Start

> **Prasyarat:** [Docker Desktop](https://www.docker.com/products/docker-desktop/) sudah terinstal dan berjalan.

```bash
# 1. Clone repository
git clone https://github.com/Flax9/Capstone-B-4-2026.git
cd Capstone-B-4-2026

# 2. Bangun dan jalankan seluruh infrastruktur (20+ container)
docker-compose up -d --build

# 3. Tunggu ~30 detik agar semua health check selesai, lalu verifikasi
docker-compose ps

# 4. Jalankan load test
docker-compose --profile testing run --rm k6-loadtester run /scripts/load_test_combined.js
```

📖 **Panduan lengkap (step-by-step)** tersedia di [`PANDUAN_SETUP.md`](./PANDUAN_SETUP.md)

---

## 🔧 Teknologi yang Digunakan

| Kategori | Teknologi | Fungsi |
|----------|-----------|--------|
| **Bahasa** | Go (Golang) | Pengembangan seluruh microservice |
| **Protokol** | gRPC + Protocol Buffers | Komunikasi antar-service (HTTP/2 multiplexed) |
| **Load Balancer** | HAProxy (L4 TCP) | Distribusi traffic ke replika service |
| **Database** | PostgreSQL (Master-Replica) | Penyimpanan data persisten (CQRS) |
| **Connection Pool** | PgBouncer | Manajemen pool koneksi database |
| **Cache** | Redis | In-memory caching (Cache-Aside pattern) |
| **Message Broker** | Apache Kafka + Zookeeper | Event streaming untuk transaksi asinkron |
| **Monitoring** | Prometheus + Grafana | Scraping metrik & visualisasi dashboard |
| **Load Testing** | K6 (Grafana Labs) | Simulasi beban hingga 1 juta req/menit |
| **Infrastruktur** | Docker + Docker Compose | Kontainerisasi & orkestrasi |

---

## 📁 Struktur Direktori

```
Capstone-B-4-2026/
│
├── auth-service/              # Layanan Autentikasi (gRPC :9001)
│   ├── config/                #   Koneksi DB, Redis, Kafka
│   ├── handlers/              #   Logic handler Login
│   ├── models/                #   Model data User
│   ├── main.go                #   Entry point + Prometheus sidecar
│   └── Dockerfile
│
├── balance-service/           # Layanan Cek Saldo (gRPC :9002)
│   ├── config/                #   Koneksi DB, Redis
│   ├── handlers/              #   Logic handler CheckBalance
│   ├── main.go
│   └── Dockerfile
│
├── transaction-service/       # Layanan Transfer (gRPC :9003)
│   ├── config/                #   Koneksi DB, Kafka
│   ├── handlers/              #   Logic handler Transfer → Kafka
│   ├── main.go
│   └── Dockerfile
│
├── transaction-worker/        # Worker Pemroses Transaksi (Kafka Consumer)
│   └── main.go
│
├── audit-worker/              # Worker Pencatat Audit Log (Kafka Consumer)
│   └── main.go
│
├── proto/                     # Definisi Protocol Buffers
│   ├── auth/auth.proto
│   ├── balance/balance.proto
│   └── transaction/transaction.proto
│
├── haproxy/                   # Konfigurasi Load Balancer
│   └── haproxy.cfg            #   server-template + Docker DNS resolver
│
├── monitoring/                # Konfigurasi Observability
│   ├── prometheus.yml         #   Scrape config untuk semua service
│   └── custom_dashboard_v3_grpc.json  # Dashboard Grafana (P95 Latency)
│
├── k6-scripts/                # Skrip Load Testing
│   ├── load_test_combined.js  #   Test gabungan (Login+Saldo+Transfer)
│   ├── load_test_combined2.js #   Test gabungan (data acak)
│   ├── load_test_login.js     #   Test khusus Login
│   ├── load_test_cek_saldo.js #   Test khusus Cek Saldo
│   └── load_test_transfer.js  #   Test khusus Transfer
│
├── database-init/             # SQL Inisialisasi Database
│   └── 01-init.sql
│
├── docker-compose.yml         # Orkestrasi seluruh infrastruktur
├── PANDUAN_SETUP.md           # Panduan lengkap setup & testing
└── README.md                  # Dokumen ini
```

---

## 🚀 Layanan gRPC

### Auth Service (Port 9001 — 2 Replika)
Autentikasi pengguna dengan mekanisme *Cache-Aside* menggunakan Redis untuk menghindari *bottleneck* database saat lonjakan beban.

| RPC Method | Input | Output |
|-----------|-------|--------|
| `AuthService/Login` | `username`, `password` | Status code, JWT Token |

### Balance Service (Port 9002 — 3 Replika)
Pengambilan data saldo *real-time* dengan prioritas pembacaan dari Redis cache, *fallback* ke PostgreSQL Replica.

| RPC Method | Input | Output |
|-----------|-------|--------|
| `BalanceService/CheckBalance` | `userId` (UUID) | Saldo terkini, Status |

### Transaction Service (Port 9003 — 3 Replika)
Instruksi transfer dana secara asinkron. Service mempublikasikan event ke Kafka dan langsung merespons *Accepted (202)*.

| RPC Method | Input | Output |
|-----------|-------|--------|
| `TransactionService/Transfer` | `senderId`, `targetAccount`, `amount` | Konfirmasi (202 Accepted) |

---

## 📊 Monitoring & Dashboard

| Dashboard | URL | Kredensial |
|-----------|-----|-----------|
| **Grafana** | http://localhost:3000 | `admin` / `admin` |
| **Prometheus** | http://localhost:9090 | — |

### Setup Grafana
1. Tambahkan Data Source → Prometheus → URL: `http://prometheus:9090`
2. Import Dashboard → Upload file `monitoring/custom_dashboard_v3_grpc.json`

Panel yang tersedia:
- **gRPC Request Rate** — Request per detik (real-time)
- **gRPC P95 Latency** — Histogram latensi persentil ke-95
- **gRPC Error Rate** — Persentase kegagalan

---

## 🔥 Load Testing

Semua skrip K6 dijalankan di dalam container Docker — **tidak perlu instalasi lokal**.

```bash
# Test gabungan (semua fitur sekaligus) — REKOMENDASI
docker-compose --profile testing run --rm k6-loadtester run /scripts/load_test_combined.js

# Test dengan data acak (bypass cache)
docker-compose --profile testing run --rm k6-loadtester run /scripts/load_test_combined2.js
```

### Hasil Benchmark Terakhir (Laptop Lokal)

| Metrik | Nilai |
|--------|-------|
| **Success Rate** | 100.00% |
| **Total Iterasi (5 menit)** | 1.009.693 |
| **Throughput** | ~3.359 RPS |
| **Avg Latency** | 429 ms |
| **P95 Latency** | 974 ms |
| **Max VUs** | 1.500 |

> 💡 Target skrip adalah 1 juta req/menit. Throughput aktual dibatasi oleh kapasitas hardware laptop (bukan limitasi arsitektur). Untuk pengujian skala penuh, gunakan infrastruktur cloud.

---

## 🛡️ Fitur Arsitektur Utama

- **CQRS (Command Query Responsibility Segregation)** — Pemisahan jalur baca (Replica) dan tulis (Master) pada PostgreSQL untuk menghindari kontestasi *lock*.
- **Event-Driven Processing** — Transaksi mutasi diproses secara asinkron melalui Kafka, menghilangkan *write-blocking* pada database.
- **Cache-Aside Pattern** — Redis digunakan sebagai lapisan pembacaan cepat pada Auth dan Balance service (TTL 60 detik).
- **L4 Load Balancing** — HAProxy mendistribusikan traffic TCP secara merata ke semua replika tanpa overhead inspeksi HTTP.
- **Static IP Networking** — Subnet Docker tetap (`172.25.0.0/16`) dengan IP statis pada HAProxy untuk menghindari kegagalan DNS.
- **Connection Pooling** — PgBouncer mengelola pool koneksi ke PostgreSQL agar tidak terjadi *connection exhaustion*.
- **gRPC Histogram Metrics** — Setiap service mengekspos metrik latensi melalui sidecar HTTP (:2112) untuk di-*scrape* Prometheus.

---

## ⚠️ SOP Penggunaan Repositori

### 1. Branching Strategy
- **Branch `main`** adalah branch stabil/produksi. **Dilarang push langsung**.
- Buat branch baru dari `main` dengan format: `feature/nama-fitur`

```bash
git pull origin main
git checkout -b feature/nama-fitur
```

### 2. Pull Request (PR) Workflow
1. Push kode ke branch Anda: `git push origin feature/nama-fitur`
2. Buat **Pull Request** di GitHub ke arah `main`.
3. Koordinasikan review di grup tim.
4. Setelah di-approve, lakukan **merge**.

### 3. Batasan Penting
- ⛔ **Jangan edit `docker-compose.yml`** tanpa koordinasi dengan tim Infrastructure.
- ⛔ **Jangan commit file binary besar** (`.exe`, `.jar`) — sudah difilter `.gitignore`.

---

## 📝 Changelog

**[17 Mei 2026] — Optimalisasi Performa & Stabilitas Load Testing**
- 🔧 Implementasi **Redis Cache-Aside** pada `auth-service` — mengurangi beban SQL repetitif hingga 99%.
- 🌐 Konfigurasi **Static IP** (`172.25.0.100`) pada HAProxy untuk bypass DNS Docker.
- ⏱️ Penambahan **warm-up delay** (5 detik) pada K6 container untuk stabilitas jaringan.
- 📊 Pembaruan HAProxy menggunakan `server-template` + Docker DNS resolver untuk distribusi ke **semua replika**.

**[16 Mei 2026] — Infrastruktur Monitoring & Audit**
- 📡 Integrasi **Prometheus** dengan metrik gRPC histogram (P95 Latency).
- 📊 Pembuatan **Grafana Dashboard V3** khusus metrik gRPC.
- 🔍 Audit menyeluruh konfigurasi Docker untuk portabilitas antar-device.

**[06 Mei 2026] — Migrasi Arsitektur ke gRPC Microservices**
- 🚀 Migrasi dari REST API monolith (Fiber v2) ke **gRPC Microservices** (Auth, Balance, Transaction).
- ⚖️ Penggantian Envoy Proxy (L7) dengan **HAProxy (L4 TCP)** untuk efisiensi gRPC.
- 📨 Implementasi **Apache Kafka** untuk pemrosesan transaksi asinkron.
- 🔄 Penambahan **Transaction Worker** dan **Audit Worker** sebagai Kafka consumer.

---

*Terakhir diperbarui: 17 Mei 2026*
