<template>
  <div>
    <div class="flex justify-between items-center mb-6">
      <h1 class="page-header">Mina annonser</h1>
      <button @click="showAddModal = true" class="btn btn-primary">
        Lägg till annons
      </button>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <div v-for="listing in listings" :key="listing.id" class="card overflow-hidden">
        <div class="p-4">
          <div class="flex justify-between items-start mb-2">
              <div>
               <p v-if="listing.product_id" class="font-medium text-slate-100">
                 {{ getProductName(listing.product_id) }}
               </p>
               <p v-else class="text-slate-500">Okänd produkt</p>
               <p class="text-lg font-bold text-primary-500">
                 {{ listing.price ? formatCurrency(listing.price) : '-' }}
               </p>
               <p v-if="listing.valuation" class="text-sm text-slate-400">
                 Värdering: {{ formatCurrency(listing.valuation) }}
               </p>
             </div>
             <span :class="statusClass(listing.status)">
               {{ listing.status }}
             </span>
           </div>

          <p class="text-sm text-slate-400 mb-2">
            {{ listing.description?.substring(0, 100) }}...
          </p>

          <div class="flex justify-between items-center text-sm text-slate-400">
            <span>{{ marketplaceName(listing.marketplace_id) }}</span>
            <a :href="listing.link" target="_blank" class="text-primary-400 hover:text-primary-300">
              Visa
            </a>
          </div>
        </div>

        <div class="px-4 py-3 bg-slate-800/50 border-t border-slate-700 flex justify-end gap-2">
          <button @click="editListing(listing)" class="text-sm text-primary-400 hover:text-primary-300">
            Redigera
          </button>
          <button @click="deleteListing(listing.id)" class="text-sm text-red-400 hover:text-red-300">
            Ta bort
          </button>
        </div>
      </div>
    </div>

    <div v-if="listings.length === 0" class="text-center py-12 text-slate-500">
      Inga annonser hittades. Lägg till din första annons!
    </div>

    <div v-if="showAddModal || editingListing" class="fixed inset-0 bg-black/70 flex items-center justify-center z-50">
      <div class="bg-slate-800 rounded-lg p-6 w-full max-w-2xl max-h-[90vh] overflow-y-auto border border-slate-700">
        <h2 class="text-xl font-bold text-slate-100 mb-4">
          {{ editingListing ? 'Redigera annons' : 'Lägg till ny annons' }}
        </h2>

        <form @submit.prevent="saveListing" class="space-y-4">
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="label">Produkt</label>
              <select v-model="form.product_id" class="input">
                <option value="">Välj produkt...</option>
                <option v-for="p in products" :key="p.id" :value="p.id">
                  {{ p.brand }} {{ p.name }}
                </option>
              </select>
            </div>
            <div>
              <label class="label">Pris (öre)</label>
              <input v-model.number="form.price" type="number" class="input" />
            </div>
            <div>
              <label class="label">Status</label>
              <select v-model="form.status" class="input">
                <option v-for="status in LISTING_STATUSES" :key="status" :value="status">
                  {{ status }}
                </option>
              </select>
            </div>
            <div>
              <label class="label">Marknadsplats</label>
              <select v-model="form.marketplace_id" class="input">
                <option value="">Välj marknadsplats...</option>
                <option v-for="m in marketplaces" :key="m.id" :value="m.id">
                  {{ m.name }}
                </option>
              </select>
            </div>
            <div class="col-span-2">
              <label class="label">Länk</label>
              <input v-model="form.link" type="text" class="input" />
            </div>
            <div class="col-span-2">
              <label class="label">Beskrivning</label>
              <textarea v-model="form.description" class="input" rows="3"></textarea>
            </div>
          </div>

          <div class="flex justify-end gap-2 pt-4">
            <button type="button" @click="closeModal" class="btn btn-secondary">
              Avbryt
            </button>
            <button type="submit" class="btn btn-primary">
              {{ editingListing ? 'Spara' : 'Lägg till' }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Listing, Product, Marketplace } from '~/types/database'
import { LISTING_STATUSES } from '~/types/database'

const api = useApi()

const listings = ref<Listing[]>([])
const products = ref<Product[]>([])
const marketplaces = ref<Marketplace[]>([])
const loading = ref(false)
const showAddModal = ref(false)
const editingListing = ref<Listing | null>(null)

const defaultForm = {
  product_id: null as number | null,
  price: null as number | null,
  link: '',
  description: '',
  marketplace_id: null as number | null,
  status: 'draft' as const,
  is_my_listing: true
}

const form = ref({ ...defaultForm })

const fetchData = async () => {
  loading.value = true
  try {
    const [listingsRes, productsRes, marketplacesRes] = await Promise.all([
      api.get<Listing[]>('/listings?mine=true'),
      api.get<Product[]>('/products'),
      api.get<Marketplace[]>('/marketplaces')
    ])
    listings.value = listingsRes
    products.value = productsRes
    marketplaces.value = marketplacesRes
  } catch (e) {
    console.error('Failed to fetch data:', e)
  } finally {
    loading.value = false
  }
}

const formatCurrency = (cents: number) => `${(cents / 100).toFixed(2)} kr`

const statusClass = (status: string) => {
  const classes: Record<string, string> = {
    draft: 'badge badge-warning',
    active: 'badge badge-success',
    sold: 'badge badge-info',
    archived: 'badge'
  }
  return classes[status] || classes.draft
}

const marketplaceName = (id: number | null) => {
  if (!id) return 'Unknown'
  return marketplaces.value.find(m => m.id === id)?.name || 'Unknown'
}

const getProductName = (productId: number | null) => {
  if (!productId) return 'Okänd produkt'
  const product = products.value.find(p => p.id === productId)
  if (!product) return 'Okänd produkt'
  return `${product.brand} ${product.name}`
}

const editListing = (listing: Listing) => {
  editingListing.value = listing
  form.value = {
    product_id: listing.product_id,
    price: listing.price,
    link: listing.link,
    description: listing.description,
    marketplace_id: listing.marketplace_id,
    status: listing.status as any,
    is_my_listing: true
  }
}

const closeModal = () => {
  showAddModal.value = false
  editingListing.value = null
  form.value = { ...defaultForm }
}

const saveListing = async () => {
  try {
    if (editingListing.value) {
      await api.put(`/listings/${editingListing.value.id}`, form.value)
    } else {
      await api.post('/listings', form.value)
    }
    closeModal()
    await fetchData()
  } catch (e) {
    console.error('Failed to save listing:', e)
  }
}

const deleteListing = async (id: number) => {
  if (!confirm('Är du säker på att du vill ta bort denna annons?')) return
  try {
    await api.delete(`/listings/${id}`)
    await fetchData()
  } catch (e) {
    console.error('Failed to delete listing:', e)
  }
}

onMounted(fetchData)
</script>
