<script lang="ts">
  // Top-bar system-health surface. Three signals at a glance:
  //   1. Live data feed (meter ingest pulse)
  //   2. Pack-loader status (loaded vs required)
  //   3. Audit-evidence chain (manifest-lock + report signing)

  export let dataFreshnessSeconds = 7;
  export let packsLoaded = 12;
  export let packsRequired = 12;
  export let auditChainOk = true;

  $: ingestStatus = dataFreshnessSeconds < 30 ? 'live' : dataFreshnessSeconds < 300 ? 'stale' : 'down';
  $: packsOk = packsLoaded === packsRequired;
</script>

<div class="flex items-center gap-6 text-xs">
  <!-- 1. Live data feed -->
  <div class="flex items-center gap-2">
    <span
      class:dot-live={ingestStatus === 'live'}
      class:dot-stale={ingestStatus === 'stale'}
      class:dot-down={ingestStatus === 'down'}
      aria-hidden="true"
    ></span>
    <div>
      <p class="font-medium text-forest-900 leading-none">
        {#if ingestStatus === 'live'}Live{:else if ingestStatus === 'stale'}Ritardato{:else}Offline{/if}
      </p>
      <p class="text-[10px] text-forest-600 mt-0.5">{dataFreshnessSeconds}s · ingest meter</p>
    </div>
  </div>

  <!-- 2. Pack-loader -->
  <div class="flex items-center gap-2 pl-6 border-l border-forest-100">
    <svg class="w-4 h-4 text-forest-700" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.75" aria-hidden="true">
      <path stroke-linecap="round" stroke-linejoin="round" d="M3.75 6.75A.75.75 0 014.5 6h15a.75.75 0 01.75.75v10.5A.75.75 0 0119.5 18h-15a.75.75 0 01-.75-.75V6.75z"/>
      <path stroke-linecap="round" stroke-linejoin="round" d="M8.25 6V4.5A1.5 1.5 0 019.75 3h4.5a1.5 1.5 0 011.5 1.5V6"/>
    </svg>
    <div>
      <p class="font-medium text-forest-900 leading-none">
        Pack <span class="font-mono">{packsLoaded}/{packsRequired}</span>
      </p>
      <p class="text-[10px] {packsOk ? 'text-forest-600' : 'text-saffron-700'} mt-0.5">
        {packsOk ? 'tutti caricati' : 'pack mancante'}
      </p>
    </div>
  </div>

  <!-- 3. Audit-evidence chain -->
  <div class="flex items-center gap-2 pl-6 border-l border-forest-100">
    <svg class="w-4 h-4 {auditChainOk ? 'text-forest-700' : 'text-rosso-700'}" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.75" aria-hidden="true">
      <path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
    </svg>
    <div>
      <p class="font-medium text-forest-900 leading-none">Audit-chain</p>
      <p class="text-[10px] {auditChainOk ? 'text-forest-600' : 'text-rosso-700'} mt-0.5">
        {auditChainOk ? 'firmato + integro' : 'verifica fallita'}
      </p>
    </div>
  </div>
</div>
