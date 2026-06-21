import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  vus: 10,
  duration: '1m',
};

export default function () {
  const url = 'https://anhgbkuzm5.execute-api.us-east-1.amazonaws.com/auth/login';
  // Primero login para obtener token (ajusta credenciales)
  const loginRes = http.post(url, JSON.stringify({
    matricula: 'TU_MATRICULA',
    password: 'TU_PASSWORD'
  }), { headers: { 'Content-Type': 'application/json' } });

  const token = JSON.parse(loginRes.body).token;

  const payload = JSON.stringify({
    email: 'test@test.com',
    subject: 'Prueba',
    message: 'Mensaje de carga'
  });

  const res = http.post(
    'https://anhgbkuzm5.execute-api.us-east-1.amazonaws.com/notifications/send',
    payload,
    { headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${token}` } }
  );

  check(res, {
    'status is 200': (r) => r.status === 200,
  });

  sleep(1);
}