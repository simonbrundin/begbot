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

    <div class="card overflow-hidden">
      <table class="table">
        <thead>
          <tr>
            <th>Märke</th>
            <th>Namn</th>
            <th>Kategori</th>
            <th>Variant</th>
            <th v-for="vt in enabledValuationTypes" :key="vt.id">{{ vt.name }}</th>
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
            <td>{{ product.model_variant || '-' }}</td>
            <template v-for="vt in enabledValuationTypes" :key="vt.id">
              <td class="text-sm">
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
import type { Product, Valuation, ValuationType } from '~/types/database'

const api = useApi()

const products = ref<Product[]>([])
const valuationsByProduct = ref<Record<number, Valuation[]>>({})
const valuationTypes = ref<ValuationType[]>([])

const enabledValuationTypes = computed(() => valuationTypes.value.filter(t => t.enabled !== false))
const loading = ref(false)
const showAddModal = ref(false)
const editingProduct = ref<Product | null>(null)

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

    // Fetch valuations per product (server requires product_id)
    if (products.value.length > 0) {
      const perProductPromises = products.value.map(p => api.get<Valuation[]>(`/valuations?product_id=${p.id}`))
      const perProductResults = await Promise.allSettled(perProductPromises)
      perProductResults.forEach((res, idx) => {
        const pid = products.value[idx].id
        if (res.status === 'fulfilled' && Array.isArray(res.value)) {
          grouped[pid] = res.value
        }
      })
    }
    valuationsByProduct.value = grouped

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
}

const closeModal = () => {
  showAddModal.value = false
  editingProduct.value = null
  form.value = { ...defaultForm }
}

const saveProduct = async () => {
  try {
    if (editingProduct.value) {
      await api.put(`/products/${editingProduct.value.id}`, form.value)
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
