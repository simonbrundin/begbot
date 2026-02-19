<template>
  <div>
    <div class="flex justify-between items-center mb-6">
      <h1 class="page-header">Produkter</h1>
      <button @click="showAddModal = true" class="btn btn-primary">
        Lägg till produkt
      </button>
    </div>

    <!-- Save status toast -->
    <div v-if="saveStatus.show" :class="`fixed right-4 bottom-4 z-50 p-3 rounded shadow ${saveStatus.type==='error'? 'bg-red-600 text-white' : 'bg-emerald-600 text-white'}`">
      {{ saveStatus.message }}
    </div>

    <div class="mb-4">
      <button @click="showWeightConfig = !showWeightConfig" class="text-sm text-slate-400 hover:text-slate-200 flex items-center gap-1">
        <span>Vikter för sammanvägd värdering</span>
        <span>{{ showWeightConfig ? '▲' : '▼' }}</span>
      </button>
      <div v-if="showWeightConfig" class="mt-2 card p-4 flex flex-wrap gap-4">
        <div v-for="vt in enabledValuationTypes" :key="vt.id" class="flex items-center gap-2">
          <label class="text-sm text-slate-300">{{ vt.name }}</label>
          <input v-model.number="weights[vt.id]" type="number" min="0" step="0.1" class="input w-20 py-1 text-sm" />
        </div>
        <div v-if="enabledValuationTypes.length === 0" class="text-sm text-slate-400">Inga aktiverade värderingstyper.</div>
      </div>
    </div>

    <div class="card overflow-hidden">
      <table class="table">
        <thead>
          <tr>
            <th>Märke</th>
            <th>Namn</th>
            <th>Kategori</th>
            <th v-for="vt in enabledValuationTypes" :key="vt.id">{{ vt.name }}</th>
            <th>Sammanvägd värdering</th>
            <th>Aktiverad</th>
            <th>Skapad</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="product in products" :key="product.id">
            <td class="font-medium text-slate-100">{{ product.brand || '-' }}</td>
            <td>{{ product.name || '-' }}</td>
            <td>{{ product.category || '-' }}</td>
            <template v-for="vt in enabledValuationTypes" :key="vt.id">
              <td class="text-sm" :class="{ 'opacity-40': !isTypeActiveForProduct(product.id, vt.id) }">
                <div v-if="valuationsByProduct[product.id]">
                  <template v-if="isEditingValuation(product.id, vt.id)">
                    <div class="flex items-center gap-2">
                      <input
                        v-model.number="editingValuationInput"
                        @keyup.enter="saveValuation(product.id, vt.id)"
                        type="number"
                        class="input input-sm w-28"
                        />
                      <button @click="saveValuation(product.id, vt.id)" class="btn btn-primary btn-sm">Spara</button>
                      <button @click="cancelEditValuation" class="btn btn-secondary btn-sm">Avbryt</button>
                    </div>
                  </template>
                  <template v-else>
                    <span
                      v-if="getValuationForType(product.id, vt.id)"
                      class="text-xs bg-slate-700 px-2 py-1 rounded cursor-pointer"
                      @click="startEditValuation(product.id, vt.id)"
                    >
                      {{ formatValuationAsSEK(getValuationForType(product.id, vt.id)!.valuation) }}
                    </span>
                    <button v-else @click="startEditValuation(product.id, vt.id)" class="text-xs text-slate-400 hover:text-primary-300">+</button>
                  </template>
                </div>
                <div v-else class="text-xs text-slate-400">-</div>
              </td>
            </template>
            <td>
              <template v-if="weightedValuations[product.id]">
                <span class="badge badge-info">
                  {{ formatValuationAsSEK(weightedValuations[product.id]!.average) }}
                </span>
                <span class="text-xs text-slate-400 ml-1" :title="`Säkerhetsprocent baserat på spridning mellan värderingstyper`">{{ weightedValuations[product.id]!.safetyPercent }}%</span>
              </template>
            </td>
            <td>
              <button
                @click="toggleEnabled(product)"
                :class="product.enabled === true ? 'badge badge-success' : 'badge'"
              >
                {{ product.enabled === true ? 'Ja' : 'Nej' }}
              </button>
            </td>
            <td class="text-sm text-slate-400">{{ formatDate(product.created_at) }}</td>
            <td>
              <button @click="editProduct(product)" class="text-primary-400 hover:text-primary-300">
                Redigera
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="showAddModal || editingProduct" class="fixed inset-0 bg-black/70 flex items-center justify-center z-50">
      <div class="bg-slate-800 rounded-lg p-6 w-full max-w-lg border border-slate-700">
        <h2 class="text-xl font-bold text-slate-100 mb-4">
          {{ editingProduct ? 'Redigera produkt' : 'Lägg till ny produkt' }}
        </h2>

        <form @submit.prevent="saveProduct" class="space-y-4">
          <div>
            <label class="label">Märke</label>
            <input v-model="form.brand" type="text" class="input" />
          </div>
          <div>
            <label class="label">Namn</label>
            <input v-model="form.name" type="text" class="input" />
          </div>
          <div>
            <label class="label">Kategori</label>
            <input v-model="form.category" type="text" class="input" placeholder="t.ex., telefon" />
          </div>
          <div>
            <label class="label">Modellvariant</label>
            <input v-model="form.model_variant" type="text" class="input" placeholder="t.ex., 256GB" />
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="label">Förpackning (öre)</label>
              <input v-model.number="form.sell_packaging_cost" type="number" class="input" />
            </div>
            <div>
              <label class="label">Frakt (öre)</label>
              <input v-model.number="form.sell_postage_cost" type="number" class="input" />
            </div>
          </div>

          <div v-if="editingProduct && enabledValuationTypes.length > 0">
            <label class="label">Aktiva värderingstyper</label>
            <div class="flex flex-wrap gap-3">
              <label
                v-for="vt in enabledValuationTypes"
                :key="vt.id"
                class="flex items-center gap-2 cursor-pointer text-sm text-slate-300"
              >
                <input
                  type="checkbox"
                  :checked="editingValuationTypeActive[vt.id] ?? true"
                  @change="editingValuationTypeActive[vt.id] = ($event.target as HTMLInputElement).checked"
                  class="accent-primary-500"
                />
                {{ vt.name }}
              </label>
            </div>
            <p class="text-xs text-slate-500 mt-1">Inaktiva typer exkluderas från sammanvägd värdering. Minst en måste vara aktiv.</p>
          </div>

          <div class="flex justify-end gap-2 pt-4">
            <button type="button" @click="closeModal" class="btn btn-secondary">
              Avbryt
            </button>
            <button type="submit" class="btn btn-primary">
              {{ editingProduct ? 'Spara' : 'Lägg till' }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Product, Valuation, ValuationType, ProductValuationTypeConfig } from '~/types/database'

const api = useApi()

const products = ref<Product[]>([])
const valuationsByProduct = ref<Record<number, Valuation[]>>({})
const valuationTypes = ref<ValuationType[]>([])
const valuationConfigsByProduct = ref<Record<number, ProductValuationTypeConfig[]>>({})

const enabledValuationTypes = computed(() => valuationTypes.value.filter(t => t.enabled !== false))
const loading = ref(false)
const showAddModal = ref(false)
const editingProduct = ref<Product | null>(null)

// Per-product valuation type active states in edit form (typeId -> isActive)
const editingValuationTypeActive = ref<Record<number, boolean>>({})

// Weights for sammanvägd värdering (keyed by valuation type id, default 1)
const weights = ref<Record<number, number>>({})
const showWeightConfig = ref(false)

watch(valuationTypes, (types) => {
  types.forEach(vt => {
    if (vt.enabled !== false && weights.value[vt.id] === undefined) {
      weights.value[vt.id] = 1
    }
  })
}, { immediate: true })

// Check if a valuation type is active for a product (defaults to true when no config)
const isTypeActiveForProduct = (productId: number, typeId: number): boolean => {
  const configs = valuationConfigsByProduct.value[productId]
  if (!configs || configs.length === 0) return true
  const config = configs.find(c => c.valuation_type_id === typeId)
  if (!config) return true
  return config.is_active
}

const computeWeightedValuation = (productId: number): { average: number; safetyPercent: number } | null => {
  const activeTypes = enabledValuationTypes.value.filter(vt => isTypeActiveForProduct(productId, vt.id))
  if (activeTypes.length === 0) return null
  const entries = activeTypes
    .map(vt => {
      const v = getValuationForType(productId, vt.id)
      return v !== null ? { valuation: v.valuation, weight: weights.value[vt.id] ?? 1 } : null
    })
    .filter((e): e is { valuation: number; weight: number } => e !== null)
  if (entries.length === 0) return null
  const totalWeight = entries.reduce((s, e) => s + e.weight, 0)
  if (totalWeight === 0) return null
  const average = entries.reduce((s, e) => s + e.valuation * e.weight, 0) / totalWeight
  let safetyPercent = 100
  if (entries.length > 1) {
    const mean = entries.reduce((s, e) => s + e.valuation, 0) / entries.length
    const variance = entries.reduce((s, e) => s + Math.pow(e.valuation - mean, 2), 0) / entries.length
    const stdDev = Math.sqrt(variance)
    safetyPercent = mean !== 0 ? Math.max(0, Math.round(100 - (stdDev / Math.abs(mean) * 100))) : 0
  }
  return { average: Math.round(average), safetyPercent }
}

const weightedValuations = computed(() => {
  const result: Record<number, { average: number; safetyPercent: number } | null> = {}
  for (const product of products.value) {
    result[product.id] = computeWeightedValuation(product.id)
  }
  return result
})

const defaultForm = {
  brand: '',
  name: '',
  category: '',
  model_variant: '',
  sell_packaging_cost: 0,
  sell_postage_cost: 0,
  enabled: false
}

const form = ref({ ...defaultForm })

const fetchData = async () => {
  loading.value = true
  try {
    // Fetch products, valuations and valuation types. Use allSettled
    // so the products list still appears even if auxiliary endpoints fail.
    const [prodsRes, typesRes] = await Promise.allSettled([
      api.get<Product[]>('/products'),
      api.get<ValuationType[]>('/valuation-types')
    ])

    if (prodsRes.status === 'fulfilled') {
      products.value = prodsRes.value
    } else {
      console.error('Failed to fetch products:', prodsRes.reason)
      products.value = []
    }

    const grouped: Record<number, Valuation[]> = {}
    const configs: Record<number, ProductValuationTypeConfig[]> = {}

    // Fetch valuations and valuation type configs per product (server requires product_id)
    if (products.value.length > 0) {
      const perProductPromises = products.value.map(p => api.get<Valuation[]>(`/valuations?product_id=${p.id}`))
      const configPromises = products.value.map(p => api.get<ProductValuationTypeConfig[]>(`/products/${p.id}/valuation-type-config`))
      const [perProductResults, configResults] = await Promise.all([
        Promise.allSettled(perProductPromises),
        Promise.allSettled(configPromises)
      ])
      perProductResults.forEach((res, idx) => {
        const pid = products.value[idx].id
        if (res.status === 'fulfilled' && Array.isArray(res.value)) {
          grouped[pid] = res.value
        }
      })
      configResults.forEach((res, idx) => {
        const pid = products.value[idx].id
        if (res.status === 'fulfilled' && Array.isArray(res.value)) {
          configs[pid] = res.value
        }
      })
    }
    valuationsByProduct.value = grouped
    valuationConfigsByProduct.value = configs

    if (typesRes.status === 'fulfilled' && Array.isArray(typesRes.value)) {
      valuationTypes.value = typesRes.value
    } else if (typesRes.status === 'rejected') {
      console.warn('Failed to fetch valuation types:', typesRes.reason)
      valuationTypes.value = []
    }
  } finally {
    loading.value = false
  }
}

const formatDate = (dateStr?: string | null) => {
  if (!dateStr || dateStr === 'null') return '-'
  const date = new Date(dateStr)
  if (isNaN(date.getTime())) return '-'
  return date.toLocaleDateString('sv-SE')
}

const formatValuationAsSEK = (sek: number | null | undefined) => {
  if (sek === null || sek === undefined) return '-'
  return sek.toLocaleString('sv-SE')
}

const formatEnabled = (enabled: boolean) => enabled ? 'Ja' : 'Nej'

const editProduct = (product: Product) => {
  editingProduct.value = product
  form.value = {
    brand: product.brand || '',
    name: product.name || '',
    category: product.category || '',
    model_variant: product.model_variant || '',
    sell_packaging_cost: product.sell_packaging_cost,
    sell_postage_cost: product.sell_postage_cost,
    enabled: product.enabled ?? false
  }
  // Initialize per-product valuation type active state
  const activeMap: Record<number, boolean> = {}
  enabledValuationTypes.value.forEach(vt => {
    activeMap[vt.id] = isTypeActiveForProduct(product.id, vt.id)
  })
  editingValuationTypeActive.value = activeMap
}

const closeModal = () => {
  showAddModal.value = false
  editingProduct.value = null
  form.value = { ...defaultForm }
  editingValuationTypeActive.value = {}
}

const saveProduct = async () => {
  try {
    if (editingProduct.value) {
      await api.put(`/products/${editingProduct.value.id}`, form.value)
      // Save valuation type configs
      if (enabledValuationTypes.value.length > 0) {
        const activeCount = Object.values(editingValuationTypeActive.value).filter(Boolean).length
        if (activeCount === 0) {
          showSaveStatus('error', 'Minst en värderingstyp måste vara aktiv')
          return
        }
        const configs: ProductValuationTypeConfig[] = enabledValuationTypes.value.map(vt => ({
          product_id: editingProduct.value!.id,
          valuation_type_id: vt.id,
          is_active: editingValuationTypeActive.value[vt.id] ?? true
        }))
        await api.put(`/products/${editingProduct.value.id}/valuation-type-config`, { configs } as any)
      }
    } else {
      await api.post('/products', form.value)
    }
    closeModal()
    await fetchData()
  } catch (e) {
    console.error('Failed to save product:', e)
  }
}

const toggleEnabled = async (product: Product) => {
  try {
    await api.put(`/products/${product.id}`, { ...product, enabled: !product.enabled })
    await fetchData()
  } catch (e) {
    console.error('Failed to toggle enabled:', e)
  }
}

const getValuationForType = (productId: number, typeId: number) => {
  const vals = valuationsByProduct.value[productId]
  if (!vals) return null
  return vals.find(v => v.valuation_type_id === typeId) || null
}

// Inline edit state for valuations
const editingValuation = ref<{ productId: number; typeId: number; value: number | null; id?: number } | null>(null)
const editingValuationInput = ref<number | null>(null)

// Simple save status toast
const saveStatus = ref<{ show: boolean; type: 'success' | 'error' | null; message: string }>({ show: false, type: null, message: '' })
const showSaveStatus = (type: 'success' | 'error', message: string, ms = 2500) => {
  saveStatus.value = { show: true, type, message }
  setTimeout(() => { saveStatus.value.show = false }, ms)
}

const isEditingValuation = (productId: number, typeId: number) => {
  return !!(editingValuation.value && editingValuation.value.productId === productId && editingValuation.value.typeId === typeId)
}

const startEditValuation = (productId: number, typeId: number) => {
  const v = getValuationForType(productId, typeId)
  editingValuation.value = { productId, typeId, value: v?.valuation ?? null, id: v?.id }
  editingValuationInput.value = v?.valuation ?? null
}

const cancelEditValuation = () => {
  editingValuation.value = null
  editingValuationInput.value = null
}

const saveValuation = async (productId: number, typeId: number) => {
  if (!editingValuation.value) return
  const val = editingValuationInput.value
  try {
    console.debug('saveValuation called', { productId, typeId, val, id: editingValuation.value?.id })
    if (editingValuation.value.id) {
      const res = await api.put(`/valuations/${editingValuation.value.id}`, { valuation: val })
      // log response for debugging
      console.debug('PUT /valuations response:', res)
      // update local state optimistically so UI reflects change immediately
      const pid = productId
      const arr = valuationsByProduct.value[pid]
      if (arr) {
        const idx = arr.findIndex(v => v.id === editingValuation.value!.id)
        if (idx !== -1) {
          arr[idx].valuation = val as number
        }
      }
    } else {
      const res = await api.post('/valuations', { product_id: productId, valuation_type_id: typeId, valuation: val })
      console.debug('POST /valuations response:', res)
      // if created, add to local state
      const created: any = res
      if (created && created.id) {
        const pid = productId
        if (!valuationsByProduct.value[pid]) valuationsByProduct.value[pid] = []
        valuationsByProduct.value[pid].push({ id: created.id, product_id: pid, valuation_type_id: typeId, valuation: val, created_at: new Date().toISOString() } as any)
      }
    }
    // refresh in background, but UI already updated optimistically
    fetchData()
    editingValuation.value = null
    showSaveStatus('success', 'Värdering sparad')
  } catch (e) {
    console.error('Failed to save valuation:', e)
    showSaveStatus('error', 'Kunde inte spara värdering')
  }
}

onMounted(fetchData)
</script>
