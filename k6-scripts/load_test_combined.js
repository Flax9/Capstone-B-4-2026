import http from 'k6/http';
import { sleep, check } from 'k6';

// Uji Latensi Gabungan: API Login, Cek Saldo, dan Transfer
// Target: mensimulasikan total 1.000.000 iterasi yang dipilih secara random
export const options = {
    scenarios: {
        uji_beban_gabungan: {
            // Menggunakan executor khusus untuk mematok kecepatan (Rate) yang konstan
            executor: 'constant-arrival-rate',
            rate: 1000000,          // Target: 1.000.000 request
            timeUnit: '1m',         // Dalam waktu: 1 menit
            duration: '1m',         // Durasi tes berjalan selama 1 menit
            preAllocatedVUs: 3000,  // Pasukan awal dinaikkan dua kali lipat
            maxVUs: 10000,          // Pasukan cadangan dinaikkan mentok ke 10.000 VUs
        },
    },
};

const BASE_URL = 'http://localhost:9000';

// --- SETUP: Login sekali untuk mendapatkan token JWT untuk request yang butuh auth ---
export function setup() {
    const payload = JSON.stringify({ username: 'nasabah_01', password: 'rahasia' });
    const params = {
        headers: {
            'Content-Type': 'application/json',
            'X-Test-Bypass': 'b7fc809a-super-secret-key-capstone'
        }
    };
    const res = http.post(`${BASE_URL}/api/auth/login`, payload, params);

    let token = '';
    if (res.status === 200) {
        token = res.json('token');
    } else {
        console.error('Setup failed: Login error');
    }

    return { token: token };
}

export default function (data) {
    // Generate angka random antara 0 dan 1
    const rand = Math.random();

    if (rand < 0.33) {
        // 1. API Login
        const payload = JSON.stringify({
            username: 'nasabah_01',
            password: 'rahasia',
        });

        const params = {
            headers: {
                'Content-Type': 'application/json',
                'X-Test-Bypass': 'b7fc809a-super-secret-key-capstone'
            },
        };

        const res = http.post(`${BASE_URL}/api/auth/login`, payload, params);

        check(res, {
            'login status is 200': (r) => r.status === 200,
            'login latency < 500ms': (r) => r.timings.duration < 500,
        });

    } else if (rand < 0.66) {
        // 2. API Cek Saldo
        const accountId = '924de2cf-e950-4f92-8e37-ae2eb7dda7e5';

        const params = {
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${data.token}`,
                'X-Test-Bypass': 'b7fc809a-super-secret-key-capstone',
            },
        };

        const res = http.get(`${BASE_URL}/api/accounts/${accountId}`, params);

        check(res, {
            'cek_saldo status is 200': (r) => r.status === 200,
            'cek_saldo latency < 200ms': (r) => r.timings.duration < 200,
        });

    } else {
        // 3. API Transfer
        const payload = JSON.stringify({
            from_account_id: '924de2cf-e950-4f92-8e37-ae2eb7dda7e5',
            to_account_id: 'e3acd2bc-94d1-475e-ac7a-12fe405ad426',
            amount: 10, // Nominal lebih kecil agar saldo bertahan lebih lama saat load test besar
        });

        const params = {
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${data.token}`,
                'X-Test-Bypass': 'b7fc809a-super-secret-key-capstone',
            },
        };

        const res = http.post(`${BASE_URL}/api/transactions/transfer`, payload, params);

        check(res, {
            'transfer status is 202': (r) => r.status === 202,
            'transfer latency < 1000ms': (r) => r.timings.duration < 1000,
        });
    }

    sleep(0.1);
}
