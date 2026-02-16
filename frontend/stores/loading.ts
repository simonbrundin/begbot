import { defineStore } from 'pinia'

export const useLoadingStore = defineStore('loading', () => {
  const requestCount = ref(0)
  const minimumDelay = 200 // ms before showing spinner
  const showSpinner = ref(false)
  let timeoutId: ReturnType<typeof setTimeout> | null = null

  const isLoading = computed(() => requestCount.value > 0)

  const startLoading = () => {
    requestCount.value++
    
    // Clear any existing timeout
    if (timeoutId) {
      clearTimeout(timeoutId)
    }
    
    // Set timeout to show spinner after delay
    if (!showSpinner.value) {
      timeoutId = setTimeout(() => {
        if (isLoading.value) {
          showSpinner.value = true
        }
      }, minimumDelay)
    }
  }

  const stopLoading = () => {
    requestCount.value = Math.max(0, requestCount.value - 1)
    
    // If no more requests, hide spinner and clear timeout
    if (requestCount.value === 0) {
      if (timeoutId) {
        clearTimeout(timeoutId)
        timeoutId = null
      }
      showSpinner.value = false
    }
  }

  // Reset function for error cases
  const reset = () => {
    requestCount.value = 0
    if (timeoutId) {
      clearTimeout(timeoutId)
      timeoutId = null
    }
    showSpinner.value = false
  }

  return {
    requestCount,
    showSpinner,
    isLoading,
    startLoading,
    stopLoading,
    reset
  }
})
