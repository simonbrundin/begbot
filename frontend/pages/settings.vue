<template>
  <div>
    <h2 class="text-2xl font-bold mb-4">Inställningar</h2>
    <form @submit.prevent="save">
      <div class="space-y-4 max-w-md">
        <div>
          <label class="label">Minsta vinst (SEK)</label>
          <input type="number" v-model.number="rules.min_profit_sek" class="input w-full" />
        </div>
        <div>
          <label class="label">Minsta rabatt (%)</label>
          <input type="number" v-model.number="rules.min_discount" class="input w-full" />
        </div>
        <div>
          <label class="label">Minsta säkerhetsprocent (%)</label>
          <input type="number" v-model.number="rules.min_confidence" class="input w-full" />
        </div>
        <div>
          <label class="label">Minsta antal annonser</label>
          <input type="number" v-model.number="rules.min_ads" class="input w-full" />
        </div>
        <div class="flex gap-2">
          <button type="submit" class="btn btn-primary">Spara</button>
          <button type="button" @click="fetchRules" class="btn btn-secondary">Ladda</button>
        </div>
        <div v-if="message" class="text-sm text-emerald-400">{{ message }}</div>
      </div>
    </form>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
const api = useApi()

const rules = ref<{ min_profit_sek: number | null; min_discount: number | null; min_confidence: number | null; min_ads: number | null }>({ min_profit_sek: null, min_discount: null, min_confidence: null, min_ads: null })
const message = ref<string | null>(null)

const fetchRules = async () => {
  try {
    const data = await api.get<any>('/trading-rules')
    rules.value.min_profit_sek = data.min_profit_sek ?? null
    rules.value.min_discount = data.min_discount ?? null
    rules.value.min_confidence = data.min_confidence ?? null
    rules.value.min_ads = data.min_ads ?? null
    message.value = null
  } catch (e: any) {
    message.value = 'Kunde inte hämta inställningar'
    console.error(e)
  }
}

const save = async () => {
  try {
    await api.put('/trading-rules', {
      min_profit_sek: rules.value.min_profit_sek,
      min_discount: rules.value.min_discount,
      min_confidence: rules.value.min_confidence,
      min_ads: rules.value.min_ads,
    })
    message.value = 'Sparat'
  } catch (e: any) {
    message.value = 'Kunde inte spara inställningar'
    console.error(e)
  }
}

onMounted(() => {
  fetchRules()
})
</script>

<style scoped>
.label { display:block; margin-bottom:0.5rem; color:var(--un-prose-lead); }
.input { padding:0.5rem; border-radius:0.375rem; width:100%; }
.btn { padding:0.5rem 1rem; border-radius:0.375rem }
.btn-primary { background-color:#059669; color:white }
.btn-secondary { background-color:#334155; color:white }
</style>
