<script lang="ts">
  import '../app.css';
  import { page } from '$app/stores';
  import SystemStatus from '$components/SystemStatus.svelte';
  import { productName, productDescription } from '$lib/branding';

  const nav: Array<{ href: string; label: string; group?: string }> = [
    { href: '/',         label: 'Panoramica', group: 'monitoraggio' },
    { href: '/meters',   label: 'Contatori',  group: 'monitoraggio' },
    { href: '/readings', label: 'Letture',    group: 'monitoraggio' },
    { href: '/carbon',   label: 'Emissioni',  group: 'monitoraggio' },
    { href: '/reports',  label: 'Dossier regolatori', group: 'compliance' },
    { href: '/settings', label: 'Impostazioni', group: 'amministrazione' }
  ];

  $: groups = nav.reduce<Record<string, typeof nav>>((acc, item) => {
    const g = item.group ?? 'altro';
    (acc[g] = acc[g] ?? []).push(item);
    return acc;
  }, {});

  const groupLabel: Record<string, string> = {
    monitoraggio:    'Monitoraggio',
    compliance:      'Compliance',
    amministrazione: 'Amministrazione'
  };

  const titleByPath: Record<string, { title: string; subtitle: string }> = {
    '/':         { title: 'Panoramica consumi',     subtitle: 'Stato live, KPI mensili, dossier regolatori' },
    '/meters':   { title: 'Contatori',              subtitle: 'Inventario completo dei contatori e dei canali' },
    '/readings': { title: 'Serie storiche',         subtitle: 'Aggregati 15 min · 1h · 1g · esportazione audit' },
    '/carbon':   { title: 'Impronta di carbonio',   subtitle: 'Scope 1 · 2 · 3 con tracciabilità factor source' },
    '/reports':  { title: 'Dossier regolatori',     subtitle: 'CSRD ESRS E1 · Piano 5.0 · Conto Termico · Audit 102/2014 · TEE' },
    '/settings': { title: 'Impostazioni',           subtitle: 'White-label, identità, factor source, tenant' }
  };

  $: header = titleByPath[$page.url.pathname] ?? { title: productName, subtitle: '' };
</script>

<svelte:head>
  <title>{header.title} · {productName}</title>
  <meta name="description" content={productDescription} />
</svelte:head>

<div class="min-h-full flex bg-cream-50">
  <!-- ── Sidebar ───────────────────────────────────────────────────── -->
  <aside class="w-64 bg-forest-950 text-cream-50 flex flex-col shrink-0">
    <div class="p-5 border-b border-forest-900">
      <div class="flex items-center gap-2.5">
        <div class="w-8 h-8 rounded-md bg-forest-700 flex items-center justify-center">
          <svg class="w-4 h-4 text-cream-50" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.25" aria-hidden="true">
            <path stroke-linecap="round" stroke-linejoin="round" d="M3.75 12.75A2.25 2.25 0 016 10.5h3a.75.75 0 01.75.75v9.75a.75.75 0 01-.75.75H6a2.25 2.25 0 01-2.25-2.25v-6zM14.25 6.75A2.25 2.25 0 0116.5 4.5h3a.75.75 0 01.75.75v15.75a.75.75 0 01-.75.75h-3a2.25 2.25 0 01-2.25-2.25V6.75z"/>
          </svg>
        </div>
        <div>
          <h1 class="text-base font-semibold tracking-tight">{productName}</h1>
          <p class="text-[10px] text-forest-300 uppercase tracking-wider mt-0.5">Italian flagship</p>
        </div>
      </div>
    </div>

    <nav class="flex-1 p-4 space-y-5 overflow-y-auto" aria-label="Navigazione principale">
      {#each Object.entries(groups) as [groupKey, items]}
        <div>
          <p class="px-3 mb-1.5 text-[10px] uppercase tracking-wider font-semibold text-forest-400">
            {groupLabel[groupKey] ?? groupKey}
          </p>
          <ul class="space-y-0.5">
            {#each items as item}
              <li>
                <a
                  href={item.href}
                  class="block px-3 py-1.5 rounded text-sm transition-colors
                         {$page.url.pathname === item.href
                           ? 'bg-forest-800 text-cream-50 font-medium'
                           : 'text-forest-200 hover:bg-forest-900 hover:text-cream-50'}"
                  aria-current={$page.url.pathname === item.href ? 'page' : undefined}
                >
                  {item.label}
                </a>
              </li>
            {/each}
          </ul>
        </div>
      {/each}
    </nav>

    <div class="p-4 border-t border-forest-900 text-[10px] text-forest-300 space-y-1">
      <p class="text-forest-200 font-medium">Industria Esempio S.r.l.</p>
      <p>Tenant <code class="font-mono text-forest-400">acme-2026</code></p>
      <p>Topology A · eu-south-1</p>
      <p class="pt-2 border-t border-forest-900 mt-2">v0.1.0 · Sprint S5</p>
    </div>
  </aside>

  <!-- ── Main column ───────────────────────────────────────────────── -->
  <div class="flex-1 flex flex-col min-w-0">
    <header class="bg-white border-b border-forest-100 shadow-card">
      <div class="px-8 py-3 flex items-center justify-between gap-4 border-b border-forest-100/60">
        <SystemStatus />
        <div class="flex items-center gap-3">
          <button class="btn-ghost btn-sm">
            <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.75" aria-hidden="true">
              <path stroke-linecap="round" stroke-linejoin="round" d="M14.857 17.082a23.848 23.848 0 005.454-1.31A8.967 8.967 0 0118 9.75v-.7V9A6 6 0 006 9v.75a8.967 8.967 0 01-2.312 6.022c1.733.64 3.56 1.085 5.455 1.31m5.714 0a24.255 24.255 0 01-5.714 0m5.714 0a3 3 0 11-5.714 0"/>
            </svg>
            <span class="font-mono text-[11px]">3</span>
          </button>
          <div class="flex items-center gap-2 text-sm">
            <div class="w-7 h-7 rounded-full bg-forest-100 flex items-center justify-center">
              <span class="text-[11px] font-semibold text-forest-900">RM</span>
            </div>
            <div>
              <p class="font-medium text-forest-900 leading-none">Renan Macena</p>
              <p class="text-[10px] text-forest-600 mt-0.5">platform-admin</p>
            </div>
          </div>
        </div>
      </div>
      <div class="px-8 py-4 flex items-end justify-between gap-4">
        <div>
          <h2 class="text-xl font-semibold text-forest-950 tracking-tight">{header.title}</h2>
          {#if header.subtitle}
            <p class="text-sm text-forest-700 mt-0.5">{header.subtitle}</p>
          {/if}
        </div>
        <p class="text-xs text-forest-600 font-mono">
          {new Date().toLocaleDateString('it-IT', { day: '2-digit', month: 'long', year: 'numeric' })}
        </p>
      </div>
    </header>

    <main class="flex-1 overflow-auto bg-cream-50">
      <div class="p-8 max-w-screen-2xl mx-auto">
        <slot />
      </div>
    </main>
  </div>
</div>
