<script lang="ts">
  import MeterList from '$components/MeterList.svelte';
  import type { PageData } from './$types';

  export let data: PageData;
  $: meters = data?.meters ?? [];
  $: error = data?.error ?? null;
</script>

<section class="space-y-6">
  <div class="flex items-center justify-between">
    <div>
      <h2 class="text-xl font-semibold text-forest-900">Contatori</h2>
      <p class="text-sm text-forest-700">Gestione centralizzata di contatori elettrici, gas, termici, idrici e fotovoltaici.</p>
    </div>
    <button class="btn-primary">+ Nuovo contatore</button>
  </div>

  {#if error}
    <div class="card bg-red-50 border border-red-200 text-red-900 p-4 text-sm">
      <p class="font-semibold">Errore caricamento dati live (mostrando preview):</p>
      <p class="font-mono text-xs mt-1">{error}</p>
    </div>
  {:else if meters.length > 0}
    <div class="card bg-forest-50 border border-forest-200 text-forest-900 p-3 text-sm">
      <span class="font-semibold">{meters.length}</span> contatori live dal database (tenant dev).
    </div>
  {/if}

  <MeterList {meters} />
</section>
