<script lang="ts">
  import Scope123Breakdown from '$components/Scope123Breakdown.svelte';
  import CarbonFootprint from '$components/CarbonFootprint.svelte';
  import EmissionFactorEditor from '$components/EmissionFactorEditor.svelte';

  // Factor sources currently loaded — surfaced from `manifest.lock.json` in
  // production. Demo-populated here.
  interface FactorSource {
    id: string;
    name: string;
    publisher: string;
    publishedAt: string;
    coverage: string;
    valid: { from: string; to: string };
  }

  const factorSources: FactorSource[] = [
    {
      id: 'ispra-2026-04',
      name: 'ISPRA — Coefficienti emissione mix elettrico nazionale',
      publisher: 'ISPRA · Rapporto 404/2026',
      publishedAt: '2026-04-01',
      coverage: 'Scope 2 location-based · IT',
      valid: { from: '2026-01-01', to: '2026-12-31' }
    },
    {
      id: 'gse-aib-2026',
      name: 'AIB Residual Mix · IT',
      publisher: 'AIB · pubblicato via GSE',
      publishedAt: '2026-03-15',
      coverage: 'Scope 2 market-based · IT',
      valid: { from: '2026-01-01', to: '2026-12-31' }
    },
    {
      id: 'terna-2026-04',
      name: 'Terna — mix nazionale giornaliero',
      publisher: 'Terna · Trasparenza',
      publishedAt: '2026-04-30',
      coverage: 'Scope 2 hourly granularity · IT',
      valid: { from: '2026-04-30', to: '2026-04-30' }
    },
    {
      id: 'ipcc-ar6-2025',
      name: 'GHG Protocol — combustibili fossili',
      publisher: 'IPCC AR6 · 2025',
      publishedAt: '2025-09-12',
      coverage: 'Scope 1 stationary + mobile combustion',
      valid: { from: '2025-01-01', to: '2027-12-31' }
    }
  ];
</script>

<section class="space-y-8">
  <!-- ── Scope KPI strip ───────────────────────────────────────────── -->
  <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
    <div class="card">
      <div class="flex items-center justify-between">
        <p class="kpi-label">Scope 1 · diretto</p>
        <span class="badge-scope-1">on-site</span>
      </div>
      <p class="kpi-value">94,2 <span class="kpi-unit">t CO₂e</span></p>
      <p class="muted text-xs mt-1">Caldaia metano · veicoli flotta · gruppo elettrogeno</p>
    </div>
    <div class="card">
      <div class="flex items-center justify-between">
        <p class="kpi-label">Scope 2 · indiretto</p>
        <span class="badge-scope-2">elettrico</span>
      </div>
      <p class="kpi-value">186,4 <span class="kpi-unit">t CO₂e</span></p>
      <p class="muted text-xs mt-1">location-based · ISPRA mix nazionale 2026</p>
    </div>
    <div class="card">
      <div class="flex items-center justify-between">
        <p class="kpi-label">Scope 3 · catena valore</p>
        <span class="badge-scope-3">cat. 1+4+5</span>
      </div>
      <p class="kpi-value">412,8 <span class="kpi-unit">t CO₂e</span></p>
      <p class="muted text-xs mt-1">Acquisti beni e servizi · trasporti upstream · rifiuti</p>
    </div>
  </div>

  <!-- ── Charts row ────────────────────────────────────────────────── -->
  <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
    <div class="card">
      <header class="flex items-end justify-between mb-4">
        <div>
          <h3>Scope 1 · 2 · 3 (YTD)</h3>
          <p class="muted text-xs mt-0.5">693,4 t CO₂e totali · location-based</p>
        </div>
        <div class="flex gap-1">
          <button class="btn-ghost btn-sm bg-forest-50">Location</button>
          <button class="btn-ghost btn-sm">Market</button>
        </div>
      </header>
      <Scope123Breakdown />
    </div>
    <div class="card">
      <header class="flex items-end justify-between mb-4">
        <div>
          <h3>Andamento 12 mesi</h3>
          <p class="muted text-xs mt-0.5">Trend mensile · Scope 1+2 combinato</p>
        </div>
      </header>
      <CarbonFootprint />
    </div>
  </div>

  <!-- ── Factor sources (audit-grade tracking) ─────────────────────── -->
  <section class="card">
    <header class="flex items-end justify-between mb-4">
      <div>
        <h3>Factor source caricati</h3>
        <p class="muted mt-0.5">
          Ogni KPI di scope è calcolato contro il factor valido al midpoint del periodo (Doctrine Rule 90).
          Cambiare factor source forza una rigenerazione del report con un nuovo manifest_lock.
        </p>
      </div>
      <span class="badge-success">4 attivi</span>
    </header>
    <div class="overflow-x-auto">
      <table class="table-clean">
        <thead>
          <tr>
            <th>Factor source</th>
            <th>Publisher</th>
            <th>Pubblicato</th>
            <th>Copertura</th>
            <th>Validità</th>
          </tr>
        </thead>
        <tbody>
          {#each factorSources as fs}
            <tr>
              <td>
                <p class="font-medium">{fs.name}</p>
                <code class="font-mono text-[11px] text-forest-600">{fs.id}</code>
              </td>
              <td class="text-forest-700">{fs.publisher}</td>
              <td class="font-mono text-xs">{fs.publishedAt}</td>
              <td class="text-forest-700">{fs.coverage}</td>
              <td class="font-mono text-xs">
                <span>{fs.valid.from}</span>
                <span class="text-forest-500"> → </span>
                <span>{fs.valid.to}</span>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  </section>

  <!-- ── Per-source override editor (if engagement needs custom values) -->
  <section class="card">
    <header class="flex items-end justify-between mb-4">
      <div>
        <h3>Editor fattori (override per stabilimento)</h3>
        <p class="muted mt-0.5">
          Ogni override è temporal-keyed e produce un audit trail con motivazione + EGE counter-signature.
        </p>
      </div>
    </header>
    <EmissionFactorEditor />
  </section>
</section>
