import http from 'k6/http';
import { sleep } from 'k6';

export const options = {
  stages: [
    { duration: '1m', target: 100 },
    { duration: '1m', target: 250 },
    { duration: '1m', target: 500 },
    { duration: '1m', target: 750 },
    { duration: '1m', target: 1000 },
    { duration: '1m', target: 1500 },
    { duration: '1m', target: 2000 },
  ],
};

export default function () {
  const payload = JSON.stringify({
    email: 'test@test.com',
    subject: 'Estres',
    message: 'Test estres'
  });

  http.post(
    'https://anhgbkuzm5.execute-api.us-east-1.amazonaws.com/notifications/send',
    payload,
    { headers: { 'Content-Type': 'application/json' } }
  );
  sleep(0.5);
}