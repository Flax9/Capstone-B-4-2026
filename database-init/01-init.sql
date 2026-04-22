-- KATEGORI 1 DDL : TABEL MASTER (Data Statis/Jarang Berubah)
-- Digunakan untuk validasi login dan informasi rekening dasar.

CREATE TABLE IF NOT EXISTS users (
    user_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    full_name VARCHAR(100) NOT NULL,
    token_version INT DEFAULT 1, -- Untuk mendukung strategi pembatalan JWT
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS accounts (
    account_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(user_id),
    account_number VARCHAR(20) UNIQUE NOT NULL,
    balance NUMERIC(15, 2) DEFAULT 0.00, -- Data ini akan di-cache di Redis
    currency VARCHAR(3) DEFAULT 'IDR',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- KATEGORI 2: TABEL TRANSAKSI (Data Dinamis/Volume Tinggi)
-- Fokus utama penulisan asinkron dari Message Queue.

CREATE TABLE IF NOT EXISTS transactions (
    transaction_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reference_number VARCHAR(50) UNIQUE NOT NULL,
    from_account_id UUID REFERENCES accounts(account_id),
    to_account_id UUID REFERENCES accounts(account_id),
    amount NUMERIC(15, 2) NOT NULL,
    transaction_type VARCHAR(20) NOT NULL, -- Contoh: 'TRANSFER', 'PAYMENT'
    status VARCHAR(20) DEFAULT 'PENDING', -- PENDING, SUCCESS, FAILED
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS audit_logs (
    log_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID, -- Tidak di-RESTRICT dengan REFERENCES agar bisa melacak user yang gagal login (user tidak ditemukan)
    action VARCHAR(50) NOT NULL, -- Contoh: 'LOGIN_SUCCESS', 'LOGIN_FAILED', 'TRANSFER_INITIATED'
    ip_address VARCHAR(45), -- Mendukung IPv4 dan IPv6
    user_agent TEXT, -- Info perangkat/browser nasabah
    details JSONB, -- Sangat berguna di Postgres untuk menyimpan payload dinamis (misal: nominal transfer, device ID)
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- INDEXING UNTUK PERFORMA (Mengatasi Latensi saat Peak Load)
CREATE INDEX IF NOT EXISTS idx_accounts_user_id ON accounts(user_id);
CREATE INDEX IF NOT EXISTS idx_transactions_from_account ON transactions(from_account_id);
CREATE INDEX IF NOT EXISTS idx_transactions_created_at ON transactions(created_at);

-- INDEXING UNTUK OBSERVABILITY & AUDIT
-- Sangat penting agar tim SRE/Admin bisa mencari log dengan cepat tanpa membebani database
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at);
-- Index khusus untuk mencari data di dalam JSONB (opsional tapi disarankan untuk scale besar)
CREATE INDEX IF NOT EXISTS idx_audit_logs_details ON audit_logs USING GIN (details);


--------------------------------
-- KATEGORI 2 DML :

-- Menambah User Master (Contoh Nasabah)
-- ON CONFLICT (username) DO NOTHING ditambahkan agar terhindar dari Error "Duplicate Key" jika dijalankan ulang (Idempotent)
INSERT INTO users (username, password_hash, full_name) VALUES 
('nasabah_01', 'hashed_password_here', 'Ego Lanang Jagad'),
('nasabah_02', 'hashed_password_here', 'Vanessa Dyah')
ON CONFLICT (username) DO NOTHING;

-- Menambah Data Rekening Master
-- Sengaja dijahit (Hardcoded) UUID secara paksa untuk dicocokkan dengan skrip K6 Rafael
INSERT INTO accounts (account_id, user_id, account_number, balance) VALUES 
('924de2cf-e950-4f92-8e37-ae2eb7dda7e5', (SELECT user_id FROM users WHERE username = 'nasabah_01'), '1234567890', 5000000.00),
('e3acd2bc-94d1-475e-ac7a-12fe405ad426', (SELECT user_id FROM users WHERE username = 'nasabah_02'), '0987654321', 1000000.00)
ON CONFLICT (account_number) DO NOTHING;

-- Menambah Contoh Data Transaksi Awal
INSERT INTO transactions (reference_number, from_account_id, to_account_id, amount, transaction_type, status) VALUES 
('REF-2026-0001', 
 (SELECT account_id FROM accounts WHERE account_number = '1234567890'), 
 (SELECT account_id FROM accounts WHERE account_number = '0987654321'), 
 150000.00, 'TRANSFER', 'SUCCESS')
ON CONFLICT (reference_number) DO NOTHING;

-- Contoh 1: Log saat nasabah berhasil login dan mendapatkan JWT
INSERT INTO audit_logs (user_id, action, ip_address, user_agent, details) 
VALUES (
    (SELECT user_id FROM users WHERE username = 'nasabah_01'), 
    'LOGIN_SUCCESS', 
    '192.168.1.15', 
    'Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X)', 
    '{"auth_method": "password", "token_issued": true}'::jsonb
);

-- Contoh 2: Log saat transaksi transfer dipicu (sebelum masuk antrean/Message Queue)
INSERT INTO audit_logs (user_id, action, ip_address, user_agent, details) 
VALUES (
    (SELECT user_id FROM users WHERE username = 'nasabah_01'), 
    'TRANSFER_INITIATED', 
    '192.168.1.15', 
    'MobileApp/v2.1.0 (Android 13)', 
    '{"reference_number": "REF-2026-0001", "amount": 150000.00, "to_account": "0987654321"}'::jsonb
);

-- Contoh 3: Log anomali (gagal login karena password salah / brute force)
INSERT INTO audit_logs (user_id, action, ip_address, user_agent, details) 
VALUES (
    (SELECT user_id FROM users WHERE username = 'nasabah_02'), 
    'LOGIN_FAILED', 
    '114.125.10.22', 
    'Unknown Script/1.0', 
    '{"reason": "invalid_password", "attempt_count": 3}'::jsonb
);
