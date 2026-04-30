<script lang="ts">
  import RegulatoryDossier from '$components/RegulatoryDossier.svelte';

  // Active dossiers (status-driven; mirrors the home-page strip).
  const activeDossiers = [
    {
      title: 'ESRS E1 — Cambiamenti climatici',
      regulator: 'EFRAG · CSRD wave 2',
      period: 'FY 2025',
      dueDate: '2026-04-30',
      status: 'ready' as const,
      signed: true,
      provenanceHash: 'sha256:7a1c4f9d2b8e6c3f'
    },
    {
      title: 'Piano Transizione 5.0',
      regulator: 'GSE · MIMIT-MASE',
      period: 'Q2 2026',
      dueDate: '2026-06-30',
      status: 'inflight' as const,
      signed: false,
      provenanceHash: 'sha256:3e9b8a5c1f4d7e2a'
    },
    {
      title: 'Conto Termico 2.0',
      regulator: 'GSE',
      period: 'Apr 2026',
      dueDate: '2026-05-31',
      status: 'due' as const,
      signed: false,
      provenanceHash: null
    },
    {
      title: 'Audit Energetico D.Lgs. 102/2014',
      regulator: 'ENEA',
      period: '2024–2027',
      dueDate: '2027-12-05',
      status: 'inflight' as const,
      signed: true,
      provenanceHash: null
    },
    {
      title: 'Certificati Bianchi TEE',
      regulator: 'GSE',
      period: 'Q1 2026',
      dueDate: '2026-05-15',
      status: 'due' as const,
      signed: false,
      provenanceHash: null
    },
    {
      title: 'Carbon Footprint mensile',
      regulator: 'ISO 14064-1',
      period: 'Apr 2026',
      dueDate: '2026-05-10',
      status: 'ready' as const,
      signed: true,
      provenanceHash: 'sha256:c4d5a6b7e8f9a0b1'
    }
  ];

  // Submitted-history rows. In production: pulled from `reports` Hypertable.
  interface HistoryRow {
    type: string;
    period: string;
    state: 'firmato' | 'trasmesso' | 'accettato';
    submittedAt: string;
    bytes: string;
    docHash: string;
    factorPack: string;
  }

  const history: HistoryRow[] = [
    { type: 'ESRS E1 (CSRD)',           period: 'FY 2024',  state: 'firmato',    submittedAt: '2025-04-12', bytes: '4.2 MB', docHash: 'sha256:8b2e1f4a',  factorPack: 'ispra@2025.04.01' },
    { type: 'Piano 5.0 attestazione',   period: 'Q1 2026',  state: 'accettato',  submittedAt: '2026-04-02', bytes: '1.8 MB', docHash: 'sha256:3e9b8a5c',  factorPack: 'gse@2026.03.15'   },
    { type: 'Conto Termico 2.0',        period: 'Q4 2025',  state: 'trasmesso',  submittedAt: '2026-01-31', bytes: '0.9 MB', docHash: 'sha256:f1c8d2e7',  factorPack: 'gse@2025.10.01'   },
    { type: 'Audit D.Lgs. 102/2014',    period: '2020–2024', state: 'trasmesso', submittedAt: '2024-11-30', bytes: '12.4 MB', docHash: 'sha256:9a4b7e3d', factorPack: 'ispra@2024.04.01' },
    { type: 'Certificati Bianchi TEE',  period: 'Q4 2025',  state: 'firmato',    submittedAt: '2026-02-14', bytes: '0.6 MB', docHash: 'sha256:5d6e7f8a',  factorPack: 'gse@2025.10.01'   },
    { type: 'Carbon Footprint',         period: 'Mar 2026', state: 'firmato',    submittedAt: '2026-04-10', bytes: '0.4 MB', docHash: 'sha256:c4d5a6b7',  factorPack: 'ispra@2026.04.01' }
  ];

  const stateBadge: Record<HistoryRow['state'], string> = {
    firmato:   'badge-success',
    trasmesso: 'badge-success',
    accettato: 'badge-success'
  };
</script>

<section class="space-y-8">
  <!-- ── Active dossiers ───────────────────────────────────────────── -->
  <section>
    <header class="flex items-end justify-between mb-4">
      <div>
        <p class="eyebrow">Stato corrente</p>
        <h2 class="mt-0.5">Dossier attivi</h2>
        <p class="muted mt-1">Sei dossier nel ciclo regolatorio italiano. Stato live, scadenze, firma digitale e hash di provenienza.</p>
      </div>
      <button class="btn-primary">
        <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2" aria-hidden="true">
          <path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15"/>
        </svg>
        Nuovo dossier
      </button>
    </header>
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      {#each activeDossiers as d}
        <RegulatoryDossier {...d} />
      {/each}
    </div>
  </section>

  <!-- ── Audit-grade evidence band ─────────────────────────────────── -->
  <section class="card-accent">
    <header class="flex items-start justify-between gap-6">
      <div>
        <p class="eyebrow text-forest-700">Audit-grade reproducibility · Doctrine Rules 89, 91, 131, 144</p>
        <h2 class="mt-1">Cosa accompagna ogni dossier</h2>
      </div>
    </header>
    <div class="mt-5 grid grid-cols-1 md:grid-cols-4 gap-4">
      <div class="bg-white/70 rounded-lg p-4 border border-forest-100">
        <p class="text-xs font-medium text-forest-700">PDF firmato</p>
        <p class="muted text-xs mt-1.5">PDF/A-2b deterministico, firma del deployment via KMS. Verifica con <code class="font-mono">cosign verify-blob</code>.</p>
      </div>
      <div class="bg-white/70 rounded-lg p-4 border border-forest-100">
        <p class="text-xs font-medium text-forest-700">Provenance bundle</p>
        <p class="muted text-xs mt-1.5"><code class="font-mono">provenance.json</code> con period, factor-pack versions, manifest_lock_hash, code_hash, report_pack_version.</p>
      </div>
      <div class="bg-white/70 rounded-lg p-4 border border-forest-100">
        <p class="text-xs font-medium text-forest-700">Validazione formale</p>
        <p class="muted text-xs mt-1.5">Output validato contro lo schema del regolatore (EFRAG XBRL · GSE XSD · ENEA XSD) prima della firma.</p>
      </div>
      <div class="bg-white/70 rounded-lg p-4 border border-forest-100">
        <p class="text-xs font-medium text-forest-700">Replay determinismo</p>
        <p class="muted text-xs mt-1.5">Re-esecuzione del Builder produce lo stesso byte-stream. Conformance suite ne è la prova.</p>
      </div>
    </div>
  </section>

  <!-- ── History table ─────────────────────────────────────────────── -->
  <section class="card">
    <header class="flex items-end justify-between mb-4">
      <div>
        <h3>Cronologia trasmissioni</h3>
        <p class="muted mt-0.5">Ogni riga è ricostruibile bit-per-bit dal proprio <code class="font-mono">manifest_lock</code>.</p>
      </div>
      <button class="btn-secondary btn-sm">
        <svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2" aria-hidden="true">
          <path stroke-linecap="round" stroke-linejoin="round" d="M3 16.5v2.25A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75V16.5m-13.5-9L12 3m0 0l4.5 4.5M12 3v13.5"/>
        </svg>
        Esporta evidence pack
      </button>
    </header>
    <div class="overflow-x-auto">
      <table class="table-clean">
        <thead>
          <tr>
            <th>Dossier</th>
            <th>Periodo</th>
            <th>Stato</th>
            <th>Trasmesso</th>
            <th>Dimensione</th>
            <th>Hash documento</th>
            <th>Factor pack</th>
            <th class="text-right">Azioni</th>
          </tr>
        </thead>
        <tbody>
          {#each history as row}
            <tr>
              <td class="font-medium">{row.type}</td>
              <td>{row.period}</td>
              <td><span class={stateBadge[row.state]}>{row.state}</span></td>
              <td class="font-mono text-xs">{row.submittedAt}</td>
              <td class="text-forest-700">{row.bytes}</td>
              <td><code class="font-mono text-[11px] text-forest-700">{row.docHash}</code></td>
              <td><code class="font-mono text-[11px] text-forest-700">{row.factorPack}</code></td>
              <td class="text-right">
                <button class="btn-ghost btn-sm" type="button" aria-label="Scarica PDF firmato">PDF</button>
                <button class="btn-ghost btn-sm" type="button" aria-label="Scarica provenance.json">JSON</button>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  </section>
</section>
