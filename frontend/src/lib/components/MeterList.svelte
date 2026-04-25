<script lang="ts">
  // SSR data passed as `meters` prop from +page.server.ts (or an empty array
  // when the page hasn't supplied one — pages other than /meters use the
  // hardcoded preview list below).
  export let meters: Array<{
    id: string;
    label?: string;
    meter_type?: string;
    protocol?: string;
    site?: string;
    cost_centre?: string;
    active?: boolean;
    serial_no?: string;
  }> = [];

  $: items = meters && meters.length > 0
    ? meters
    : [
        // Fallback preview (used only if SSR didn't supply data).
        { id: 'CE-001', label: 'Quadro Generale CE-001', meter_type: 'electricity_3p', protocol: 'modbus_tcp', site: 'Stabilimento A', active: true },
        { id: 'CE-002', label: 'Cabina CE-002', meter_type: 'electricity_3p', protocol: 'modbus_tcp', site: 'Stabilimento A', active: true },
        { id: 'GAS-001', label: 'Linea forni', meter_type: 'gas', protocol: 'mbus', site: 'Stabilimento A', active: true }
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
      {#each items as m}
        <tr class="hover:bg-cream-100">
          <td class="py-2 font-mono text-xs">
            {m.serial_no ?? (m.id?.length > 12 ? m.id.slice(0, 8) + '…' + m.id.slice(-4) : m.id)}
          </td>
          <td>{m.label ?? '—'}</td>
          <td>{m.meter_type ?? '—'}</td>
          <td>{m.protocol ?? '—'}</td>
          <td>{m.site ?? '—'}</td>
          <td>
            <span class="badge-scope-2">{m.active === false ? 'inactive' : 'online'}</span>
          </td>
        </tr>
      {/each}
    </tbody>
  </table>
</div>
