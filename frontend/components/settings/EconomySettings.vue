<template>
  <div class="settings-section">
    <h3 class="section-title">Ekonomiinställningar</h3>
    <div class="space-y-4">
      <div>
        <label class="label">Omsättningsdagar</label>
        <input type="number" v-model.number="settings.turnover_days" class="input w-full" />
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

const settings = ref<{ turnover_days: number | null }>({
  turnover_days: null,
})
const saving = ref(false)
const message = ref<string | null>(null)

const fetchSettings = async () => {
  try {
    const data = await api.get<any>('/settings/economy')
    settings.value.turnover_days = data.turnover_days ?? null
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
    await api.put('/settings/economy', {
      turnover_days: settings.value.turnover_days,
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
