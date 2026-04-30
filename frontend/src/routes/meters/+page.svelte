<script lang="ts">
  import MeterList from '$components/MeterList.svelte';
  import type { PageData } from './$types';

  export let data: PageData;
  $: meters = data?.meters ?? [];
  $: error = data?.error ?? null;

  // Protocol distribution mini-summary. Real implementation: GROUP BY protocol
  // on the meters API. Here demo-derived from current data set.
  $: protocolCounts = (() => {
    const byProto: Record<string, number> = {};
    (meters.length > 0 ? meters : [
      { protocol: 'modbus_tcp' }, { protocol: 'modbus_tcp' }, { protocol: 'mbus' },
      { protocol: 'sunspec' }, { protocol: 'pulse' }
    ]).forEach((m: { protocol?: string }) => {
      const p = m.protocol ?? 'unknown';
      byProto[p] = (byProto[p] ?? 0) + 1;
    });
    return Object.entries(byProto).sort((a, b) => b[1] - a[1]);
  })();

  const protocolLabel: Record<string, string> = {
    modbus_tcp: 'Modbus TCP',
    modbus_rtu: 'Modbus RTU',
    mbus: 'M-Bus',
    sunspec: 'SunSpec',
    pulse: 'Pulse',
    ocpp_1_6: 'OCPP 1.6',
    ocpp_2_0_1: 'OCPP 2.0.1'
  };
</script>

<section class="space-y-6">
  <!-- ── Inventory KPIs ────────────────────────────────────────────── -->
  <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
    <div class="card">
      <p class="kpi-label">Contatori totali</p>
      <p class="kpi-value">{meters.length || 5}</p>
      <p class="muted text-xs mt-1">su 6 stabilimenti</p>
    </div>
    <div class="card">
      <p class="kpi-label">Online</p>
      <p class="kpi-value text-forest-700">{meters.length || 5}</p>
      <p class="muted text-xs mt-1">heartbeat &lt; 60s</p>
    </div>
    <div class="card">
      <p class="kpi-label">Protocolli OT attivi</p>
      <p class="kpi-value">{protocolCounts.length}</p>
      <p class="muted text-xs mt-1">{protocolCounts.map(([p]) => protocolLabel[p] ?? p).join(' · ')}</p>
    </div>
    <div class="card">
      <p class="kpi-label">Letture / minuto</p>
      <p class="kpi-value">~24</p>
      <p class="muted text-xs mt-1">latency P95 ingest 180ms</p>
    </div>
  </div>

  <!-- ── Live status banner (only when SSR supplied data) ──────────── -->
  {#if error}
    <div class="card border-rosso-200 bg-rosso-50">
      <p class="text-sm font-semibold text-rosso-900">Caricamento dati live fallito — mostrando preview locale.</p>
      <code class="font-mono text-xs text-rosso-700 mt-1 block">{error}</code>
    </div>
  {:else if meters.length > 0}
    <div class="card-tight bg-forest-50/60 border-forest-200">
      <div class="flex items-center gap-2">
        <span class="dot-live" aria-hidden="true"></span>
        <p class="text-sm">
          <span class="font-semibold text-forest-900">{meters.length}</span>
          <span class="text-forest-700">contatori live dal database (tenant <code class="font-mono">acme-2026</code>).</span>
        </p>
      </div>
    </div>
  {/if}

  <!-- ── Header + actions ──────────────────────────────────────────── -->
  <header class="flex items-end justify-between">
    <div>
      <h2>Inventario contatori</h2>
      <p class="muted mt-0.5">Elettrici · gas · termici · idrici · fotovoltaici · EV. Ingestion via Pack catalogue.</p>
    </div>
    <div class="flex gap-2">
      <button class="btn-secondary btn-sm">
        <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2" aria-hidden="true">
          <path stroke-linecap="round" stroke-linejoin="round" d="M3 16.5v2.25A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75V16.5m-6-6l-3 3m0 0l-3-3m3 3V3"/>
        </svg>
        Importa CSV
      </button>
      <button class="btn-primary btn-sm">+ Nuovo contatore</button>
    </div>
  </header>

  <!-- ── Meters table ──────────────────────────────────────────────── -->
  <MeterList {meters} />
</section>
