<script lang="ts">
  const alarms = [
    { id: 'a1', sev: 'warning', msg: 'Consumo anomalo — pressa 200t (+34% vs baseline)', ts: '14:12' },
    { id: 'a2', sev: 'info', msg: 'Fattore di emissione ISPRA 2024 aggiornato', ts: '10:03' },
    { id: 'a3', sev: 'critical', msg: 'Contatore GAS-001 offline da 42 minuti', ts: '09:58' },
    { id: 'a4', sev: 'warning', msg: 'Picco kW contrattuale sfiorato (98%)', ts: 'ieri' },
    { id: 'a5', sev: 'info', msg: 'CSRD: dossier E1 al 78% di completamento', ts: 'ieri' }
  ];
  const col = (s: string) =>
    s === 'critical' ? 'bg-red-100 text-red-800' : s === 'warning' ? 'bg-earth-100 text-earth-900' : 'bg-forest-100 text-forest-800';
</script>

<ul class="divide-y divide-forest-100 text-sm">
  {#each alarms as a}
    <li class="py-2 flex items-start gap-3">
      <span class="text-xs font-semibold px-2 py-0.5 rounded {col(a.sev)}">{a.sev}</span>
      <div class="flex-1">
        <p class="text-forest-900">{a.msg}</p>
        <p class="text-xs text-forest-600">{a.ts}</p>
      </div>
    </li>
  {/each}
</ul>
