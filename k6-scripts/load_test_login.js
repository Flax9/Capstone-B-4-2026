import http from 'k6/http';
import { sleep, check } from 'k6';

// Uji Latensi: API Login
// Target: mensimulasikan 100 user concurrent selama 1 menit
export const options = {
    vus: 100,
    duration: '1m',
};

export default function () {
    const payload = JSON.stringify({
        username: 'nasabah_01',
        password: 'rahasia',
    });

    const params = {
        headers: { 'Content-Type': 'application/json' },
    };

    const res = http.post('http://backend-api:9000/api/auth/login', payload, params);

    check(res, {
        'status is 200': (r) => r.status === 200,
        'latency < 500ms': (r) => r.timings.duration < 500,
    });

    sleep(1);
}
