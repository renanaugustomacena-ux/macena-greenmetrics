<script lang="ts">
  const factors = [
    { code: 'IT_ELEC_MIX_2024', scope: 2, unit: 'kWh', value: 0.245, source: 'ISPRA 2024 (provvisorio)', validFrom: '2024-01-01' },
    { code: 'IT_ELEC_MIX_2023', scope: 2, unit: 'kWh', value: 0.250, source: 'ISPRA Rapporto 404', validFrom: '2023-01-01' },
    { code: 'NG_STATIONARY_COMBUSTION', scope: 1, unit: 'Sm3', value: 1.975, source: 'D.M. 11/05/2022', validFrom: '2022-05-11' },
    { code: 'DIESEL_COMBUSTION', scope: 1, unit: 'L', value: 2.650, source: 'ISPRA 2024', validFrom: '2024-01-01' },
    { code: 'LPG_COMBUSTION', scope: 1, unit: 'L', value: 1.510, source: 'ISPRA 2024', validFrom: '2024-01-01' },
    { code: 'HEATING_OIL_COMBUSTION', scope: 1, unit: 'L', value: 2.771, source: 'ISPRA 2024', validFrom: '2024-01-01' },
    { code: 'DISTRICT_HEAT_AVERAGE_IT', scope: 2, unit: 'kWh', value: 0.200, source: 'GSE Conto Termico 2.0', validFrom: '2024-01-01' }
  ];
</script>

<table class="min-w-full text-sm">
  <thead class="text-xs uppercase text-forest-700 border-b border-forest-100">
    <tr>
      <th class="text-left py-2">Codice</th>
      <th class="text-left">Scope</th>
      <th class="text-left">Unità</th>
      <th class="text-right">kg CO2e / unità</th>
      <th class="text-left">Fonte</th>
      <th class="text-left">Valido dal</th>
    </tr>
  </thead>
  <tbody class="divide-y divide-forest-50">
    {#each factors as f}
      <tr>
        <td class="py-2 font-mono text-xs">{f.code}</td>
        <td>
          {#if f.scope === 1}<span class="badge-scope-1">Scope 1</span>
          {:else if f.scope === 2}<span class="badge-scope-2">Scope 2</span>
          {:else}<span class="badge-scope-3">Scope 3</span>{/if}
        </td>
        <td>{f.unit}</td>
        <td class="text-right font-mono">{f.value.toFixed(3)}</td>
        <td>{f.source}</td>
        <td>{f.validFrom}</td>
      </tr>
    {/each}
  </tbody>
</table>
