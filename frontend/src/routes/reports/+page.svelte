<script lang="ts">
  const reports = [
    { label: 'Consumo mensile', type: 'monthly_consumption', desc: 'Energia per centro di costo, linea, turno.' },
    { label: 'CO2 Footprint', type: 'co2_footprint', desc: 'Scope 1 + 2 + 3 in kg CO2e per periodo.' },
    { label: 'ESRS E1 (CSRD)', type: 'esrs_e1_csrd', desc: 'Dossier climate change — direttiva UE 2022/2464.' },
    { label: 'Piano Transizione 5.0', type: 'piano_5_0_attestazione', desc: 'Attestazione risparmio energetico 3% / 5% per credito 5–45%.' },
    { label: 'Conto Termico 2.0', type: 'conto_termico_2_0', desc: 'Domanda GSE per incentivi termici rinnovabili.' },
    { label: 'Certificati Bianchi TEE', type: 'certificati_bianchi_tee', desc: 'Titoli di Efficienza Energetica — submission GSE.' },
    { label: 'Audit D.Lgs. 102/2014', type: 'audit_dlgs_102_2014', desc: 'Diagnosi energetica quadriennale — portale ENEA.' }
  ];
</script>

<section class="space-y-6">
  <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
    {#each reports as r}
      <div class="card">
        <div class="flex items-start justify-between">
          <div>
            <h3 class="text-base font-semibold">{r.label}</h3>
            <p class="text-sm text-forest-700 mt-1">{r.desc}</p>
          </div>
          <button class="btn-primary">Genera</button>
        </div>
      </div>
    {/each}
  </div>

  <div class="card">
    <h3 class="text-base font-semibold mb-4">Cronologia report</h3>
    <table class="min-w-full text-sm">
      <thead class="text-xs uppercase text-forest-700">
        <tr><th class="text-left py-2">Tipo</th><th class="text-left">Periodo</th><th class="text-left">Stato</th><th class="text-left">Data</th><th></th></tr>
      </thead>
      <tbody class="divide-y divide-forest-100">
        <tr><td class="py-2">ESRS E1 (CSRD)</td><td>2025</td><td><span class="badge-scope-2">generato</span></td><td>2026-03-15</td><td><a class="text-forest-700 underline" href="#">Scarica</a></td></tr>
        <tr><td class="py-2">Piano 5.0</td><td>Q1 2026</td><td><span class="badge-scope-2">accettato</span></td><td>2026-04-02</td><td><a class="text-forest-700 underline" href="#">Scarica</a></td></tr>
        <tr><td class="py-2">Audit 102/2014</td><td>2024</td><td><span class="badge-scope-1">inviato</span></td><td>2024-11-30</td><td><a class="text-forest-700 underline" href="#">Scarica</a></td></tr>
      </tbody>
    </table>
  </div>
</section>
