<script lang="ts">
  import EnergyChart from '$components/EnergyChart.svelte';
  let resolution: '15min' | '1h' | '1d' = '1h';
</script>

<section class="space-y-6">
  <div class="card">
    <div class="flex flex-wrap items-end gap-4">
      <div>
        <label class="block text-xs text-forest-700 mb-1" for="meter">Contatore</label>
        <select id="meter" class="border rounded px-3 py-2">
          <option>Quadro Generale CE-001</option>
          <option>Linea Forni GAS-001</option>
          <option>Cabina CE-002</option>
          <option>FV tetto nord</option>
        </select>
      </div>
      <div>
        <label class="block text-xs text-forest-700 mb-1" for="from">Dal</label>
        <input id="from" type="date" class="border rounded px-3 py-2" />
      </div>
      <div>
        <label class="block text-xs text-forest-700 mb-1" for="to">Al</label>
        <input id="to" type="date" class="border rounded px-3 py-2" />
      </div>
      <div>
        <label class="block text-xs text-forest-700 mb-1" for="res">Risoluzione</label>
        <select id="res" bind:value={resolution} class="border rounded px-3 py-2">
          <option value="15min">15 minuti</option>
          <option value="1h">1 ora</option>
          <option value="1d">1 giorno</option>
        </select>
      </div>
      <button class="btn-primary">Applica</button>
      <button class="btn-secondary">Esporta CSV</button>
    </div>
  </div>

  <div class="card">
    <h3 class="text-base font-semibold mb-4">Andamento — risoluzione {resolution}</h3>
    <EnergyChart />
  </div>
</section>
