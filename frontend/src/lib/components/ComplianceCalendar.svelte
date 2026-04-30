<script lang="ts">
  // Compliance calendar — next 90 days of Italian regulatory deadlines.
  // Driven by region-pack metadata; here pre-populated for the demo flow.

  interface Deadline {
    date: string;
    daysAway: number;
    title: string;
    regulator: string;
    severity: 'high' | 'medium' | 'low';
  }

  const deadlines: Deadline[] = [
    { date: '2026-05-15', daysAway: 14,  title: 'Trasmissione TEE Q1',         regulator: 'GSE',          severity: 'high' },
    { date: '2026-05-31', daysAway: 30,  title: 'Aggiornamento ISPRA factor table', regulator: 'ISPRA',     severity: 'medium' },
    { date: '2026-06-30', daysAway: 60,  title: 'Saldo Piano 5.0 Q2',           regulator: 'GSE/MIMIT',    severity: 'high' },
    { date: '2026-07-31', daysAway: 91,  title: 'Aggiornamento Conto Termico 2.0', regulator: 'GSE',        severity: 'medium' }
  ];

  const severityClass: Record<Deadline['severity'], string> = {
    high:   'border-l-rosso-500',
    medium: 'border-l-saffron-500',
    low:    'border-l-forest-500'
  };
</script>

<div class="space-y-2">
  {#each deadlines as item}
    <div class="flex items-start justify-between gap-3 py-2 pl-3 pr-2 border-l-2 {severityClass[item.severity]} bg-cream-50/40 rounded-r">
      <div class="flex-1 min-w-0">
        <p class="text-sm font-medium text-forest-950 leading-tight truncate">{item.title}</p>
        <p class="text-xs text-forest-700 mt-0.5">
          {item.regulator}
          <span class="text-forest-500">·</span>
          {item.date}
        </p>
      </div>
      <div class="text-right shrink-0">
        <p class="text-base font-semibold text-forest-950">{item.daysAway}</p>
        <p class="text-[10px] uppercase tracking-wide text-forest-700">giorni</p>
      </div>
    </div>
  {/each}
</div>
