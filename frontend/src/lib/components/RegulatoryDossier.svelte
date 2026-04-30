<script lang="ts">
  // Card surface for a single Italian regulatory dossier.
  // Shown in a strip on the home dashboard + on /reports.
  //
  // Status: ready | inflight | due | overdue | submitted

  type DossierStatus = 'ready' | 'inflight' | 'due' | 'overdue' | 'submitted';

  export let title: string;
  export let regulator: string;
  export let period: string;
  export let dueDate: string;
  export let status: DossierStatus;
  export let signed = false;
  export let provenanceHash: string | null = null;

  const statusConfig: Record<DossierStatus, { label: string; tone: 'success' | 'warning' | 'danger' | 'neutral' }> = {
    ready:     { label: 'Pronto',     tone: 'success' },
    inflight:  { label: 'In lavorazione', tone: 'neutral' },
    due:       { label: 'In scadenza', tone: 'warning' },
    overdue:   { label: 'Scaduto',     tone: 'danger' },
    submitted: { label: 'Trasmesso',   tone: 'success' }
  };

  $: badgeClass = `badge-${statusConfig[status].tone}`;
</script>

<article class="card-tight flex flex-col gap-3 hover:shadow-elevated transition-shadow">
  <header class="flex items-start justify-between gap-2">
    <div>
      <p class="eyebrow">{regulator}</p>
      <h3 class="mt-1 text-sm font-semibold text-forest-950 leading-tight">{title}</h3>
    </div>
    <span class={badgeClass}>{statusConfig[status].label}</span>
  </header>

  <dl class="grid grid-cols-2 gap-2 text-xs">
    <div>
      <dt class="text-forest-700">Periodo</dt>
      <dd class="font-medium text-forest-900">{period}</dd>
    </div>
    <div>
      <dt class="text-forest-700">Scadenza</dt>
      <dd class="font-medium text-forest-900">{dueDate}</dd>
    </div>
  </dl>

  {#if signed || provenanceHash}
    <footer class="flex items-center justify-between gap-2 pt-2 border-t border-forest-100">
      {#if signed}
        <span class="inline-flex items-center gap-1 text-xs text-forest-700">
          <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2" aria-hidden="true">
            <path stroke-linecap="round" stroke-linejoin="round" d="M9 12l2 2 4-4m5.5 2A9.5 9.5 0 113 12a9.5 9.5 0 0117 0z"/>
          </svg>
          Firmato
        </span>
      {/if}
      {#if provenanceHash}
        <code class="font-mono text-[10px] text-forest-600 truncate" title={provenanceHash}>
          {provenanceHash.slice(0, 12)}…
        </code>
      {/if}
    </footer>
  {/if}
</article>
