// SSR data loader for /meters.
//
// Server-side fetch: runs inside the SvelteKit Node container, reaches the
// backend over the docker compose network, performs the CSRF dance + dev
// login + meter list. Returns the result as page data prop.
//
// Dev-only convenience: hard-coded operator credentials. PRODUCTION must
// replace this with session-bound auth (cookie-based or token-passing).
import type { PageServerLoad } from './$types';
import { env as privateEnv } from '$env/dynamic/private';
import { env as publicEnv } from '$env/dynamic/public';

const BACKEND = (privateEnv.INTERNAL_API_BASE || publicEnv.PUBLIC_API_BASE || 'http://greenmetrics-backend:8082/api/v1').replace(/\/$/, '');
const DEV_EMAIL = privateEnv.DEV_EMAIL || 'dev@greenmetrics.local';
const DEV_PASSWORD = privateEnv.DEV_PASSWORD || 'OperatorPass2026!';

async function devLogin(): Promise<string> {
  // Step 1: GET to obtain gm_csrf cookie.
  const probe = await fetch(`${BACKEND}/meters`, { method: 'GET' });
  const setCookie = probe.headers.get('set-cookie') ?? '';
  const csrfMatch = setCookie.match(/gm_csrf=([^;,]+)/);
  const csrf = csrfMatch ? csrfMatch[1] : '';

  // Step 2: POST /auth/login with cookie + matching header.
  const r = await fetch(`${BACKEND}/auth/login`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...(csrf
        ? { 'Cookie': `gm_csrf=${csrf}`, 'X-CSRF-Token': csrf }
        : {}),
    },
    body: JSON.stringify({ email: DEV_EMAIL, password: DEV_PASSWORD }),
  });
  if (!r.ok) {
    const body = await r.text();
    throw new Error(`login ${r.status}: ${body.slice(0, 200)}`);
  }
  const j = (await r.json()) as { access_token?: string };
  if (!j.access_token) throw new Error('login: missing access_token');
  return j.access_token;
}

export const load: PageServerLoad = async () => {
  try {
    const token = await devLogin();
    const r = await fetch(`${BACKEND}/meters`, {
      headers: { 'Authorization': `Bearer ${token}` },
    });
    if (!r.ok) {
      const body = await r.text();
      return { meters: [], error: `meters ${r.status}: ${body.slice(0, 200)}` };
    }
    const data = (await r.json()) as { items?: unknown[] };
    return { meters: data.items ?? [], error: null };
  } catch (e) {
    return { meters: [], error: e instanceof Error ? e.message : String(e) };
  }
};
