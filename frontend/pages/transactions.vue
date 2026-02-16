<template>
  <div>
    <div class="flex justify-between items-center mb-6">
      <h1 class="page-header">Transaktioner</h1>
      <button @click="showAddModal = true" class="btn btn-primary">
        Lägg till transaktion
      </button>
    </div>

    <div class="grid grid-cols-3 gap-4 mb-6">
      <div class="stat-card">
        <p class="stat-label">Total inkomst</p>
        <p class="stat-value text-emerald-400">{{ formatCurrency(totalIncome) }}</p>
      </div>
      <div class="stat-card">
        <p class="stat-label">Total utgift</p>
        <p class="stat-value text-red-400">{{ formatCurrency(totalExpenses) }}</p>
      </div>
      <div class="stat-card">
        <p class="stat-label">Netto</p>
        <p :class="netAmount >= 0 ? 'stat-value text-emerald-400' : 'stat-value text-red-400'">
          {{ formatCurrency(netAmount) }}
        </p>
      </div>
    </div>

    <div class="card overflow-hidden">
      <table class="table">
        <thead>
          <tr>
            <th>Datum</th>
            <th>Typ</th>
            <th>Belopp</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="tx in transactions" :key="tx.id">
            <td>{{ formatDate(tx.date) }}</td>
            <td>
              <span class="badge" :class="typeClass(tx.transaction_type)">
                {{ transactionTypeName(tx.transaction_type) }}
              </span>
            </td>
            <td :class="tx.amount >= 0 ? 'text-emerald-400' : 'text-red-400'">
              {{ formatCurrency(tx.amount) }}
            </td>
            <td>
              <button @click="deleteTransaction(tx.id)" class="text-red-400 hover:text-red-300 text-sm">
                Ta bort
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="showAddModal" class="fixed inset-0 bg-black/70 flex items-center justify-center z-50">
      <div class="bg-slate-800 rounded-lg p-6 w-full max-w-md border border-slate-700">
        <h2 class="text-xl font-bold text-slate-100 mb-4">Lägg till transaktion</h2>

        <form @submit.prevent="saveTransaction" class="space-y-4">
          <div>
            <label class="label">Datum</label>
            <input v-model="form.date" type="date" class="input" required />
          </div>
          <div>
            <label class="label">Typ</label>
            <select v-model="form.transaction_type" class="input" required>
              <option value="">Välj typ...</option>
              <option v-for="t in transactionTypes" :key="t.id" :value="t.id">
                {{ t.name }}
              </option>
            </select>
          </div>
          <div>
            <label class="label">Belopp (öre, negativt för utgift)</label>
            <input v-model.number="form.amount" type="number" class="input" required />
          </div>

          <div class="flex justify-end gap-2 pt-4">
            <button type="button" @click="showAddModal = false" class="btn btn-secondary">
              Avbryt
            </button>
            <button type="submit" class="btn btn-primary">Lägg till</button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Transaction, TransactionType } from '~/types/database'

const api = useApi()

const transactions = ref<Transaction[]>([])
const transactionTypes = ref<TransactionType[]>([])
const loading = ref(false)
const showAddModal = ref(false)

const form = ref({
  date: new Date().toISOString().split('T')[0],
  transaction_type: null as number | null,
  amount: 0
})

const totalIncome = computed(() =>
  transactions.value.filter(t => t.amount > 0).reduce((sum, t) => sum + t.amount, 0)
)

const totalExpenses = computed(() =>
  Math.abs(transactions.value.filter(t => t.amount < 0).reduce((sum, t) => sum + t.amount, 0))
)

const netAmount = computed(() =>
  transactions.value.reduce((sum, t) => sum + t.amount, 0)
)

const fetchData = async () => {
  loading.value = true
  try {
    const [txRes, typesRes] = await Promise.all([
      api.get<Transaction[]>('/transactions'),
      api.get<TransactionType[]>('/transaction-types')
    ])
    transactions.value = txRes
    transactionTypes.value = typesRes
  } catch (e) {
    console.error('Failed to fetch data:', e)
  } finally {
    loading.value = false
  }
}

const formatCurrency = (cents: number) => `${(cents / 100).toFixed(2)} kr`
const formatDate = (dateStr: string) => new Date(dateStr).toLocaleDateString('sv-SE')

const typeClass = (typeId: number | null) => {
  if (!typeId) return 'bg-slate-700 text-slate-300'
  const type = transactionTypes.value.find(t => t.id === typeId)
  if (type?.name?.toLowerCase().includes('income') || type?.name?.toLowerCase().includes('sell')) {
    return 'badge badge-success'
  }
  return 'badge badge-danger'
}

const transactionTypeName = (typeId: number | null) => {
  if (!typeId) return 'Unknown'
  return transactionTypes.value.find(t => t.id === typeId)?.name || 'Unknown'
}

const saveTransaction = async () => {
  try {
    await api.post('/transactions', form.value)
    showAddModal.value = false
    form.value = {
      date: new Date().toISOString().split('T')[0],
      transaction_type: null,
      amount: 0
    }
    await fetchData()
  } catch (e) {
    console.error('Failed to save transaction:', e)
  }
}

const deleteTransaction = async (id: number) => {
  if (!confirm('Ta bort denna transaktion?')) return
  try {
    await api.delete(`/transactions/${id}`)
    await fetchData()
  } catch (e) {
    console.error('Failed to delete transaction:', e)
  }
}

onMounted(fetchData)
</script>
