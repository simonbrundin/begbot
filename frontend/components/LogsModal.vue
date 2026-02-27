<template>
  <Transition
    enter-active-class="transition-opacity duration-200 ease-out"
    enter-from-class="opacity-0"
    enter-to-class="opacity-100"
    leave-active-class="transition-opacity duration-150 ease-in"
    leave-from-class="opacity-100"
    leave-to-class="opacity-0"
  >
    <div v-if="show" class="fixed inset-0 bg-black/70 flex items-center justify-center z-50" @click.self="$emit('close')">
      <div class="bg-slate-800 rounded-lg w-full max-w-4xl max-h-[80vh] border border-slate-700 shadow-xl flex flex-col">
        <div class="flex justify-between items-center p-4 border-b border-slate-700">
          <div>
            <h2 class="text-lg font-bold text-slate-100">Loggar</h2>
            <p class="text-sm text-slate-400">Scraping run #{{ runId }}</p>
          </div>
          <button @click="$emit('close')" class="text-slate-400 hover:text-slate-200">
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        
        <div class="flex-1 overflow-hidden">
          <div 
            ref="logContainer"
            class="bg-slate-950 p-4 font-mono text-sm h-[50vh] overflow-y-auto"
          >
            <div v-if="loading" class="text-slate-500 italic">
              Laddar loggar...
            </div>
            
            <div v-else-if="error" class="text-red-400">
              {{ error }}
            </div>
            
            <div v-else-if="logs.length === 0" class="text-slate-500 italic">
              Inga loggar tillgängliga
            </div>
            
            <div v-else>
              <div 
                v-for="(log, index) in logs" 
                :key="log.id || index"
                class="flex gap-3 py-1"
              >
                <span class="text-slate-500 shrink-0">
                  {{ formatTime(log.created_at) }}
                </span>
                <span 
                  class="shrink-0 w-16 font-semibold"
                  :class="getLevelClass(log.level)"
                >
                  {{ log.level.toUpperCase() }}
                </span>
                <span class="text-slate-300 break-all">
                  {{ log.message }}
                </span>
              </div>
            </div>
          </div>
        </div>
        
        <div class="flex justify-end gap-3 p-4 border-t border-slate-700">
          <button @click="$emit('close')" class="btn btn-secondary">
            Stäng
          </button>
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
import type { ScrapingRunLog } from "~/types/database"

interface Props {
  show: boolean
  runId: number
  logs: ScrapingRunLog[]
  loading?: boolean
  error?: string | null
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
  error: null
})

defineEmits<{
  close: []
}>()

const logContainer = ref<HTMLDivElement>()

watch(() => props.logs.length, () => {
  nextTick(() => {
    if (logContainer.value) {
      logContainer.value.scrollTop = logContainer.value.scrollHeight
    }
  })
})

watch(() => props.show, (newVal) => {
  if (newVal) {
    nextTick(() => {
      if (logContainer.value) {
        logContainer.value.scrollTop = logContainer.value.scrollHeight
      }
    })
  }
})

const formatTime = (timestamp: string): string => {
  const date = new Date(timestamp)
  return date.toLocaleString('sv-SE', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}

const getLevelClass = (level: string): string => {
  switch (level) {
    case 'error':
      return 'text-red-400'
    case 'warning':
      return 'text-amber-400'
    case 'info':
    default:
      return 'text-blue-400'
  }
}
</script>
