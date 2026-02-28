<template>
  <div class="settings-section">
    <h3 class="section-title">Mejlinställningar</h3>
    <div class="space-y-4">
      <div class="flex items-center gap-3">
        <input type="checkbox" id="is_active" v-model="settings.is_active" class="w-5 h-5" />
        <label for="is_active" class="label-checkbox">Aktivera mejl</label>
      </div>
      <div class="flex items-center gap-3">
        <input type="checkbox" id="only_enabled_products" v-model="settings.only_enabled_products" class="w-5 h-5" />
        <label for="only_enabled_products" class="label-checkbox">Bara aktiverade produkter</label>
      </div>
      <div>
        <label class="label">Rabatt (%)</label>
        <input type="number" v-model.number="settings.min_discount" class="input w-full" />
      </div>
      <div>
        <label class="label">Vinst (SEK)</label>
        <input type="number" v-model.number="settings.min_profit_sek" class="input w-full" />
      </div>
      <button @click="save" class="btn btn-primary" :disabled="saving">
        {{ saving ? 'Sparar...' : 'Spara' }}
      </button>
      <div v-if="message" class="text-sm" :class="message.includes('Kunde') ? 'text-red-400' : 'text-emerald-400'">
        {{ message }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
const api = useApi()

const settings = ref<{ is_active: boolean; only_enabled_products: boolean; min_discount: number | null; min_profit_sek: number | null }>({
  is_active: true,
  only_enabled_products: true,
  min_discount: null,
  min_profit_sek: null,
})
const saving = ref(false)
const message = ref<string | null>(null)

const fetchSettings = async () => {
  try {
    const data = await api.get<any>('/settings/email')
    settings.value.is_active = data.is_active ?? true
    settings.value.only_enabled_products = data.only_enabled_products ?? true
    settings.value.min_discount = data.min_discount ?? null
    settings.value.min_profit_sek = data.min_profit_sek ?? null
    message.value = null
  } catch (e: any) {
    message.value = 'Kunde inte hämta inställningar'
    console.error(e)
  }
}

const save = async () => {
  saving.value = true
  message.value = null
  try {
    await api.put('/settings/email', {
      is_active: settings.value.is_active,
      only_enabled_products: settings.value.only_enabled_products,
      min_discount: settings.value.min_discount,
      min_profit_sek: settings.value.min_profit_sek,
    })
    message.value = 'Sparat'
  } catch (e: any) {
    message.value = 'Kunde inte spara inställningar'
    console.error(e)
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  fetchSettings()
})
</script>

<style scoped>
.settings-section {
  background: var(--un-prose);
  padding: 1.5rem;
  border-radius: 0.5rem;
  margin-bottom: 1.5rem;
}
.section-title {
  font-size: 1.25rem;
  font-weight: 600;
  margin-bottom: 1rem;
  color: var(--un-prose-lead);
}
.label {
  display: block;
  margin-bottom: 0.5rem;
  color: var(--un-prose-lead);
}
.label-checkbox {
  color: var(--un-prose-lead);
  cursor: pointer;
}
.input {
  padding: 0.5rem;
  border-radius: 0.375rem;
  width: 100%;
  background: var(--un-prose-bg);
  border: 1px solid var(--un-prose-border);
  color: var(--un-prose);
}
.btn {
  padding: 0.5rem 1rem;
  border-radius: 0.375rem;
}
.btn-primary {
  background-color: #059669;
  color: white;
  border: none;
  cursor: pointer;
}
.btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
