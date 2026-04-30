<script lang="ts">
  import '../app.css';
  import { page } from '$app/stores';

  const nav = [
    { href: '/', label: 'Dashboard' },
    { href: '/meters', label: 'Contatori' },
    { href: '/readings', label: 'Letture' },
    { href: '/reports', label: 'Report' },
    { href: '/carbon', label: 'Emissioni' },
    { href: '/settings', label: 'Impostazioni' }
  ];
</script>

<div class="min-h-full flex">
  <aside class="w-64 bg-forest-900 text-cream-50 flex flex-col">
    <div class="p-6 border-b border-forest-800">
      <h1 class="text-xl font-bold tracking-tight">GreenMetrics</h1>
      <p class="text-xs text-forest-200 mt-1">Energia · Emissioni · CSRD</p>
    </div>
    <nav class="flex-1 p-4 space-y-1" aria-label="Navigazione principale">
      {#each nav as item}
        <a
          href={item.href}
          class="block px-3 py-2 rounded-md text-sm transition-colors
                 {$page.url.pathname === item.href
                   ? 'bg-forest-800 text-cream-50 font-semibold'
                   : 'text-forest-200 hover:bg-forest-800 hover:text-cream-50'}"
          aria-current={$page.url.pathname === item.href ? 'page' : undefined}
        >
          {item.label}
        </a>
      {/each}
    </nav>
    <div class="p-4 border-t border-forest-800 text-xs text-forest-300">
      <p>Made in Verona</p>
      <p class="mt-1">v0.1.0</p>
    </div>
  </aside>

  <main class="flex-1 overflow-auto bg-cream-50">
    <header class="bg-white border-b border-forest-100 px-8 py-4 flex items-center justify-between">
      <h2 class="text-lg font-semibold text-forest-900">
        {#if $page.url.pathname === '/'}Panoramica consumi{:else if $page.url.pathname === '/meters'}Contatori{:else if $page.url.pathname === '/readings'}Serie storiche{:else if $page.url.pathname === '/reports'}Reportistica{:else if $page.url.pathname === '/carbon'}Impronta di carbonio{:else}Impostazioni{/if}
      </h2>
      <div class="text-sm text-forest-700">Industria Esempio S.r.l.</div>
    </header>
    <div class="p-8">
      <slot />
    </div>
  </main>
</div>
