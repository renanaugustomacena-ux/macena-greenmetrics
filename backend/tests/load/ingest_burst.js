// k6 burst test — 20× normal rate for 2 min; verify backpressure + drop policy.

import http from 'k6/http';
import { check } from 'k6';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

export const options = {
  scenarios: {
    burst: {
      executor: 'constant-arrival-rate',
      rate: 100000,              // 100k RPS (20× the 5k target)
      timeUnit: '1s',
      duration: '2m',
      preAllocatedVUs: 1000,
      maxVUs: 5000,
    },
  },
  thresholds: {
    // Allow 503 Retry-After during burst — verify we degrade gracefully.
    'http_req_failed{path:ingest}': ['rate<0.5'],
    // Verify p99 latency for accepted requests stays bounded (queue saturated → reject not slow).
    'http_req_duration{status:202}': ['p(99)<150'],
    // 503 Retry-After should be the dominant non-2xx response, not 5xx.
    'http_req_failed{status:503}': ['count>0'],
  },
  tags: { test: 'ingest_burst', sla: 'graceful-degrade' },
};

const BASE = __ENV.API_BASE || 'https://staging.greenmetrics.it';
const TOKEN = __ENV.API_TOKEN || '';

export default function () {
  const payload = JSON.stringify({
    meter_id: '00000000-0000-4000-8000-000000000001',
    readings: [{ ts: new Date().toISOString(), meter_id: '00000000-0000-4000-8000-000000000001', channel_id: '00000000-0000-4000-8000-000000000002', value: 1, unit: 'kWh', quality_code: 0 }],
  });
  const res = http.post(`${BASE}/api/v1/readings/ingest`, payload, {
    headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${TOKEN}`, 'Idempotency-Key': uuidv4() },
    tags: { path: 'ingest' },
  });
  check(res, { 'accept_or_retry': (r) => r.status === 202 || r.status === 503 });
}
