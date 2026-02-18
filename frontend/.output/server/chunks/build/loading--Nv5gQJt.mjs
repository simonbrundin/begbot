import { defineStore } from 'pinia';
import { ref, computed } from 'vue';

const useLoadingStore = defineStore("loading", () => {
  const requestCount = ref(0);
  const minimumDelay = 200;
  const showSpinner = ref(false);
  let timeoutId = null;
  const isLoading = computed(() => requestCount.value > 0);
  const startLoading = () => {
    requestCount.value++;
    if (timeoutId) {
      clearTimeout(timeoutId);
    }
    if (!showSpinner.value) {
      timeoutId = setTimeout(() => {
        if (isLoading.value) {
          showSpinner.value = true;
        }
      }, minimumDelay);
    }
  };
  const stopLoading = () => {
    requestCount.value = Math.max(0, requestCount.value - 1);
    if (requestCount.value === 0) {
      if (timeoutId) {
        clearTimeout(timeoutId);
        timeoutId = null;
      }
      showSpinner.value = false;
    }
  };
  const reset = () => {
    requestCount.value = 0;
    if (timeoutId) {
      clearTimeout(timeoutId);
      timeoutId = null;
    }
    showSpinner.value = false;
  };
  return {
    requestCount,
    showSpinner,
    isLoading,
    startLoading,
    stopLoading,
    reset
  };
});

export { useLoadingStore as u };
//# sourceMappingURL=loading--Nv5gQJt.mjs.map
