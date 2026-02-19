type HttpMethod = 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE'

interface ApiOptions {
  method?: HttpMethod
  body?: Record<string, unknown>
}

export const useApi = () => {
  const config = useRuntimeConfig()
  const loadingStore = useLoadingStore()
  const client = useSupabaseClient()
  
  const apiBase = config.public.apiBase

  const fetch = async <T>(endpoint: string, options?: ApiOptions): Promise<T> => {
    const url = `${apiBase}/api${endpoint.startsWith('/') ? endpoint : `/${endpoint}`}`
    console.debug('useApi request', { url, method: options?.method || 'GET', body: options?.body })

    loadingStore.startLoading()

    try {
      // Get the current session token
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
      console.debug('useApi response', { url, result })
      return result
    } catch (err) {
      console.error('useApi error', { url, err })
      throw err
    } finally {
      loadingStore.stopLoading()
    }
  }

  const get = <T>(endpoint: string): Promise<T> => {
    return fetch<T>(endpoint, { method: 'GET' })
  }

  const post = <T>(endpoint: string, body: Record<string, unknown>): Promise<T> => {
    return fetch<T>(endpoint, { method: 'POST', body })
  }

  const put = <T>(endpoint: string, body: Record<string, unknown>): Promise<T> => {
    return fetch<T>(endpoint, { method: 'PUT', body })
  }

  const patch = <T>(endpoint: string, body: Record<string, unknown>): Promise<T> => {
    return fetch<T>(endpoint, { method: 'PATCH', body })
  }

  const del = <T>(endpoint: string): Promise<T> => {
    return fetch<T>(endpoint, { method: 'DELETE' })
  }

  return {
    fetch,
    get,
    post,
    put,
    patch,
    delete: del
  }
}