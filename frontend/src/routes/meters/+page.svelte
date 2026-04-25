<script lang="ts">
  import MeterList from '$components/MeterList.svelte';
</script>

<section class="space-y-6">
  <div class="flex items-center justify-between">
    <div>
      <h2 class="text-xl font-semibold text-forest-900">Contatori</h2>
      <p class="text-sm text-forest-700">Gestione centralizzata di contatori elettrici, gas, termici, idrici e fotovoltaici.</p>
    </div>
    <button class="btn-primary">+ Nuovo contatore</button>
  </div>
  <MeterList />
</section>
