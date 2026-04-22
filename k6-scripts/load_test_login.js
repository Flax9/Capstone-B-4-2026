import http from 'k6/http';
import { sleep, check } from 'k6';

// Uji Latensi: API Login
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

export default function () {
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

    const res = http.post('http://localhost:9000/api/auth/login', payload, params);

    check(res, {
        'status is 200': (r) => r.status === 200,
        'latency < 500ms': (r) => r.timings.duration < 500,
    });

    sleep(0.1); // Istirahat 100ms (10x lebih cepat menembak dari sebelumnya)
}
