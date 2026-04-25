<script lang="ts">
  // Hour (0-23) × weekday heat-map. Colour = kWh intensity.
  const weekdays = ['Lun', 'Mar', 'Mer', 'Gio', 'Ven', 'Sab', 'Dom'];
  const hours = Array.from({ length: 24 }, (_, i) => i);
  function intensity(day: number, hour: number) {
    const isWeekend = day >= 5 ? 0.4 : 1.0;
    const base = hour >= 6 && hour < 22 ? 1 + 0.7 * Math.sin(((hour - 6) / 16) * Math.PI) : 0.35;
    return isWeekend * base;
  }
  function color(v: number) {
    const ramp = Math.max(0, Math.min(1, v / 1.7));
    // Interpolate cream → forest-800.
    const r = Math.round(255 - ramp * (255 - 22));
    const g = Math.round(250 - ramp * (250 - 101));
    const b = Math.round(224 - ramp * (224 - 52));
    return `rgb(${r},${g},${b})`;
  }
</script>

<div class="w-full overflow-x-auto">
  <table class="border-collapse text-xs">
    <thead>
      <tr>
        <th class="text-left pr-2"></th>
        {#each hours as h}<th class="w-6 text-center text-forest-700">{h}</th>{/each}
      </tr>
    </thead>
    <tbody>
      {#each weekdays as d, di}
        <tr>
          <td class="pr-2 text-forest-700">{d}</td>
          {#each hours as h}
            <td class="w-6 h-6" style="background-color: {color(intensity(di, h))};" title={`${d} ${h}:00`}></td>
          {/each}
        </tr>
      {/each}
    </tbody>
  </table>
  <div class="flex items-center gap-2 mt-3 text-xs text-forest-700">
    <span>Basso</span>
    <div class="h-2 w-40 rounded" style="background: linear-gradient(to right, rgb(255,250,224), rgb(22,101,52));"></div>
    <span>Alto</span>
  </div>
</div>
