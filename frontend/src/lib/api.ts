// Typed API client — thin fetch wrapper with retries and structured errors.
//
// API base URL resolution (SvelteKit-typed):
//   * Primary: $env/dynamic/public — runtime-injected, typed string|undefined.
//     Set via PUBLIC_API_BASE in the deploy environment; lets us redeploy
//     without rebuilding the JS bundle.
//   * Fallback: '/api/v1' relative path (works when frontend + backend are
//     served from the same origin behind a reverse proxy).
import { env as publicEnv } from '$env/dynamic/public';

const BASE = publicEnv.PUBLIC_API_BASE || '/api/v1';

export interface ProblemJSON {
  type: string;
  title: string;
  status: number;
  detail?: string;
  instance?: string;
  code?: string;
}

export class APIError extends Error {
  readonly status: number;
  readonly problem: ProblemJSON;
  readonly requestId?: string;
  constructor(p: ProblemJSON, requestId?: string) {
    super(p.detail || p.title);
    this.status = p.status;
    this.problem = p;
    this.requestId = requestId;
  }
}

function buildHeaders(token?: string, extra?: HeadersInit): HeadersInit {
  const h = new Headers(extra);
  h.set('content-type', 'application/json');
  if (token) h.set('authorization', `Bearer ${token}`);
  return h;
}

async function request<T>(
  path: string,
  opts: RequestInit & { token?: string; retries?: number } = {}
): Promise<T> {
  const url = path.startsWith('http') ? path : `${BASE}${path}`;
  const retries = opts.retries ?? 1;
  const init: RequestInit = {
    ...opts,
    headers: buildHeaders(opts.token, opts.headers)
  };
  let lastErr: unknown;
  for (let attempt = 0; attempt <= retries; attempt++) {
    try {
      const resp = await fetch(url, init);
      if (!resp.ok) {
        const body: ProblemJSON = await resp
          .json()
          .catch(() => ({ type: 'about:blank', title: resp.statusText, status: resp.status }));
        throw new APIError(body, resp.headers.get('x-request-id') ?? undefined);
      }
      if (resp.status === 204) return undefined as T;
      return (await resp.json()) as T;
    } catch (err) {
      lastErr = err;
      if (err instanceof APIError && err.status < 500) throw err;
      if (attempt === retries) throw err;
      await new Promise((r) => setTimeout(r, 200 * (attempt + 1)));
    }
  }
  throw lastErr;
}

/* -------------------------------------------------------------------------- */
/*  Typed API surface                                                         */
/* -------------------------------------------------------------------------- */

export interface HealthResponse {
  status: string;
  service: string;
  version: string;
  uptime_seconds: number;
  time: string;
  dependencies: Record<string, string>;
}
export function getHealth() {
  return request<HealthResponse>('/../health');
}

export interface Meter {
  id: string;
  tenant_id: string;
  label: string;
  meter_type: string;
  protocol: string;
  unit: string;
  site: string;
  cost_centre?: string;
  serial_no?: string;
  pod_code?: string;
  pdr_code?: string;
  active: boolean;
  created_at: string;
  updated_at: string;
}
export function listMeters(token: string) {
  return request<{ items: Meter[]; total: number }>('/meters', { token });
}
export function createMeter(token: string, meter: Partial<Meter>) {
  return request<Meter>('/meters', { method: 'POST', body: JSON.stringify(meter), token });
}

export interface Aggregate {
  bucket: string;
  meter_id: string;
  channel_id?: string;
  sum_value: number;
  avg_value: number;
  max_value: number;
  unit: string;
}
export function queryAggregated(
  token: string,
  params: { meter_id: string; resolution: '15min' | '1h' | '1d'; from: string; to: string }
) {
  const qs = new URLSearchParams(params).toString();
  return request<{ resolution: string; items: Aggregate[]; count: number }>(
    `/readings/aggregated?${qs}`,
    { token }
  );
}

export type ReportType =
  | 'monthly_consumption'
  | 'co2_footprint'
  | 'esrs_e1_csrd'
  | 'piano_5_0_attestazione'
  | 'conto_termico_2_0'
  | 'certificati_bianchi_tee'
  | 'audit_dlgs_102_2014';

export interface Report {
  id: string;
  tenant_id: string;
  type: ReportType;
  period_from: string;
  period_to: string;
  status: string;
  payload: Record<string, unknown>;
  file_url?: string;
  generated_by: string;
  created_at: string;
  updated_at: string;
}
export function generateReport(
  token: string,
  body: { type: ReportType; period_from: string; period_to: string; options?: Record<string, unknown> }
) {
  return request<Report>('/reports', { method: 'POST', body: JSON.stringify(body), token });
}

export interface Alert {
  id: string;
  tenant_id: string;
  meter_id?: string;
  kind: string;
  severity: 'info' | 'warning' | 'critical';
  message: string;
  context?: Record<string, unknown>;
  triggered_at: string;
  acked_at?: string;
  acked_by?: string;
  resolved_at?: string;
}
export function listAlerts(token: string) {
  return request<{ items: Alert[]; total: number }>('/alerts', { token });
}

export async function login(email: string, password: string) {
  return request<{ access_token: string; refresh_token: string; expires_in: number; token_type: string }>(
    '/auth/login',
    { method: 'POST', body: JSON.stringify({ email, password }) }
  );
}
