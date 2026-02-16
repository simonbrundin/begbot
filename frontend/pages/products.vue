<template>
  <div>
    <div class="flex justify-between items-center mb-6">
      <h1 class="page-header">Produkter</h1>
      <button @click="showAddModal = true" class="btn btn-primary">
        Lägg till produkt
      </button>
    </div>

    <div class="card overflow-hidden">
      <table class="table">
        <thead>
          <tr>
            <th>Märke</th>
            <th>Namn</th>
            <th>Kategori</th>
            <th>Variant</th>
            <th>Aktiverad</th>
            <th>Skapad</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="product in products" :key="product.id">
            <td class="font-medium text-slate-100">{{ product.brand || '-' }}</td>
            <td>{{ product.name || '-' }}</td>
            <td>{{ product.category || '-' }}</td>
            <td>{{ product.model_variant || '-' }}</td>
            <td>
              <button
                @click="toggleEnabled(product)"
                :class="product.enabled ? 'badge badge-success' : 'badge'"
              >
                {{ product.enabled ? 'Ja' : 'Nej' }}
              </button>
            </td>
            <td class="text-sm text-slate-400">{{ formatDate(product.created_at) }}</td>
            <td>
              <button @click="editProduct(product)" class="text-primary-400 hover:text-primary-300">
                Redigera
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="showAddModal || editingProduct" class="fixed inset-0 bg-black/70 flex items-center justify-center z-50">
      <div class="bg-slate-800 rounded-lg p-6 w-full max-w-lg border border-slate-700">
        <h2 class="text-xl font-bold text-slate-100 mb-4">
          {{ editingProduct ? 'Redigera produkt' : 'Lägg till ny produkt' }}
        </h2>

        <form @submit.prevent="saveProduct" class="space-y-4">
          <div>
            <label class="label">Märke</label>
            <input v-model="form.brand" type="text" class="input" />
          </div>
          <div>
            <label class="label">Namn</label>
            <input v-model="form.name" type="text" class="input" />
          </div>
          <div>
            <label class="label">Kategori</label>
            <input v-model="form.category" type="text" class="input" placeholder="t.ex., telefon" />
          </div>
          <div>
            <label class="label">Modellvariant</label>
            <input v-model="form.model_variant" type="text" class="input" placeholder="t.ex., 256GB" />
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="label">Förpackning (öre)</label>
              <input v-model.number="form.sell_packaging_cost" type="number" class="input" />
            </div>
            <div>
              <label class="label">Frakt (öre)</label>
              <input v-model.number="form.sell_postage_cost" type="number" class="input" />
            </div>
          </div>

          <div class="flex justify-end gap-2 pt-4">
            <button type="button" @click="closeModal" class="btn btn-secondary">
              Avbryt
            </button>
            <button type="submit" class="btn btn-primary">
              {{ editingProduct ? 'Spara' : 'Lägg till' }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Product } from '~/types/database'

const api = useApi()

const products = ref<Product[]>([])
const loading = ref(false)
const showAddModal = ref(false)
const editingProduct = ref<Product | null>(null)

const defaultForm = {
  brand: '',
  name: '',
  category: '',
  model_variant: '',
  sell_packaging_cost: 0,
  sell_postage_cost: 0,
  enabled: false
}

const form = ref({ ...defaultForm })

const fetchData = async () => {
  loading.value = true
  try {
    products.value = await api.get<Product[]>('/products')
  } catch (e) {
    console.error('Failed to fetch products:', e)
  } finally {
    loading.value = false
  }
}

const formatDate = (dateStr: string) => new Date(dateStr).toLocaleDateString('sv-SE')

const editProduct = (product: Product) => {
  editingProduct.value = product
  form.value = {
    brand: product.brand || '',
    name: product.name || '',
    category: product.category || '',
    model_variant: product.model_variant || '',
    sell_packaging_cost: product.sell_packaging_cost,
    sell_postage_cost: product.sell_postage_cost,
    enabled: product.enabled
  }
}

const closeModal = () => {
  showAddModal.value = false
  editingProduct.value = null
  form.value = { ...defaultForm }
}

const saveProduct = async () => {
  try {
    if (editingProduct.value) {
      await api.put(`/products/${editingProduct.value.id}`, form.value)
    } else {
      await api.post('/products', form.value)
    }
    closeModal()
    await fetchData()
  } catch (e) {
    console.error('Failed to save product:', e)
  }
}

const toggleEnabled = async (product: Product) => {
  try {
    await api.put(`/products/${product.id}`, { ...product, enabled: !product.enabled })
    await fetchData()
  } catch (e) {
    console.error('Failed to toggle enabled:', e)
  }
}

onMounted(fetchData)
</script>
