<script lang="ts">
</script>

<section class="space-y-6">
  <div class="card">
    <h3 class="text-base font-semibold mb-4">Azienda</h3>
    <form class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <label class="block">
        <span class="text-xs text-forest-700">Ragione sociale</span>
        <input class="mt-1 w-full border rounded px-3 py-2" value="Industria Esempio S.r.l." />
      </label>
      <label class="block">
        <span class="text-xs text-forest-700">Partita IVA</span>
        <input class="mt-1 w-full border rounded px-3 py-2" value="IT01234567890" />
      </label>
      <label class="block">
        <span class="text-xs text-forest-700">Codice ATECO</span>
        <input class="mt-1 w-full border rounded px-3 py-2" value="10.89.09" />
      </label>
      <label class="block">
        <span class="text-xs text-forest-700">PEC</span>
        <input class="mt-1 w-full border rounded px-3 py-2" value="industria.esempio@pec.it" />
      </label>
    </form>
  </div>

  <div class="card">
    <h3 class="text-base font-semibold mb-4">Normative applicabili</h3>
    <ul class="space-y-2 text-sm">
      <li><input type="checkbox" checked class="accent-forest-700 mr-2" /> CSRD / ESRS E1 — obbligo dichiarazione di sostenibilità</li>
      <li><input type="checkbox" checked class="accent-forest-700 mr-2" /> D.Lgs. 102/2014 — audit energetico quadriennale</li>
      <li><input type="checkbox" checked class="accent-forest-700 mr-2" /> Piano Transizione 5.0 — candidatura credito d'imposta</li>
      <li><input type="checkbox" class="accent-forest-700 mr-2" /> Conto Termico 2.0 — domande GSE</li>
      <li><input type="checkbox" class="accent-forest-700 mr-2" /> Certificati Bianchi TEE</li>
      <li><input type="checkbox" class="accent-forest-700 mr-2" /> ETS — cap-and-trade (se applicabile)</li>
    </ul>
  </div>

  <div class="card">
    <h3 class="text-base font-semibold mb-4">Integrazioni DSO</h3>
    <p class="text-sm text-forest-700 mb-4">Configura accessi ai portali di distribuzione per importare curve POD/PDR.</p>
    <div class="space-y-3 text-sm">
      <div class="flex items-center justify-between">
        <span>E-Distribuzione SMD</span>
        <button class="btn-secondary">Connetti</button>
      </div>
      <div class="flex items-center justify-between">
        <span>Terna Transparency</span>
        <button class="btn-secondary">Connetti</button>
      </div>
      <div class="flex items-center justify-between">
        <span>SPD multi-DSO</span>
        <span class="text-forest-500">Certificato cliente richiesto</span>
      </div>
    </div>
  </div>
</section>
