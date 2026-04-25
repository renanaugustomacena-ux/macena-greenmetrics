<script lang="ts">
  import EnergyChart from '$components/EnergyChart.svelte';
  import CarbonFootprint from '$components/CarbonFootprint.svelte';
  import ESGScorecard from '$components/ESGScorecard.svelte';
  import AlarmFeed from '$components/AlarmFeed.svelte';
  import Scope123Breakdown from '$components/Scope123Breakdown.svelte';
  import ConsumptionHeatmap from '$components/ConsumptionHeatmap.svelte';
</script>

<section class="space-y-6">
  <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
    <div class="card">
      <p class="text-xs uppercase text-forest-600">Consumo mese corrente</p>
      <p class="mt-2 text-3xl font-semibold text-forest-900">128.430 <span class="text-base text-forest-700">kWh</span></p>
      <p class="mt-1 text-sm text-forest-600">-4,8% vs baseline</p>
    </div>
    <div class="card">
      <p class="text-xs uppercase text-forest-600">Emissioni YTD</p>
      <p class="mt-2 text-3xl font-semibold text-forest-900">312,6 <span class="text-base text-forest-700">t CO2e</span></p>
      <p class="mt-1 text-sm text-forest-600">Scope 1+2, location-based</p>
    </div>
    <div class="card">
      <p class="text-xs uppercase text-forest-600">Risparmio Piano 5.0</p>
      <p class="mt-2 text-3xl font-semibold text-forest-900">€ 186.400</p>
      <p class="mt-1 text-sm text-forest-600">Credito d'imposta stimato</p>
    </div>
    <div class="card">
      <p class="text-xs uppercase text-forest-600">CSRD Readiness</p>
      <p class="mt-2 text-3xl font-semibold text-forest-900">78%</p>
      <p class="mt-1 text-sm text-forest-600">ESRS E1 data-points coperti</p>
    </div>
  </div>

  <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
    <div class="lg:col-span-2 card">
      <h3 class="text-base font-semibold mb-4">Consumo energetico — ultime 4 settimane</h3>
      <EnergyChart />
    </div>
    <div class="card">
      <h3 class="text-base font-semibold mb-4">Impronta di carbonio</h3>
      <CarbonFootprint />
    </div>
  </div>

  <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
    <div class="card lg:col-span-1">
      <h3 class="text-base font-semibold mb-4">Scope 1 · 2 · 3</h3>
      <Scope123Breakdown />
    </div>
    <div class="card lg:col-span-1">
      <h3 class="text-base font-semibold mb-4">Scorecard ESG</h3>
      <ESGScorecard />
    </div>
    <div class="card lg:col-span-1">
      <h3 class="text-base font-semibold mb-4">Allarmi recenti</h3>
      <AlarmFeed />
    </div>
  </div>

  <div class="card">
    <h3 class="text-base font-semibold mb-4">Mappa di consumo — ora × giorno</h3>
    <ConsumptionHeatmap />
  </div>
</section>
