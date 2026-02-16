import { ref, onUnmounted, watch, readonly } from 'vue'
import type { Ref } from 'vue'

export interface LogEntry {
  timestamp: string
  level: 'info' | 'warning' | 'error'
  message: string
}

export function useScraperLogs(jobId: Ref<string | null>) {
  const logs = ref<LogEntry[]>([])
  const isConnected = ref(false)
  const error = ref<string | null>(null)
  
  let eventSource: EventSource | null = null
  let reconnectAttempts = 0
  const MAX_RECONNECT_ATTEMPTS = 3

  const connect = () => {
    if (!jobId.value) return
    
    // Clear previous logs when connecting to a new job
    logs.value = []
    error.value = null
    reconnectAttempts = 0
    
    const config = useRuntimeConfig()
    const apiBase = config.public.apiBase || 'http://localhost:8081'
    const url = `${apiBase}/api/fetch-ads/logs/${jobId.value}`
    
    console.log('Connecting to SSE:', url)
    eventSource = new EventSource(url)
    
    eventSource.onopen = () => {
      console.log('SSE connection opened')
      isConnected.value = true
      error.value = null
      reconnectAttempts = 0
    }
    
    eventSource.onmessage = (event) => {
      console.log('SSE message received:', event.data)
      try {
        const log = JSON.parse(event.data) as LogEntry
        logs.value.push(log)
      } catch (e) {
        console.error('Failed to parse log entry:', e)
      }
    }
    
    eventSource.onerror = (err) => {
      console.error('SSE error:', err)
      isConnected.value = false
      
      if (reconnectAttempts < MAX_RECONNECT_ATTEMPTS) {
        reconnectAttempts++
        console.log(`Reconnecting... attempt ${reconnectAttempts}`)
        setTimeout(() => {
          disconnect()
          connect()
        }, 1000 * reconnectAttempts)
      } else {
        error.value = 'Anslutningen bröts. Försök uppdatera sidan.'
      }
    }
  }
  
  const disconnect = () => {
    if (eventSource) {
      eventSource.close()
      eventSource = null
    }
    isConnected.value = false
  }
  
  const clearLogs = () => {
    logs.value = []
  }
  
  // Watch for jobId changes
  watch(jobId, (newJobId, oldJobId) => {
    console.log('JobId changed:', newJobId, 'old:', oldJobId)
    if (newJobId && newJobId !== oldJobId) {
      disconnect()
      connect()
    } else if (!newJobId) {
      disconnect()
      logs.value = []
    }
  })
  
  // Cleanup on unmount
  onUnmounted(() => {
    disconnect()
  })
  
  return {
    logs,
    isConnected,
    error,
    connect,
    disconnect,
    clearLogs
  }
}
