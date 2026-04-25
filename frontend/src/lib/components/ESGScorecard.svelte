<script lang="ts">
  const rows = [
    { label: 'Dati ESRS E1', value: 78 },
    { label: 'Copertura contatori', value: 92 },
    { label: 'Riduzione YoY', value: 4.8 },
    { label: 'Fonti rinnovabili', value: 31 }
  ];
</script>

<div class="space-y-3">
  {#each rows as r}
    <div>
      <div class="flex justify-between text-sm">
        <span class="text-forest-700">{r.label}</span>
        <span class="font-semibold text-forest-900">{r.value}%</span>
      </div>
      <div class="h-2 bg-forest-100 rounded mt-1">
        <div class="h-full rounded bg-forest-700" style="width:{r.value}%"></div>
      </div>
    </div>
  {/each}
</div>
