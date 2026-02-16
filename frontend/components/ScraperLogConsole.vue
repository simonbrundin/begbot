<template>
  <div class="card p-4 mt-4">
    <div class="flex justify-between items-center mb-3">
      <h3 class="text-sm font-medium text-slate-300">Scraper-logg</h3>
      <div class="flex items-center gap-3">
        <span 
          v-if="isConnected" 
          class="flex items-center gap-1 text-xs text-emerald-400"
        >
          <span class="w-2 h-2 bg-emerald-400 rounded-full animate-pulse"></span>
          Live
        </span>
        <button 
          @click="$emit('clear')" 
          class="text-xs text-slate-400 hover:text-slate-300"
          :disabled="logs.length === 0"
        >
          Rensa
        </button>
      </div>
    </div>
    
    <div 
      ref="logContainer"
      class="bg-slate-950 rounded-lg p-3 font-mono text-xs h-64 overflow-y-auto space-y-1"
    >
      <div v-if="logs.length === 0" class="text-slate-600 italic">
        Väntar på loggmeddelanden...
      </div>
      
      <div 
        v-for="(log, index) in logs" 
        :key="index"
        class="flex gap-2"
      >
        <span class="text-slate-500 shrink-0">
          {{ formatTime(log.timestamp) }}
        </span>
        <span 
          class="shrink-0 w-14"
          :class="getLevelClass(log.level)"
        >
          [{{ log.level.toUpperCase() }}]
        </span>
        <span class="text-slate-300 break-all">
          {{ log.message }}
        </span>
      </div>
    </div>
    
    <div v-if="error" class="mt-2 text-sm text-red-400">
      {{ error }}
    </div>
  </div>
</template>

<script setup lang="ts">
interface LogEntry {
  timestamp: string
  level: 'info' | 'warning' | 'error'
  message: string
}

interface Props {
  logs: LogEntry[]
  isConnected: boolean
  error: string | null
}

const props = defineProps<Props>()
defineEmits<{
  clear: []
}>()

const logContainer = ref<HTMLDivElement>()

// Auto-scroll to bottom when new logs arrive
watch(() => props.logs.length, () => {
  nextTick(() => {
    if (logContainer.value) {
      logContainer.value.scrollTop = logContainer.value.scrollHeight
    }
  })
})

const formatTime = (timestamp: string): string => {
  const date = new Date(timestamp)
  return date.toLocaleTimeString('sv-SE', { 
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
