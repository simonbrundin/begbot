export default defineNuxtConfig({
  devtools: { enabled: false },
  modules: [
    '@nuxtjs/tailwindcss',
    '@nuxtjs/supabase',
    '@pinia/nuxt',
    '@nuxt/icon'
  ],
  supabase: {
    redirect: false,
    // In local dev we prefer client-side session persistence (localStorage)
    // to avoid SSR cookie/domain issues that can cause sessions to be lost on refresh.
    useSsrCookies: false,
    // Be explicit about client options so session persistence is consistent across reloads
    clientOptions: {
      auth: {
        persistSession: true,
        detectSessionInUrl: false,
        // Use a clear storage key so it's easy to inspect in DevTools
        storageKey: 'supabase.auth'
      }
    },
    redirectOptions: {
      login: '/login',
      callback: '/confirm',
      exclude: ['/', '/listings', '/products', '/transactions', '/analytics', '/scraping', '/ads', '/conversations']
    }
  },
  css: ['~/assets/css/main.css'],
  app: {
    head: {
      title: 'Begbot Dashboard',
      meta: [
        { name: 'description', content: 'Inventory and listing management' }
      ]
    }
  },
  runtimeConfig: {
    public: {
      apiBase: process.env.API_BASE_URL || 'http://localhost:8081'
    }
  },
  compatibilityDate: '2024-11-01'
})
