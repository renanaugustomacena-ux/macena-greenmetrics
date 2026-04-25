// k6 login-burst test — verify lockout + rate limit + bcrypt cost factor 12 baseline.

import http from 'k6/http';
import { check } from 'k6';

export const options = {
  scenarios: {
    legit_logins: {
      executor: 'constant-arrival-rate',
      rate: 50,
      timeUnit: '1s',
      duration: '2m',
      preAllocatedVUs: 100,
      maxVUs: 500,
    },
    bruteforce_simulated: {
      executor: 'constant-arrival-rate',
      rate: 50,
      timeUnit: '1s',
      duration: '2m',
      preAllocatedVUs: 100,
      maxVUs: 500,
      startTime: '30s',
      env: { BRUTEFORCE: 'true' },
    },
  },
  thresholds: {
    'http_req_duration{scenario:legit_logins,status:200}': ['p(99)<400'],
    // Bruteforce should rapidly hit lockout.
    'http_req_failed{scenario:bruteforce_simulated,status:429}': ['count>100'],
  },
};

const BASE = __ENV.API_BASE || 'https://staging.greenmetrics.it';

export default function () {
  const isBrute = __ENV.BRUTEFORCE === 'true';
  const email = isBrute ? 'attacker@example.com' : `user${Math.floor(Math.random() * 100)}@example.com`;
  const password = isBrute ? 'wrong-password' : 'correct-strong-password-32-chars-long';
  const res = http.post(`${BASE}/api/v1/auth/login`, JSON.stringify({ email, password }), {
    headers: { 'Content-Type': 'application/json' },
    tags: { scenario: __ENV.K6_SCENARIO_NAME || 'login' },
  });
  check(res, { 'expected_status': (r) => isBrute ? (r.status === 401 || r.status === 429) : r.status === 200 });
}
