// k6 dev-stack bench: modest sustained load against the docker-compose backend.
// Runs at 100 RPS for 30 s — suitable for a laptop, not for SLO regression
// gating (use ingest.js + ingest_burst.js against staging for that).
//
// Run via:
//   docker run --rm --network greenmetrics_greenmetrics \
//     -v "$PWD":/scripts \
//     -e API_BASE=http://greenmetrics-backend:8082 \
//     -e API_TOKEN="$(cat /tmp/greenmetrics-token.txt)" \
//     -e METER_ID=00000000-0000-4000-8000-0000000eee01 \
//     -e CHANNEL_ID=00000000-0000-4000-8000-0000ccc00001 \
//     grafana/k6 run /scripts/bench-dev.js

import http from 'k6/http';
import { check } from 'k6';
import { uuidv4 } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

const BASE     = __ENV.API_BASE     || 'http://localhost:8082';
const TOKEN    = __ENV.API_TOKEN    || '';
const METER    = __ENV.METER_ID     || '00000000-0000-4000-8000-0000000eee01';
const CHANNEL  = __ENV.CHANNEL_ID   || '00000000-0000-4000-8000-0000ccc00001';

export const options = {
  scenarios: {
    sustained: {
      executor: 'constant-arrival-rate',
      rate: parseInt(__ENV.RATE || '100', 10),
      timeUnit: '1s',
      duration: __ENV.DURATION || '30s',
      preAllocatedVUs: 20,
      maxVUs: 200,
    },
  },
  thresholds: {
    'http_req_duration{path:ingest}': ['p(99)<300', 'p(95)<150'],
    'http_req_failed{path:ingest}': ['rate<0.02'],
    'checks{check:accepted}': ['rate>0.98'],
  },
  summaryTrendStats: ['avg', 'min', 'med', 'p(90)', 'p(95)', 'p(99)', 'max'],
};

export default function () {
  // Each iteration must produce a unique (tenant, meter, channel, ts) triple
  // to satisfy the readings_dedup_idx unique constraint. JS Date is ms
  // precision; tight k6 loops collide. Spread ts across a wide window per
  // (VU, iteration) so the bench measures throughput, not dedup conflicts.
  const tsMs = Date.now() + (__VU * 1_000_000) + (__ITER * 7);
  const payload = JSON.stringify({
    readings: [
      {
        ts: new Date(tsMs).toISOString(),
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
  check(res, { 'accepted': (r) => r.status === 200 || r.status === 202 }, { check: 'accepted' });
}
