<script lang="ts">
  import { onMount } from 'svelte';
  let canvas: HTMLCanvasElement;

  // Synthetic demo data — replace with queryAggregated() in production.
  const weeks = 4;
  const hoursPerWeek = 24 * 7;
  const values: number[] = [];
  for (let i = 0; i < weeks * hoursPerWeek; i++) {
    const hour = i % 24;
    const weekday = Math.floor((i / 24) % 7);
    const base = weekday < 5 ? 480 : 220;
    const shape = hour >= 6 && hour <= 22 ? 1 + 0.8 * Math.sin(((hour - 6) / 16) * Math.PI) : 0.55;
    values.push(Math.round(base * shape + (Math.random() - 0.5) * 50));
  }

  onMount(async () => {
    const { Chart, registerables } = await import('chart.js');
    Chart.register(...registerables);
    new Chart(canvas.getContext('2d')!, {
      type: 'line',
      data: {
        labels: values.map((_, i) => `h${i}`),
        datasets: [
          {
            label: 'kWh',
            data: values,
            borderColor: '#166534',
            backgroundColor: 'rgba(22,101,52,0.15)',
            fill: true,
            tension: 0.25,
            pointRadius: 0,
            borderWidth: 2
          }
        ]
      },
      options: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: { legend: { display: false } },
        scales: {
          x: { ticks: { display: false }, grid: { display: false } },
          y: { ticks: { color: '#14532d' }, grid: { color: 'rgba(22,101,52,0.08)' } }
        }
      }
    });
  });
</script>

<div class="w-full h-64">
  <canvas bind:this={canvas} aria-label="Grafico a linee consumo energetico"></canvas>
</div>
