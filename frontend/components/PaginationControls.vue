<template>
  <div v-if="totalPages > 1" class="flex items-center justify-between px-4 py-3 bg-slate-800 border-t border-slate-700">
    <div class="text-sm text-slate-400">
      Visar {{ startItem }}-{{ endItem }} av {{ totalCount }} poster
    </div>
    
    <div class="flex items-center gap-2">
      <button
        @click="$emit('update:page', page - 1)"
        :disabled="page <= 1"
        class="px-3 py-1.5 rounded text-sm transition-colors"
        :class="page <= 1 
          ? 'bg-slate-700 text-slate-500 cursor-not-allowed' 
          : 'bg-slate-700 text-slate-300 hover:bg-slate-600'"
      >
        Föregående
      </button>
      
      <div class="flex items-center gap-1">
        <template v-for="p in visiblePages" :key="p">
          <span v-if="p === '...'" class="px-2 text-slate-500">...</span>
          <button
            v-else
            @click="$emit('update:page', p as number)"
            class="px-3 py-1.5 rounded text-sm transition-colors"
            :class="page === p 
              ? 'bg-primary-600 text-white' 
              : 'bg-slate-700 text-slate-300 hover:bg-slate-600'"
          >
            {{ p }}
          </button>
        </template>
      </div>
      
      <button
        @click="$emit('update:page', page + 1)"
        :disabled="page >= totalPages"
        class="px-3 py-1.5 rounded text-sm transition-colors"
        :class="page >= totalPages 
          ? 'bg-slate-700 text-slate-500 cursor-not-allowed' 
          : 'bg-slate-700 text-slate-300 hover:bg-slate-600'"
      >
        Nästa
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  page: number
  pageSize: number
  totalCount: number
  totalPages: number
}>()

defineEmits<{
  'update:page': [page: number]
}>()

const startItem = computed(() => {
  if (props.totalCount === 0) return 0
  return (props.page - 1) * props.pageSize + 1
})

const endItem = computed(() => {
  const end = props.page * props.pageSize
  return Math.min(end, props.totalCount)
})

const visiblePages = computed(() => {
  const pages: (number | string)[] = []
  const total = props.totalPages
  const current = props.page
  
  if (total <= 7) {
    for (let i = 1; i <= total; i++) {
      pages.push(i)
    }
  } else {
    pages.push(1)
    
    if (current > 3) {
      pages.push('...')
    }
    
    const start = Math.max(2, current - 1)
    const end = Math.min(total - 1, current + 1)
    
    for (let i = start; i <= end; i++) {
      pages.push(i)
    }
    
    if (current < total - 2) {
      pages.push('...')
    }
    
    pages.push(total)
  }
  
  return pages
})
</script>
