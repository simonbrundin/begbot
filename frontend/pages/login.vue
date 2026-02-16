<template>
  <div class="min-h-screen flex items-center justify-center bg-slate-900">
    <div class="max-w-md w-full">
      <div class="card p-8">
        <div class="text-center mb-8">
          <h1 class="text-2xl font-bold text-primary-500">Begbot</h1>
          <p class="text-slate-400 mt-2">Logga in för att fortsätta</p>
        </div>

        <form @submit.prevent="handleLogin" class="space-y-4">
          <div>
            <label class="label">E-post</label>
            <input
              v-model="email"
              type="email"
              class="input"
              placeholder="du@example.com"
              required
            />
          </div>

          <div>
            <label class="label">Lösenord</label>
            <input
              v-model="password"
              type="password"
              class="input"
              placeholder="••••••••"
              required
            />
          </div>

          <div v-if="error" class="p-3 bg-red-900/50 text-red-400 rounded-lg text-sm border border-red-800">
            {{ error }}
          </div>

          <button
            type="submit"
            class="btn btn-primary w-full"
            :disabled="loading"
          >
            <span v-if="loading">Loggar in...</span>
            <span v-else>Logga in</span>
          </button>
        </form>

        <p class="text-center text-sm text-slate-500 mt-6">
          Använd dina Supabase-uppgifter
        </p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const client = useSupabaseClient()
const user = useSupabaseUser()
const router = useRouter()

const email = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)

watchEffect(() => {
  if (user.value) {
    router.push('/')
  }
})

const handleLogin = async () => {
  loading.value = true
  error.value = ''
  
  try {
    const { error: authError } = await client.auth.signInWithPassword({
      email: email.value,
      password: password.value
    })
    
    if (authError) {
      error.value = authError.message
    }
    } catch (e) {
    error.value = 'Ett fel uppstod vid inloggning'
  } finally {
    loading.value = false
  }
}
</script>
