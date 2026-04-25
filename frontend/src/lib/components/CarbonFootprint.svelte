<script lang="ts">
  import { onMount } from 'svelte';
  let canvas: HTMLCanvasElement;

  onMount(async () => {
    const { Chart, registerables } = await import('chart.js');
    Chart.register(...registerables);
    new Chart(canvas.getContext('2d')!, {
      type: 'doughnut',
      data: {
        labels: ['Scope 1 (gas, combustibili)', 'Scope 2 (energia elettrica)', 'Scope 3 (catena del valore)'],
        datasets: [
          {
            data: [84, 176, 52],
            backgroundColor: ['#854d0e', '#16a34a', '#d5a337'],
            borderColor: '#ffffff',
            borderWidth: 2
          }
        ]
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          legend: { position: 'bottom', labels: { boxWidth: 12, font: { size: 11 } } }
        },
        cutout: '62%'
      }
    });
  });
</script>

<div class="w-full h-64">
  <canvas bind:this={canvas} aria-label="Torta impronta di carbonio per scope"></canvas>
</div>
<p class="text-xs text-forest-700 mt-4">Totale YTD: 312,6 t CO2e · Fattori: ISPRA 2024 Rapporto 404</p>
