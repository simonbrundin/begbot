<template>
  <Transition
    enter-active-class="transition-opacity duration-300 ease-out"
    enter-from-class="opacity-0"
    enter-to-class="opacity-100"
    leave-active-class="transition-opacity duration-200 ease-in"
    leave-from-class="opacity-100"
    leave-to-class="opacity-0"
  >
    <div
      v-if="visible"
      class="fixed inset-0 bg-slate-900/60 backdrop-blur-sm flex items-center justify-center z-50"
      aria-label="Laddar..."
      role="status"
    >
      <div class="relative">
        <!-- Outer ring -->
        <div class="w-16 h-16 rounded-full border-4 border-slate-700"></div>
        <!-- Spinning ring -->
        <div class="absolute inset-0 w-16 h-16 rounded-full border-4 border-emerald-400 border-t-transparent animate-spin"></div>
        <!-- Inner dot -->
        <div class="absolute inset-0 flex items-center justify-center">
          <div class="w-2 h-2 rounded-full bg-emerald-400 animate-pulse"></div>
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
defineProps<{
  visible: boolean
}>()
</script>
