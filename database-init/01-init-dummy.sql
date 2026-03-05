-- File SQL dummy untuk memberikan contoh inisiasi awal bagi Seva
-- Semua file berekstensi .sql di dalam folder ini otomatis dieksekusi oleh container Postgres saat pertama instalasi

CREATE TABLE IF NOT EXISTS dummy_users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO dummy_users (name) VALUES ('System Base Init');
