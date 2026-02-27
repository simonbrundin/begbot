## Frontend API Composable

### useApi Pattern

Use the `useApi` composable for all API calls. It handles authentication, loading state, and provides typed methods.

```typescript
export const useApi = () => {
  const config = useRuntimeConfig()
  const loadingStore = useLoadingStore()
  const client = useSupabaseClient()
  
  const apiBase = config.public.apiBase

  const fetch = async <T>(endpoint: string, options?: ApiOptions): Promise<T> => {
    const url = `${apiBase}/api${endpoint.startsWith('/') ? endpoint : `/${endpoint}`}`
    
    loadingStore.startLoading()

    try {
      const { data: { session } } = await client.auth.getSession()
      const headers: Record<string, string> = {}
      
      if (session?.access_token) {
        headers['Authorization'] = `Bearer ${session.access_token}`
      }

      const result = await $fetch<T>(url, {
        method: options?.method || 'GET',
        body: options?.body,
        headers
      })
      return result
    } catch (err) {
      console.error('useApi error', { url, err })
      throw err
    } finally {
      loadingStore.stopLoading()
    }
  }

  // Convenience methods
  const get = <T>(endpoint: string): Promise<T> => fetch<T>(endpoint, { method: 'GET' })
  const post = <T>(endpoint: string, body: Record<string, unknown>): Promise<T> => 
    fetch<T>(endpoint, { method: 'POST', body })
  const put = <T>(endpoint: string, body: Record<string, unknown>): Promise<T> => 
    fetch<T>(endpoint, { method: 'PUT', body })
  const patch = <T>(endpoint: string, body: Record<string, unknown>): Promise<T> => 
    fetch<T>(endpoint, { method body })
  const: 'PATCH', del = <T>(endpoint: string): Promise<T> => fetch<T>(endpoint, { method: 'DELETE' })

  return { fetch, get, post, put, patch, delete: del }
}
```

### Usage

```typescript
const api = useApi()

// GET
const users = await api.get<User[]>('/users')

// POST
const newUser = await api.post<User>('/users', { name: 'John' })

// With typed response
interface CreateProductResponse { id: number }
const result = await api.post<CreateProductResponse>('/products', productData)
```

### Loading State

The composable automatically tracks loading state via `useLoadingStore`. This store uses a request count with minimum delay to avoid flashing spinners on fast requests.
