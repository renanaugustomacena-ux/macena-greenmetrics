<script lang="ts">
  import Scope123Breakdown from '$components/Scope123Breakdown.svelte';
  import CarbonFootprint from '$components/CarbonFootprint.svelte';
  import EmissionFactorEditor from '$components/EmissionFactorEditor.svelte';
</script>

<section class="space-y-6">
  <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
    <div class="card">
      <h3 class="text-base font-semibold mb-4">Scope 1 · 2 · 3 (YTD)</h3>
      <Scope123Breakdown />
    </div>
    <div class="card">
      <h3 class="text-base font-semibold mb-4">Andamento 12 mesi</h3>
      <CarbonFootprint />
    </div>
  </div>

  <div class="card">
    <h3 class="text-base font-semibold mb-4">Fattori di emissione</h3>
    <p class="text-sm text-forest-700 mb-4">
      Factor set corrente: ISPRA 2024 Rapporto 404 + D.M. 11/05/2022. Aggiornamento automatico da feed ISPRA.
    </p>
    <EmissionFactorEditor />
  </div>
</section>
