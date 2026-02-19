<template>
  <div>
    <div class="flex justify-between items-center mb-6">
      <h1 class="page-header">Meddelanden</h1>
      <div class="flex gap-2">
        <button 
          @click="showNeedsReview = !showNeedsReview" 
          :class="showNeedsReview ? 'btn btn-primary' : 'btn btn-secondary'"
        >
          {{ showNeedsReview ? 'Behöver granskning' : 'Alla konversationer' }}
        </button>
      </div>
    </div>

    <div v-if="loading" class="text-center py-12 text-slate-500">
      Laddar konversationer...
    </div>

    <div v-else-if="conversations.length === 0" class="text-center py-12 text-slate-500">
      <p v-if="showNeedsReview">Inga konversationer behöver granskning just nu.</p>
      <p v-else>Inga konversationer hittades.</p>
    </div>

    <div v-else class="space-y-4">
      <div 
        v-for="conv in conversations" 
        :key="conv.id" 
        class="card p-4 hover:border-primary-500 cursor-pointer transition-colors"
        @click="navigateTo(`/conversations/${conv.id}`)"
      >
        <div class="flex justify-between items-start mb-2">
          <div class="flex-1">
            <h3 class="text-lg font-semibold text-slate-100">{{ conv.listing_title }}</h3>
            <p class="text-sm text-slate-400">{{ conv.marketplace_name }}</p>
          </div>
          <div class="text-right">
            <p v-if="conv.listing_price" class="text-lg font-bold text-primary-500">
              {{ formatCurrency(conv.listing_price) }}
            </p>
            <span v-if="conv.pending_count > 0" class="inline-block mt-1 px-2 py-1 bg-yellow-500/20 text-yellow-400 text-xs rounded">
              {{ conv.pending_count }} att granska
            </span>
          </div>
        </div>
        
        <div class="flex justify-between items-center text-sm text-slate-400 mt-2">
          <span :class="statusClass(conv.status)">{{ conv.status }}</span>
          <span>{{ formatDate(conv.updated_at) }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'

interface Conversation {
  id: number
  listing_id: number
  marketplace_id: number
  status: string
  created_at: string
  updated_at: string
  listing_title: string
  listing_price?: number
  marketplace_name: string
  pending_count: number
}

const api = useApi()
const conversations = ref<Conversation[]>([])
const loading = ref(true)
const showNeedsReview = ref(true)

const fetchConversations = async () => {
  loading.value = true
  try {
    const endpoint = showNeedsReview.value 
      ? '/conversations?needs_review=true'
      : '/conversations'
    conversations.value = await api.get<Conversation[]>(endpoint)
  } catch (error) {
    console.error('Failed to fetch conversations:', error)
  } finally {
    loading.value = false
  }
}

const formatCurrency = (amount: number) => {
  return new Intl.NumberFormat('sv-SE', { style: 'currency', currency: 'SEK' }).format(amount / 100)
}

const formatDate = (dateStr: string) => {
  const date = new Date(dateStr)
  return new Intl.DateTimeFormat('sv-SE', { 
    year: 'numeric', 
    month: 'short', 
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  }).format(date)
}

const statusClass = (status: string) => {
  const classes: Record<string, string> = {
    active: 'px-2 py-1 bg-green-500/20 text-green-400 text-xs rounded',
    closed: 'px-2 py-1 bg-gray-500/20 text-gray-400 text-xs rounded',
    archived: 'px-2 py-1 bg-slate-500/20 text-slate-400 text-xs rounded'
  }
  return classes[status] || 'px-2 py-1 bg-slate-500/20 text-slate-400 text-xs rounded'
}

watch(showNeedsReview, () => {
  fetchConversations()
})

onMounted(() => {
  fetchConversations()
})
</script>

<style scoped>
.page-header {
  @apply text-3xl font-bold text-slate-100;
}

.card {
  @apply bg-slate-800 rounded-lg border border-slate-700;
}

.btn {
  @apply px-4 py-2 rounded-lg font-medium transition-colors;
}

.btn-primary {
  @apply bg-primary-600 text-white hover:bg-primary-700;
}

.btn-secondary {
  @apply bg-slate-700 text-slate-200 hover:bg-slate-600;
}
</style>
