<template>
  <div>
    <h1 class="page-header">Analyser</h1>

    <div class="grid grid-cols-4 gap-4 mb-6">
      <div class="stat-card">
        <p class="stat-label">Total vinst</p>
        <p class="stat-value text-emerald-400">{{ formatCurrency(totalProfit) }}</p>
      </div>
      <div class="stat-card">
        <p class="stat-label">Lager värde</p>
        <p class="stat-value">{{ formatCurrency(inventoryValue) }}</p>
      </div>
      <div class="stat-card">
        <p class="stat-label">Sålda artiklar</p>
        <p class="stat-value">{{ soldItems.length }}</p>
      </div>
      <div class="stat-card">
        <p class="stat-label">Artiklar i lager</p>
        <p class="stat-value">{{ inStockItems.length }}</p>
      </div>
    </div>

    <div class="grid grid-cols-2 gap-6">
      <div class="card p-6">
        <h2 class="section-title">Artiklar per status</h2>
        <div class="space-y-3">
          <div v-for="(count, statusId) in statusCounts" :key="statusId" class="flex items-center">
            <span class="w-24 text-sm text-slate-400">{{ TRADE_STATUSES[parseInt(statusId)] }}</span>
            <div class="flex-1 bg-slate-700 rounded-full h-4 mx-2">
              <div
                class="h-full rounded-full transition-all"
                :class="statusColor(statusId)"
                :style="{ width: `${(count / items.length) * 100}%` }"
              ></div>
            </div>
            <span class="w-8 text-sm font-medium">{{ count }}</span>
          </div>
        </div>
      </div>

      <div class="card p-6">
        <h2 class="section-title">Vinstfördelning (sålda artiklar)</h2>
        <div class="space-y-3">
          <div class="flex justify-between">
            <span class="text-slate-400">Total intäkt</span>
            <span class="font-medium text-slate-100">{{ formatCurrency(totalRevenue) }}</span>
          </div>
          <div class="flex justify-between">
            <span class="text-slate-400">Kostnad för sålda varor</span>
            <span class="font-medium text-red-400">-{{ formatCurrency(totalCOGS) }}</span>
          </div>
          <div class="flex justify-between">
            <span class="text-slate-400">Fraktkostnader</span>
            <span class="font-medium text-red-400">-{{ formatCurrency(totalShipping) }}</span>
          </div>
          <div class="border-t border-slate-700 pt-2 flex justify-between">
            <span class="font-bold text-slate-100">Nettoresultat</span>
            <span class="font-bold text-emerald-400">{{ formatCurrency(totalProfit) }}</span>
          </div>
        </div>
      </div>
    </div>

    <div class="card p-6 mt-6">
      <h2 class="section-title">Senaste försäljningar</h2>
      <table class="table">
        <thead>
          <tr>
            <th>Produkt</th>
            <th>Köpris</th>
            <th>Försäljningspris</th>
            <th>Vinst</th>
            <th>Såld datum</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in soldItems.slice(0, 10)" :key="item.id">
            <td>
              <span v-if="item.product" class="text-slate-100">{{ item.product.brand }} {{ item.product.name }}</span>
              <span v-else class="text-slate-500">Okänd</span>
            </td>
            <td>{{ formatCurrency(item.buy_price + item.buy_shipping_cost) }}</td>
            <td>
              {{ item.sell_price ? formatCurrency(item.sell_price + (item.sell_shipping_collected || 0)) : '-' }}
            </td>
            <td class="text-emerald-400 font-medium">{{ calculateProfit(item) }}</td>
            <td class="text-sm text-slate-400">
              {{ item.sell_date ? formatDate(item.sell_date) : '-' }}
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { TradedItem } from '~/types/database'
import { TRADE_STATUSES } from '~/types/database'

const api = useApi()

const items = ref<TradedItem[]>([])
const loading = ref(false)

const soldItems = computed(() => items.value.filter(i => i.status_id === 5))
const inStockItems = computed(() => items.value.filter(i => i.status_id === 3))

const totalProfit = computed(() =>
  soldItems.value.reduce((sum, item) => {
    const sellTotal = (item.sell_price || 0) + (item.sell_shipping_collected || 0)
    const buyTotal = item.buy_price + item.buy_shipping_cost
    const sellCost = (item.sell_packaging_cost || 0) + (item.sell_postage_cost || 0)
    return sum + (sellTotal - (buyTotal + sellCost))
  }, 0)
)

const totalRevenue = computed(() =>
  soldItems.value.reduce((sum, item) => sum + (item.sell_price || 0) + (item.sell_shipping_collected || 0), 0)
)

const totalCOGS = computed(() =>
  soldItems.value.reduce((sum, item) => sum + item.buy_price + item.buy_shipping_cost, 0)
)

const totalShipping = computed(() =>
  soldItems.value.reduce((sum, item) => sum + (item.sell_packaging_cost || 0) + (item.sell_postage_cost || 0), 0)
)

const inventoryValue = computed(() =>
  inStockItems.value.reduce((sum, item) => sum + item.buy_price + item.buy_shipping_cost, 0)
)

const statusCounts = computed(() => {
  const counts: Record<number, number> = {}
  items.value.forEach(item => {
    counts[item.status_id] = (counts[item.status_id] || 0) + 1
  })
  return counts
})

const fetchData = async () => {
  loading.value = true
  try {
    items.value = await api.get<TradedItem[]>('/inventory')
  } catch (e) {
    console.error('Failed to fetch data:', e)
  } finally {
    loading.value = false
  }
}

const formatCurrency = (cents: number) => `${(cents / 100).toFixed(2)} kr`
const formatDate = (dateStr: string) => new Date(dateStr).toLocaleDateString('sv-SE')

const calculateProfit = (item: TradedItem) => {
  const sellTotal = (item.sell_price || 0) + (item.sell_shipping_collected || 0)
  const buyTotal = item.buy_price + item.buy_shipping_cost
  const sellCost = (item.sell_packaging_cost || 0) + (item.sell_postage_cost || 0)
  const profit = sellTotal - (buyTotal + sellCost)
  return `+${formatCurrency(profit)}`
}

const statusColor = (statusId: string | number) => {
  const id = typeof statusId === 'string' ? parseInt(statusId) : statusId
  const colors: Record<number, string> = {
    1: 'bg-slate-500',
    2: 'bg-sky-500',
    3: 'bg-amber-500',
    4: 'bg-purple-500',
    5: 'bg-emerald-500'
  }
  return colors[id] || 'bg-slate-500'
}

onMounted(fetchData)
</script>
