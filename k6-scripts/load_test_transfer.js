import http from 'k6/http';
import { sleep, check } from 'k6';

// Uji Latensi: API Transfer (Write - via PostgreSQL Master)
// Target: mensimulasikan 100 user concurrent selama 1 menit
export const options = {
    scenarios: {
        uji_beban: {
            executor: 'shared-iterations',
            vus: 300,
            iterations: 300000,
            maxDuration: '30m',
        },
    },
};

// --- SETUP: Login sekali untuk mendapatkan token JWT ---
export function setup() {
    const payload = JSON.stringify({ username: 'nasabah_01', password: 'rahasia' });
    const params = { 
        headers: { 
            'Content-Type': 'application/json',
            'X-Test-Bypass': 'b7fc809a-super-secret-key-capstone'
        } 
    };
    const res = http.post('http://localhost:9000/api/auth/login', payload, params);
    const token = res.json('token');
    return { token: token };
}

export default function (data) {
    const payload = JSON.stringify({
        from_account_id: '924de2cf-e950-4f92-8e37-ae2eb7dda7e5', // UUID Pengirim
        to_account_id:   'e3acd2bc-94d1-475e-ac7a-12fe405ad426', // UUID Penerima
        amount:          100,                                     // Nominal kecil agar saldo tidak habis
    });

    const params = {
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${data.token}`,
            'X-Test-Bypass': 'b7fc809a-super-secret-key-capstone',
        },
    };

    const res = http.post('http://localhost:9000/api/transactions/transfer', payload, params);

    check(res, {
        'status is 200': (r) => r.status === 200,
        'latency < 1000ms': (r) => r.timings.duration < 1000, // Target lebih longgar karena Write ke DB
    });

    sleep(0.1);
}
