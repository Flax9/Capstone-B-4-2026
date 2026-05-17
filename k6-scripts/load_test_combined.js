import grpc from 'k6/net/grpc';
import { check } from 'k6';

// Instansiasi 3 Klien gRPC
const clientAuth = new grpc.Client();
const clientBalance = new grpc.Client();
const clientTrx = new grpc.Client();

// Muat kontrak Protocol Buffers
// Path absolut /proto/ agar bisa bekerja di dalam Docker container
clientAuth.load(['/proto/auth'], 'auth.proto');
clientBalance.load(['/proto/balance'], 'balance.proto');
clientTrx.load(['/proto/transaction'], 'transaction.proto');

export const options = {
    scenarios: {
        uji_beban_gabungan: {
            executor: 'constant-arrival-rate',
            rate: 300000,          // Target: 1.000.000 request
            timeUnit: '1m',
            duration: '2m',
            preAllocatedVUs: 3000,
            maxVUs: 10000,
        },
    },
};

// State per-VU untuk memastikan koneksi TCP persisten (Multiplexing)
let connected = false;

export default function () {
    // Hanya buka 1 koneksi per Virtual User seumur hidup (menghindari port exhaustion)
    if (!connected) {
        const targetHost = __ENV.TARGET_HOST || 'localhost';
        clientAuth.connect(`${targetHost}:9001`, { plaintext: true });
        clientBalance.connect(`${targetHost}:9002`, { plaintext: true });
        clientTrx.connect(`${targetHost}:9003`, { plaintext: true });
        connected = true;
    }

    const rand = Math.random();

    if (rand < 0.33) {
        // 1. gRPC Login
        const response = clientAuth.invoke('auth.AuthService/Login', {
            username: 'nasabah_01',
            password: 'rahasia'
        });

        check(response, {
            'login status is OK': (r) => r && r.status === grpc.StatusOK,
            'login code is 200': (r) => r && r.message && r.message.statusCode === 200,
        });

    } else if (rand < 0.66) {
        // 2. gRPC Cek Saldo
        const response = clientBalance.invoke('balance.BalanceService/CheckBalance', {
            userId: '924de2cf-e950-4f92-8e37-ae2eb7dda7e5'
        });

        check(response, {
            'cek_saldo status is OK': (r) => r && r.status === grpc.StatusOK,
            'cek_saldo code is 200': (r) => r && r.message && r.message.statusCode === 200,
        });

    } else {
        // 3. gRPC Transfer
        const response = clientTrx.invoke('transaction.TransactionService/Transfer', {
            senderId: '924de2cf-e950-4f92-8e37-ae2eb7dda7e5',
            targetAccount: 'e3acd2bc-94d1-475e-ac7a-12fe405ad426',
            amount: 10
        });

        check(response, {
            'transfer status is OK': (r) => r && r.status === grpc.StatusOK,
            'transfer code is 202': (r) => r && r.message && r.message.statusCode === 202,
        });
    }

    // Tidak ada pemanggilan client.close() agar koneksi tetap persisten!
}
