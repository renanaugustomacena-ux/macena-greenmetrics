<script lang="ts">
  const meters = [
    { id: 'CE-001', label: 'Quadro Generale CE-001', type: 'electricity_3p', protocol: 'modbus_tcp', site: 'Stabilimento A', status: 'online' },
    { id: 'CE-002', label: 'Cabina CE-002', type: 'electricity_3p', protocol: 'modbus_tcp', site: 'Stabilimento A', status: 'online' },
    { id: 'GAS-001', label: 'Linea forni', type: 'gas', protocol: 'mbus', site: 'Stabilimento A', status: 'online' },
    { id: 'TH-001', label: 'Centrale termica', type: 'thermal', protocol: 'mbus', site: 'Stabilimento B', status: 'online' },
    { id: 'PV-001', label: 'Fotovoltaico tetto nord', type: 'pv_inverter', protocol: 'sunspec', site: 'Stabilimento A', status: 'online' },
    { id: 'EV-001', label: 'Colonnina OCPP ingresso', type: 'ev_charger', protocol: 'ocpp', site: 'Stabilimento A', status: 'idle' },
    { id: 'H2O-001', label: 'Contatore acqua produttivo', type: 'water', protocol: 'pulse', site: 'Stabilimento A', status: 'online' },
    { id: 'CE-101', label: 'Pressa 200t', type: 'electricity_3p', protocol: 'modbus_rtu', site: 'Stabilimento A', status: 'online' }
  ];
</script>

<div class="card overflow-x-auto">
  <table class="min-w-full text-sm">
    <thead class="text-xs uppercase text-forest-700 border-b border-forest-100">
      <tr>
        <th class="text-left py-2">ID</th>
        <th class="text-left">Etichetta</th>
        <th class="text-left">Tipo</th>
        <th class="text-left">Protocollo</th>
        <th class="text-left">Sito</th>
        <th class="text-left">Stato</th>
      </tr>
    </thead>
    <tbody class="divide-y divide-forest-50">
      {#each meters as m}
        <tr class="hover:bg-cream-100">
          <td class="py-2 font-mono">{m.id}</td>
          <td>{m.label}</td>
          <td>{m.type}</td>
          <td>{m.protocol}</td>
          <td>{m.site}</td>
          <td>
            <span class="badge-scope-2">{m.status}</span>
          </td>
        </tr>
      {/each}
    </tbody>
  </table>
</div>
