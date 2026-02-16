<template>
  <div class="min-h-screen bg-slate-900 text-slate-100">
    <LoadingSpinner :visible="loadingStore.showSpinner" />
    <aside
      class="fixed left-0 top-0 h-full w-64 bg-slate-800 border-r border-slate-700 p-4"
    >
      <div class="mb-8">
        <h1 class="text-xl font-bold text-emerald-400">Begbot</h1>
      </div>
      <nav class="space-y-1">
        <NuxtLink
          to="/"
          class="flex items-center gap-3 px-4 py-2 rounded-lg hover:bg-slate-700 text-slate-300 hover:text-white"
          active-class="bg-slate-700 text-white"
        >
          <span>Översikt</span>
        </NuxtLink>
        <NuxtLink
          to="/products"
          class="flex items-center gap-3 px-4 py-2 rounded-lg hover:bg-slate-700 text-slate-300 hover:text-white"
          active-class="bg-slate-700 text-white"
        >
          <span>Produkter</span>
        </NuxtLink>
        <NuxtLink
          to="/listings"
          class="flex items-center gap-3 px-4 py-2 rounded-lg hover:bg-slate-700 text-slate-300 hover:text-white"
          active-class="bg-slate-700 text-white"
        >
          <span>Mina annonser</span>
        </NuxtLink>
        <NuxtLink
          to="/transactions"
          class="flex items-center gap-3 px-4 py-2 rounded-lg hover:bg-slate-700 text-slate-300 hover:text-white"
          active-class="bg-slate-700 text-white"
        >
          <span>Transaktioner</span>
        </NuxtLink>
        <NuxtLink
          to="/analytics"
          class="flex items-center gap-3 px-4 py-2 rounded-lg hover:bg-slate-700 text-slate-300 hover:text-white"
          active-class="bg-slate-700 text-white"
        >
          <span>Marknadsanalys</span>
        </NuxtLink>
        <NuxtLink
          to="/scraping"
          class="flex items-center gap-3 px-4 py-2 rounded-lg hover:bg-slate-700 text-slate-300 hover:text-white"
          active-class="bg-slate-700 text-white"
        >
          <span>Scraping</span>
        </NuxtLink>
        <NuxtLink
          to="/ads"
          class="flex items-center gap-3 px-4 py-2 rounded-lg hover:bg-slate-700 text-slate-300 hover:text-white"
          active-class="bg-slate-700 text-white"
        >
          <span>Hittade annonser</span>
        </NuxtLink>
      </nav>
    </aside>
    <main class="ml-64 p-8">
      <slot />
    </main>
  </div>
</template>

<script setup lang="ts">
const loadingStore = useLoadingStore()
</script>
