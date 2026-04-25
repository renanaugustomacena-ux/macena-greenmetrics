// k6 load test — POST /api/v1/readings/ingest sustained at SLO target.
//
// Doctrine refs: Rule 24, Rule 37, Rule 44.
// SLO: p99 ≤ 120 ms, target 5 000 RPS sustained per `docs/backend/slo.md`.
// Plan: nightly CI run gates regression > +20% p99.

import http from 'k6/http';
import { check, sleep } from 'k6';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

export const options = {
  scenarios: {
    sustained: {
      executor: 'constant-arrival-rate',
      rate: 5000,                // 5k RPS target
      timeUnit: '1s',
      duration: '5m',
      preAllocatedVUs: 200,
      maxVUs: 1000,
    },
  },
  thresholds: {
    'http_req_duration{path:ingest}': ['p(99)<120', 'p(95)<60'],
    'http_req_failed{path:ingest}': ['rate<0.001'],
  },
  tags: { test: 'ingest', slo: 'p99-120ms' },
};

const BASE = __ENV.API_BASE || 'https://staging.greenmetrics.it';
const TOKEN = __ENV.API_TOKEN || '';
const METER = __ENV.METER_ID || '00000000-0000-4000-8000-000000000001';
const CHANNEL = __ENV.CHANNEL_ID || '00000000-0000-4000-8000-000000000002';

export default function () {
  const payload = JSON.stringify({
    meter_id: METER,
    readings: [
      {
        ts: new Date().toISOString(),
        meter_id: METER,
        channel_id: CHANNEL,
        value: Math.random() * 100,
        unit: 'kWh',
        quality_code: 0,
      },
    ],
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${TOKEN}`,
      'Idempotency-Key': uuidv4(),
    },
    tags: { path: 'ingest' },
  };

  const res = http.post(`${BASE}/api/v1/readings/ingest`, payload, params);
  check(res, { 'accepted': (r) => r.status === 202 || r.status === 200 });
}
