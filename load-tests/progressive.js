import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '2m', target: 50 },
    { duration: '2m', target: 100 },
    { duration: '2m', target: 250 },
    { duration: '2m', target: 500 },
  ],
};

export default function () {
  const payload = JSON.stringify({
    email: 'test@test.com',
    subject: 'Carga',
    message: 'Test progresivo'
  });

  const res = http.post(
    'https://anhgbkuzm5.execute-api.us-east-1.amazonaws.com/notifications/send',
    payload,
    { headers: { 'Content-Type': 'application/json' } }
  );

  check(res, { 'status 200': (r) => r.status === 200 });
  sleep(1);
}