<script lang="ts">
  // Pack-loader status panel — shows the loaded Pack catalogue grouped by
  // kind. Surfaces the audit-grade evidence: which Region / Protocol /
  // Factor / Report / Identity Packs the deployment is running, and at
  // which versions. Reads `manifest.lock.json` at runtime in production;
  // here demo-populated.

  interface Pack {
    id: string;
    kind: 'region' | 'factor' | 'report' | 'protocol' | 'identity';
    version: string;
    contractVersion: string;
  }

  const packs: Pack[] = [
    { id: 'region-it',         kind: 'region',   version: '1.0.0', contractVersion: '1.0.0' },
    { id: 'factor-ispra',      kind: 'factor',   version: '2026.04.01', contractVersion: '1.0.0' },
    { id: 'factor-gse',        kind: 'factor',   version: '2026.03.15', contractVersion: '1.0.0' },
    { id: 'factor-terna',      kind: 'factor',   version: '2026.04.30', contractVersion: '1.0.0' },
    { id: 'report-esrs_e1',    kind: 'report',   version: '1.2.0', contractVersion: '1.0.0' },
    { id: 'report-piano_5_0',  kind: 'report',   version: '1.1.3', contractVersion: '1.0.0' },
    { id: 'report-conto_termico', kind: 'report', version: '1.0.5', contractVersion: '1.0.0' },
    { id: 'report-audit_dlgs102', kind: 'report', version: '1.0.2', contractVersion: '1.0.0' },
    { id: 'protocol-modbus_tcp',  kind: 'protocol', version: '1.0.0', contractVersion: '1.0.0' },
    { id: 'protocol-mbus',     kind: 'protocol', version: '1.0.0', contractVersion: '1.0.0' },
    { id: 'protocol-sunspec',  kind: 'protocol', version: '1.0.0', contractVersion: '1.0.0' },
    { id: 'identity-local_db', kind: 'identity', version: '1.0.0', contractVersion: '1.0.0' }
  ];

  const kindLabel: Record<Pack['kind'], string> = {
    region:   'Region',
    factor:   'Factor',
    report:   'Report',
    protocol: 'Protocol',
    identity: 'Identity'
  };

  const kindColor: Record<Pack['kind'], string> = {
    region:   'bg-forest-100 text-forest-900',
    factor:   'bg-earth-100 text-earth-900',
    report:   'bg-cream-200 text-earth-900',
    protocol: 'bg-saffron-100 text-saffron-900',
    identity: 'bg-stone-100 text-stone-800'
  };

  // Manifest-lock hash (the cryptographic anchor — Rule 73). Fixed for the demo.
  const lockHash = 'sha256:9b3c8e2f4a5b6c7d8e9f0a1b2c3d4e5f';
</script>

<div class="space-y-3">
  <header class="flex items-center justify-between">
    <div>
      <p class="eyebrow">manifest.lock.json</p>
      <code class="font-mono text-xs text-forest-700">{lockHash}</code>
    </div>
    <span class="badge-success">12/12 caricati</span>
  </header>

  <ul class="space-y-1.5 max-h-64 overflow-y-auto pr-1">
    {#each packs as pack}
      <li class="flex items-center justify-between gap-3 py-1.5 px-2 rounded hover:bg-cream-50">
        <div class="flex items-center gap-2 min-w-0">
          <span class="text-[10px] uppercase tracking-wide font-medium px-1.5 py-0.5 rounded {kindColor[pack.kind]} shrink-0">
            {kindLabel[pack.kind]}
          </span>
          <code class="font-mono text-xs text-forest-900 truncate">{pack.id}</code>
        </div>
        <div class="text-right shrink-0">
          <p class="font-mono text-xs text-forest-900">{pack.version}</p>
          <p class="font-mono text-[10px] text-forest-600">contract {pack.contractVersion}</p>
        </div>
      </li>
    {/each}
  </ul>
</div>
