<template>
  <div>
    <h2 class="text-2xl font-bold mb-6">Inställningar</h2>
    <div class="max-w-2xl">
      <EmailSettings />
      <SaveSettings />
      <EconomySettings />
      <AutoEnableSettings />
    </div>
  </div>
</template>

<script setup lang="ts">
import EmailSettings from '~/components/settings/EmailSettings.vue'
import SaveSettings from '~/components/settings/SaveSettings.vue'
import EconomySettings from '~/components/settings/EconomySettings.vue'
import AutoEnableSettings from '~/components/settings/AutoEnableSettings.vue'
</script>
