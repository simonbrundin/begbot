<template>
  <div class="settings-section">
    <h3 class="section-title">Auto Enable Produkt inställningar</h3>
    <div class="space-y-4">
      <div>
        <label class="label">Värderingssäkerhet (%)</label>
        <input type="number" v-model.number="settings.min_confidence" class="input w-full" />
      </div>
      <div>
        <label class="label">Värde</label>
        <input type="number" v-model.number="settings.value" class="input w-full" />
      </div>
      <div>
        <label class="label">Minst antal annonser</label>
        <input type="number" v-model.number="settings.min_ads" class="input w-full" />
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

const settings = ref<{ min_confidence: number | null; value: number | null; min_ads: number | null }>({
  min_confidence: null,
  value: null,
  min_ads: null,
})
const saving = ref(false)
const message = ref<string | null>(null)

const fetchSettings = async () => {
  try {
    const data = await api.get<any>('/settings/auto-enable')
    settings.value.min_confidence = data.min_confidence ?? null
    settings.value.value = data.value ?? null
    settings.value.min_ads = data.min_ads ?? null
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
    await api.put('/settings/auto-enable', {
      min_confidence: settings.value.min_confidence,
      value: settings.value.value,
      min_ads: settings.value.min_ads,
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
