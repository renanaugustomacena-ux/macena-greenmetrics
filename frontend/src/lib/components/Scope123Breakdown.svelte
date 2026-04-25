<script lang="ts">
  const scopes = [
    { name: 'Scope 1', t: 84, pct: 27, color: 'bg-earth-700' },
    { name: 'Scope 2', t: 176, pct: 56, color: 'bg-forest-700' },
    { name: 'Scope 3', t: 52, pct: 17, color: 'bg-earth-400' }
  ];
</script>

<ul class="space-y-4 text-sm">
  {#each scopes as s}
    <li>
      <div class="flex justify-between">
        <span class="font-semibold">{s.name}</span>
        <span>{s.t} t CO2e</span>
      </div>
      <div class="h-2 mt-1 rounded bg-forest-100">
        <div class="h-full rounded {s.color}" style="width:{s.pct}%"></div>
      </div>
      <p class="text-xs text-forest-600 mt-1">{s.pct}% del totale</p>
    </li>
  {/each}
</ul>
