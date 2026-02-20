<template>
  <Transition
    enter-active-class="transition-opacity duration-200 ease-out"
    enter-from-class="opacity-0"
    enter-to-class="opacity-100"
    leave-active-class="transition-opacity duration-150 ease-in"
    leave-from-class="opacity-100"
    leave-to-class="opacity-0"
  >
    <div v-if="show" class="fixed inset-0 bg-black/70 flex items-center justify-center z-50" @click.self="handleCancel">
      <div class="bg-slate-800 rounded-lg p-6 w-full max-w-md border border-slate-700 shadow-xl">
        <h2 class="text-lg font-bold text-slate-100 mb-3">{{ title }}</h2>
        <p class="text-slate-300 mb-6">{{ message }}</p>
        <div class="flex justify-end gap-3">
          <button @click="handleCancel" class="btn btn-secondary">
            {{ cancelText }}
          </button>
          <button @click="handleConfirm" :class="confirmButtonClass || 'btn bg-red-600 hover:bg-red-500 text-white border-none'">
            {{ confirmText }}
          </button>
        </div>
      </div>
    </div>
  </Transition>
</template>

<script setup lang="ts">
defineProps<{
  show: boolean
  title: string
  message: string
  confirmText?: string
  cancelText?: string
  confirmButtonClass?: string
}>()

const emit = defineEmits<{
  confirm: []
  cancel: []
}>()

const handleConfirm = () => {
  emit('confirm')
}

const handleCancel = () => {
  emit('cancel')
}
</script>
