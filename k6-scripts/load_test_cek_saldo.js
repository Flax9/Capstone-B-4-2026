import http from 'k6/http';
import { sleep, check } from 'k6';

// Uji Latensi: API Cek Saldo (Read - via Redis Cache)
// Target: mensimulasikan 100 user concurrent selama 1 menit
export const options = {
    vus: 100,
    duration: '1m',
};

const BASE_URL = 'http://host.docker.internal:9000/api/accounts/${accountId}'
// --- SETUP: Login sekali untuk mendapatkan token JWT ---
export function setup() {
    const payload = JSON.stringify({ username: 'nasabah_01', password: 'rahasia' });
    const params = { headers: { 'Content-Type': 'application/json' } };
    const res = http.post(`${BASE_URL}api/auth/login`, payload, params);
    console.log(res.status);
    console.log(res.body);
    
    const token = res.json('token');
    return { token: token };
}

export default function (data) {
    const accountId = '924de2cf-e950-4f92-8e37-ae2eb7dda7e5'; // UUID Akun Target

    const params = {
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${data.token}`,
        },
    };

    const res = http.get(`${BASE_URL}api/accounts/${accountId}`, params);

    check(res, {
        'status is 200': (r) => r.status === 200,
        'latency < 200ms': (r) => r.timings.duration < 200, // Target lebih ketat karena ada Redis Cache
    });

    sleep(1);
}
