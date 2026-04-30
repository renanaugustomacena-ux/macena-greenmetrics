<script lang="ts">
  import EnergyChart from '$components/EnergyChart.svelte';
  import CarbonFootprint from '$components/CarbonFootprint.svelte';
  import ESGScorecard from '$components/ESGScorecard.svelte';
  import AlarmFeed from '$components/AlarmFeed.svelte';
  import Scope123Breakdown from '$components/Scope123Breakdown.svelte';
  import ConsumptionHeatmap from '$components/ConsumptionHeatmap.svelte';
  import RegulatoryDossier from '$components/RegulatoryDossier.svelte';
  import ComplianceCalendar from '$components/ComplianceCalendar.svelte';
  import PackLoaderStatus from '$components/PackLoaderStatus.svelte';
</script>

<section class="space-y-8">
  <!-- ── KPI strip ─────────────────────────────────────────────────── -->
  <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
    <div class="card">
      <p class="kpi-label">Consumo mese</p>
      <p class="kpi-value">128.430 <span class="kpi-unit">kWh</span></p>
      <p class="mt-1 kpi-delta-positive">↓ 4,8% vs baseline · 6 stabilimenti</p>
    </div>
    <div class="card">
      <p class="kpi-label">Emissioni YTD</p>
      <p class="kpi-value">312,6 <span class="kpi-unit">t CO₂e</span></p>
      <p class="mt-1 muted text-xs">Scope 1+2 location-based · ISPRA 2026</p>
    </div>
    <div class="card">
      <p class="kpi-label">Credito Piano 5.0 stimato</p>
      <p class="kpi-value">€ 186.400</p>
      <p class="mt-1 muted text-xs">attestazione EGE pronta · invio Q2</p>
    </div>
    <div class="card">
      <p class="kpi-label">CSRD ESRS E1 readiness</p>
      <p class="kpi-value">78<span class="kpi-unit">%</span></p>
      <p class="mt-1 muted text-xs">data-points coperti · prima pubblicazione 2027</p>
    </div>
  </div>

  <!-- ── Regulatory dossiers strip ─────────────────────────────────── -->
  <section>
    <header class="flex items-end justify-between mb-3">
      <div>
        <p class="eyebrow">Dossier regolatori italiani</p>
        <h2 class="mt-0.5">Stato di compliance</h2>
      </div>
      <a href="/reports" class="btn-ghost btn-sm">
        Vedi tutti →
      </a>
    </header>
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
      <RegulatoryDossier
        title="ESRS E1 — Cambiamenti climatici"
        regulator="EFRAG / CSRD wave 2"
        period="FY 2025"
        dueDate="2026-04-30"
        status="ready"
        signed
        provenanceHash="sha256:7a1c4f9d2b8e6c3f"
      />
      <RegulatoryDossier
        title="Piano Transizione 5.0 — Attestazione"
        regulator="GSE / MIMIT-MASE"
        period="Q2 2026"
        dueDate="2026-06-30"
        status="inflight"
        provenanceHash="sha256:3e9b8a5c1f4d7e2a"
      />
      <RegulatoryDossier
        title="Conto Termico 2.0 — Richiesta"
        regulator="GSE"
        period="Apr 2026"
        dueDate="2026-05-31"
        status="due"
      />
      <RegulatoryDossier
        title="Audit Energetico D.Lgs. 102/2014"
        regulator="ENEA"
        period="2024–2027"
        dueDate="2027-12-05"
        status="inflight"
        signed
      />
    </div>
  </section>

  <!-- ── Charts row ────────────────────────────────────────────────── -->
  <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
    <div class="lg:col-span-2 card">
      <header class="flex items-end justify-between mb-4">
        <h3>Consumo energetico — ultime 4 settimane</h3>
        <div class="flex gap-1">
          <button class="btn-ghost btn-sm bg-forest-50">4w</button>
          <button class="btn-ghost btn-sm">12w</button>
          <button class="btn-ghost btn-sm">YTD</button>
        </div>
      </header>
      <EnergyChart />
    </div>
    <div class="card">
      <h3 class="mb-4">Impronta di carbonio</h3>
      <CarbonFootprint />
    </div>
  </div>

  <!-- ── Three-column row: Scope · Calendar · Alarms ───────────────── -->
  <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
    <div class="card">
      <h3 class="mb-4">Scope 1 · 2 · 3</h3>
      <Scope123Breakdown />
    </div>
    <div class="card">
      <header class="flex items-center justify-between mb-3">
        <h3>Calendario compliance</h3>
        <span class="badge-warning">2 in scadenza</span>
      </header>
      <ComplianceCalendar />
    </div>
    <div class="card">
      <header class="flex items-center justify-between mb-3">
        <h3>Allarmi recenti</h3>
        <span class="badge-neutral">3 aperti</span>
      </header>
      <AlarmFeed />
    </div>
  </div>

  <!-- ── ESG + Heatmap row ─────────────────────────────────────────── -->
  <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
    <div class="card">
      <h3 class="mb-4">Scorecard ESG</h3>
      <ESGScorecard />
    </div>
    <div class="card lg:col-span-2">
      <h3 class="mb-4">Mappa di consumo — ora × giorno</h3>
      <ConsumptionHeatmap />
    </div>
  </div>

  <!-- ── Audit-grade evidence (regulator-grade differentiator) ─────── -->
  <section class="card-accent">
    <header class="flex items-end justify-between mb-4">
      <div>
        <p class="eyebrow text-forest-700">Audit-grade evidence · Rule 73 · Rule 89</p>
        <h2 class="mt-0.5">Pack-loader, manifest-lock, audit chain</h2>
        <p class="muted mt-1 max-w-2xl">
          Ogni report e ogni KPI è ricostruibile bit-per-bit da
          <code class="font-mono text-xs">(period, factors, readings, manifest_lock_hash)</code>.
          Il manifest lock anchorato qui è la stessa firma che vedi nell'audit-evidence-pack.
        </p>
      </div>
    </header>
    <div class="bg-white/70 rounded-lg p-5 border border-forest-100">
      <PackLoaderStatus />
    </div>
  </section>
</section>
