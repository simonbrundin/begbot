<template>
  <div>
    <div class="flex justify-between items-center mb-6">
      <h1 class="page-header">Lager</h1>
      <button @click="showAddModal = true" class="btn btn-primary">
        Lägg till
      </button>
    </div>

    <div class="grid grid-cols-4 gap-4 mb-6">
      <div class="stat-card">
        <p class="stat-label">Totalt antal</p>
        <p class="stat-value">{{ items.length }}</p>
      </div>
      <div class="stat-card">
        <p class="stat-label">I lager</p>
        <p class="stat-value">{{ inStockCount }}</p>
      </div>
      <div class="stat-card">
        <p class="stat-label">Utlagda</p>
        <p class="stat-value">{{ listedCount }}</p>
      </div>
      <div class="stat-card">
        <p class="stat-label">Sålda</p>
        <p class="stat-value">{{ soldCount }}</p>
      </div>
    </div>

    <div class="card p-4 mb-6">
      <div class="flex gap-4">
        <select v-model="statusFilter" class="input w-48">
          <option value="">Alla statusar</option>
          <option v-for="(name, id) in TRADE_STATUSES" :key="id" :value="id">
            {{ name }}
          </option>
        </select>
        <input
          v-model="searchQuery"
          type="text"
          class="input flex-1"
          placeholder="Sök..."
        />
      </div>
    </div>

    <div class="card overflow-hidden">
      <table class="table">
        <thead>
          <tr>
            <th>Produkt</th>
            <th>Status</th>
            <th>Köpris</th>
            <th>Försäljningspris</th>
            <th>Vinst</th>
            <th>Datum</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in filteredItems" :key="item.id">
            <td>
              <div v-if="item.product">
                <p class="font-medium text-slate-100">{{ item.product.brand }} {{ item.product.name }}</p>
                <p class="text-sm text-slate-400">{{ item.product.category }}</p>
              </div>
              <span v-else class="text-slate-500">Okänd</span>
            </td>
            <td>
              <span :class="statusBadgeClass(item.status_id)">
                {{ TRADE_STATUSES[item.status_id] || 'okänd' }}
              </span>
            </td>
            <td>{{ formatCurrency(item.buy_price) }}</td>
            <td>{{ item.sell_price ? formatCurrency(item.sell_price) : '-' }}</td>
            <td :class="profitClass(item)">
              {{ calculateProfit(item) }}
            </td>
            <td class="text-sm text-slate-400">
              {{ formatDate(item.created_at) }}
            </td>
            <td>
              <button @click="editItem(item)" class="text-primary-400 hover:text-primary-300">
                Redigera
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="showAddModal || editingItem" class="fixed inset-0 bg-black/70 flex items-center justify-center z-50">
      <div class="bg-slate-800 rounded-lg p-6 w-full max-w-2xl max-h-[90vh] overflow-y-auto border border-slate-700">
        <h2 class="text-xl font-bold text-slate-100 mb-4">
          {{ editingItem ? 'Redigera' : 'Lägg till ny' }}
        </h2>

        <form @submit.prevent="saveItem" class="space-y-4">
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="label">Produkt</label>
              <select v-model="itemForm.product_id" class="input">
                <option value="">Välj produkt...</option>
                <option v-for="p in products" :key="p.id" :value="p.id">
                  {{ p.brand }} {{ p.name }}
                </option>
              </select>
            </div>
            <div>
              <label class="label">Status</label>
              <select v-model="itemForm.status_id" class="input">
                <option v-for="(name, id) in TRADE_STATUSES" :key="id" :value="id">
                  {{ name }}
                </option>
              </select>
            </div>
            <div>
              <label class="label">Köpris (öre)</label>
              <input v-model.number="itemForm.buy_price" type="number" class="input" />
            </div>
            <div>
              <label class="label">Köpfrit (öre)</label>
              <input v-model.number="itemForm.buy_shipping_cost" type="number" class="input" />
            </div>
            <div>
              <label class="label">Försäljningspris (öre)</label>
              <input v-model.number="itemForm.sell_price" type="number" class="input" />
            </div>
            <div>
              <label class="label">Förpackning (öre)</label>
              <input v-model.number="itemForm.sell_packaging_cost" type="number" class="input" />
            </div>
            <div>
              <label class="label">Frakt (öre)</label>
              <input v-model.number="itemForm.sell_postage_cost" type="number" class="input" />
            </div>
            <div>
              <label class="label">Frakt mottaget (öre)</label>
              <input v-model.number="itemForm.sell_shipping_collected" type="number" class="input" />
            </div>
            <div>
              <label class="label">Lagerplats</label>
              <input v-model.number="itemForm.storage" type="number" class="input" />
            </div>
            <div>
              <label class="label">Länk</label>
              <input v-model="itemForm.source_link" type="text" class="input" />
            </div>
            <div>
              <label class="label">Köpt datum</label>
              <input v-model="itemForm.buy_date" type="date" class="input" />
            </div>
            <div>
              <label class="label">Sålt datum</label>
              <input v-model="itemForm.sell_date" type="date" class="input" />
            </div>
          </div>

          <div class="flex justify-end gap-2 pt-4">
            <button type="button" @click="closeModal" class="btn btn-secondary">
              Avbryt
            </button>
            <button type="submit" class="btn btn-primary">
              {{ editingItem ? 'Spara' : 'Lägg till' }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { TradedItem, Product } from '~/types/database'
import { TRADE_STATUSES } from '~/types/database'

const api = useApi()

const items = ref<TradedItem[]>([])
const products = ref<Product[]>([])
const loading = ref(false)
const showAddModal = ref(false)
const editingItem = ref<TradedItem | null>(null)
const statusFilter = ref('')
const searchQuery = ref('')

const defaultForm = {
  product_id: null as number | null,
  status_id: 1,
  buy_price: 0,
  buy_shipping_cost: 0,
  sell_price: null as number | null,
  sell_packaging_cost: 0,
  sell_postage_cost: 0,
  sell_shipping_collected: 0,
  storage: null as number | null,
  source_link: '',
  buy_date: '',
  sell_date: ''
}

const itemForm = ref({ ...defaultForm })

const inStockCount = computed(() => items.value.filter(i => i.status_id === 3).length)
const listedCount = computed(() => items.value.filter(i => i.status_id === 4).length)
const soldCount = computed(() => items.value.filter(i => i.status_id === 5).length)

const filteredItems = computed(() => {
  return items.value.filter(item => {
    if (statusFilter.value && item.status_id !== parseInt(statusFilter.value)) return false
    if (searchQuery.value) {
      const query = searchQuery.value.toLowerCase()
      const productName = item.product?.name?.toLowerCase() || ''
      const brand = item.product?.brand?.toLowerCase() || ''
      if (!productName.includes(query) && !brand.includes(query)) return false
    }
    return true
  })
})

const fetchData = async () => {
  loading.value = true
  try {
    const [itemsRes, productsRes] = await Promise.all([
      api.get<TradedItem[]>('/inventory'),
      api.get<Product[]>('/products')
    ])
    items.value = itemsRes
    products.value = productsRes
  } catch (e) {
    console.error('Failed to fetch data:', e)
  } finally {
    loading.value = false
  }
}

const editItem = (item: TradedItem) => {
  editingItem.value = item
  itemForm.value = {
    product_id: item.product_id,
    status_id: item.status_id,
    buy_price: item.buy_price,
    buy_shipping_cost: item.buy_shipping_cost,
    sell_price: item.sell_price || null,
    sell_packaging_cost: item.sell_packaging_cost || 0,
    sell_postage_cost: item.sell_postage_cost || 0,
    sell_shipping_collected: item.sell_shipping_collected || 0,
    storage: item.storage,
    source_link: item.source_link,
    buy_date: item.buy_date ? item.buy_date.split('T')[0] : '',
    sell_date: item.sell_date ? item.sell_date.split('T')[0] : ''
  }
}

const closeModal = () => {
  showAddModal.value = false
  editingItem.value = null
  itemForm.value = { ...defaultForm }
}

const saveItem = async () => {
  try {
    if (editingItem.value) {
      await api.put(`/inventory/${editingItem.value.id}`, itemForm.value)
    } else {
      await api.post('/inventory', itemForm.value)
    }
    closeModal()
    await fetchData()
  } catch (e) {
    console.error('Failed to save item:', e)
  }
}

const formatCurrency = (cents: number) => {
  return `${(cents / 100).toFixed(2)} kr`
}

const formatDate = (dateStr: string) => {
  return new Date(dateStr).toLocaleDateString('sv-SE')
}

const calculateProfit = (item: TradedItem) => {
  const sellTotal = (item.sell_price || 0) + (item.sell_shipping_collected || 0)
  const buyTotal = item.buy_price + item.buy_shipping_cost
  const sellCost = (item.sell_packaging_cost || 0) + (item.sell_postage_cost || 0)
  const profit = sellTotal - (buyTotal + sellCost)
  return profit >= 0 ? `+${formatCurrency(profit)}` : formatCurrency(profit)
}

const profitClass = (item: TradedItem) => {
  const sellTotal = (item.sell_price || 0) + (item.sell_shipping_collected || 0)
  const buyTotal = item.buy_price + item.buy_shipping_cost
  const sellCost = (item.sell_packaging_cost || 0) + (item.sell_postage_cost || 0)
  const profit = sellTotal - (buyTotal + sellCost)
  return profit >= 0 ? 'text-emerald-400' : 'text-red-400'
}

const statusBadgeClass = (statusId: number) => {
  const classes: Record<number, string> = {
    1: 'badge badge-info',
    2: 'badge badge-info',
    3: 'badge badge-warning',
    4: 'badge badge-info',
    5: 'badge badge-success'
  }
  return classes[statusId] || classes[1]
}

onMounted(fetchData)
</script>
