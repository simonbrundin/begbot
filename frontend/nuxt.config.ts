export default defineNuxtConfig({
  devtools: { enabled: false },
  modules: [
    '@nuxtjs/tailwindcss',
    '@nuxtjs/supabase',
    '@pinia/nuxt',
    '@nuxt/icon',
    '@nuxt/test-utils/module'
  ],
  supabase: {
    redirect: false,
    redirectOptions: {
      login: '/login',
      callback: '/confirm',
      exclude: ['/', '/listings', '/products', '/transactions', '/analytics', '/scraping', '/ads']
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