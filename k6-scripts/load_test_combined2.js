import grpc from 'k6/net/grpc';
import { check } from 'k6';

// Helper: generate angka acak antara min dan max (inklusif)
function randomIntBetween(min, max) {
    return Math.floor(Math.random() * (max - min + 1)) + min;
}

const clientAuth = new grpc.Client();
const clientBalance = new grpc.Client();
const clientTrx = new grpc.Client();

clientAuth.load(['/proto/auth'], 'auth.proto');
clientBalance.load(['/proto/balance'], 'balance.proto');
clientTrx.load(['/proto/transaction'], 'transaction.proto');

export const options = {
    scenarios: {
        uji_beban_gabungan: {
            executor: 'constant-arrival-rate',
            rate: 1000000,          // Target: 1.000.000 request
            timeUnit: '1m',         // per menit (~16.666 RPS)
            duration: '5m',
            // OPTIMASI VU: Turunkan drastis agar laptop tidak crash
            preAllocatedVUs: 1200,
            maxVUs: 1500,           // Maksimal hanya 4.500 TCP koneksi total (Sangat aman!)
        },
    },
};

let connected = false;

export default function () {
    if (!connected) {
        const targetHost = __ENV.TARGET_HOST || 'localhost';
        clientAuth.connect(`${targetHost}:9001`, { plaintext: true });
        clientBalance.connect(`${targetHost}:9002`, { plaintext: true });
        clientTrx.connect(`${targetHost}:9003`, { plaintext: true });
        connected = true;
    }

    const rand = Math.random();

    // OPTIMASI DATA: Buat ID secara acak untuk memukul database sungguhan, bukan cache
    // Asumsi ID nasabah Anda menggunakan pola angka atau bisa di-generate stringnya
    const randomUserId = `nasabah_${randomIntBetween(1, 100000).toString().padStart(2, '0')}`;
    const randomUuid = `924de2cf-e950-4f92-8e37-ae2eb7dda${randomIntBetween(100, 999)}`;

    if (rand < 0.33) {
        const response = clientAuth.invoke('auth.AuthService/Login', {
            username: randomUserId,
            password: 'rahasia'
        });
        check(response, { 'login OK': (r) => r && r.status === grpc.StatusOK });
    } else if (rand < 0.66) {
        const response = clientBalance.invoke('balance.BalanceService/CheckBalance', {
            userId: randomUuid
        });
        check(response, { 'cek_saldo OK': (r) => r && r.status === grpc.StatusOK });
    } else {
        const response = clientTrx.invoke('transaction.TransactionService/Transfer', {
            senderId: randomUuid,
            targetAccount: 'e3acd2bc-94d1-475e-ac7a-12fe405ad426',
            amount: 10
        });
        check(response, { 'transfer OK': (r) => r && r.status === grpc.StatusOK });
    }
}

// Tambahkan teardown untuk membersihkan koneksi dengan aman saat test selesai
export function teardown() {
    clientAuth.close();
    clientBalance.close();
    clientTrx.close();
}
