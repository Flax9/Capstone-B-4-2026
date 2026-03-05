import http from 'k6/http';
import { sleep, check } from 'k6';

export const options = {
    vus: 10,
    duration: '30s',
};

export default function () {
    const res = http.get('http://backend-api:8080/actuator/health');
    check(res, {
        'is status 200': (r) => r.status === 200,
    });
    sleep(1);
}
