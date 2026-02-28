<template>
  <div>
    <div class="flex justify-between items-center mb-6">
      <h1 class="page-header">Skickade mejl</h1>
      <div class="flex gap-2">
        <NuxtLink to="/scraping/history" class="btn btn-secondary">
          Tillbaka till scraping
        </NuxtLink>
        <button @click="fetchEmails" class="btn btn-secondary">
          Uppdatera
        </button>
      </div>
    </div>

    <div class="card p-4 mb-6">
      <div class="flex flex-wrap gap-4 items-end">
        <div>
          <label class="block text-sm text-slate-400 mb-1">Från datum</label>
          <input
            v-model="filters.fromDate"
            type="date"
            class="input"
            @change="applyFilters"
          />
        </div>
        <div>
          <label class="block text-sm text-slate-400 mb-1">Till datum</label>
          <input
            v-model="filters.toDate"
            type="date"
            class="input"
            @change="applyFilters"
          />
        </div>
        <div>
          <label class="block text-sm text-slate-400 mb-1">Produkt</label>
          <select v-model="filters.productId" class="input" @change="applyFilters">
            <option :value="null">Alla produkter</option>
            <option v-for="product in products" :key="product.id" :value="product.id">
              {{ product.name || product.brand || 'Produkt ' + product.id }}
            </option>
          </select>
        </div>
        <button v-if="hasFilters" @click="clearFilters" class="btn btn-secondary">
          Rensa filter
        </button>
      </div>
    </div>

    <div v-if="loading" class="text-center py-12 text-slate-500">
      Laddar...
    </div>

    <div v-else-if="error" class="card p-4 text-red-400">
      {{ error }}
    </div>

    <div v-else-if="emails.length === 0" class="text-center py-12 text-slate-500">
      Inga skickade mejl hittades.
    </div>

    <div v-else>
      <div class="card overflow-hidden">
        <table class="table">
          <thead>
            <tr>
              <th>Tid</th>
              <th>Annons</th>
              <th>Pris</th>
              <th>Värdering</th>
              <th>Säkerhet</th>
              <th>Profit</th>
              <th>Rabatt</th>
              <th>Produkt</th>
            </tr>
          </thead>
          <tbody>
            <tr 
              v-for="email in emails" 
              :key="email.id"
              class="cursor-pointer hover:bg-slate-800/50"
              @click="openDetails(email)"
            >
              <td class="text-sm text-slate-400">
                {{ formatDate(email.sent_at) }}
              </td>
              <td>
                <a 
                  :href="email.listing_link" 
                  target="_blank" 
                  class="text-blue-400 hover:underline"
                  @click.stop
                >
                  {{ truncate(email.listing_title, 40) }}
                </a>
              </td>
              <td>{{ email.listing_price ? email.listing_price + ' kr' : '-' }}</td>
              <td>{{ email.listing_valuation ? email.listing_valuation + ' kr' : '-' }}</td>
              <td>
                <span v-if="email.confidence" :class="email.confidence > 70 ? 'text-emerald-400' : email.confidence > 40 ? 'text-yellow-400' : 'text-slate-400'">
                  {{ email.confidence.toFixed(1) }}%
                </span>
                <span v-else class="text-slate-400">-</span>
              </td>
              <td>
                <span :class="email.profit > 0 ? 'text-emerald-400' : 'text-slate-400'">
                  {{ email.profit }} kr
                </span>
              </td>
              <td>
                <span :class="email.discount_percent > 30 ? 'text-emerald-400' : email.discount_percent > 20 ? 'text-yellow-400' : 'text-slate-400'">
                  {{ email.discount_percent.toFixed(0) }}%
                </span>
              </td>
              <td class="text-sm">
                {{ email.product_name || email.brand || '-' }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <div v-if="totalPages > 1" class="flex justify-center items-center gap-2 mt-6">
        <button
          @click="changePage(currentPage - 1)"
          :disabled="currentPage <= 1"
          class="btn btn-secondary"
        >
          Föregående
        </button>
        <span class="text-slate-400">
          Sida {{ currentPage }} av {{ totalPages }}
        </span>
        <button
          @click="changePage(currentPage + 1)"
          :disabled="currentPage >= totalPages"
          class="btn btn-secondary"
        >
          Nästa
        </button>
      </div>
    </div>

    <div v-if="selectedEmail" class="fixed inset-0 bg-black/50 flex items-center justify-center p-4 z-50" @click.self="closeDetails">
      <div class="card max-w-2xl w-full max-h-[90vh] overflow-y-auto">
        <div class="flex justify-between items-start mb-4">
          <h2 class="text-xl font-semibold">Mejlad annons</h2>
          <button @click="closeDetails" class="text-slate-400 hover:text-white">
            <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        
        <div class="space-y-4">
          <div>
            <h3 class="text-lg font-medium">{{ selectedEmail.listing_title }}</h3>
            <a :href="selectedEmail.listing_link" target="_blank" class="text-blue-400 hover:underline text-sm">
              Länk till annons →
            </a>
          </div>

          <div class="grid grid-cols-2 gap-4">
            <div class="bg-slate-800 rounded-lg p-3">
              <div class="text-sm text-slate-400">Pris</div>
              <div class="text-lg font-semibold">{{ selectedEmail.listing_price }} kr</div>
            </div>
            <div class="bg-slate-800 rounded-lg p-3">
              <div class="text-sm text-slate-400">Värdering</div>
              <div class="text-lg font-semibold">{{ selectedEmail.listing_valuation }} kr</div>
            </div>
            <div class="bg-slate-800 rounded-lg p-3">
              <div class="text-sm text-slate-400">Profit</div>
              <div class="text-lg font-semibold" :class="selectedEmail.profit > 0 ? 'text-emerald-400' : 'text-red-400'">
                {{ selectedEmail.profit }} kr
              </div>
            </div>
            <div class="bg-slate-800 rounded-lg p-3">
              <div class="text-sm text-slate-400">Rabatt</div>
              <div class="text-lg font-semibold" :class="selectedEmail.discount_percent > 30 ? 'text-emerald-400' : selectedEmail.discount_percent > 20 ? 'text-yellow-400' : 'text-red-400'">
                {{ selectedEmail.discount_percent.toFixed(1) }}%
              </div>
            </div>
          </div>

          <div v-if="selectedEmail.product_name || selectedEmail.brand" class="bg-slate-800 rounded-lg p-3">
            <div class="text-sm text-slate-400">Produkt</div>
            <div>{{ selectedEmail.product_name || selectedEmail.brand }}</div>
          </div>

          <div class="text-sm text-slate-400">
            Skickat: {{ formatDate(selectedEmail.sent_at) }}
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { SentEmail, Product } from "~/types/database";

interface PaginatedResponse {
  data: SentEmail[];
  total_count: number;
  page: number;
  page_size: number;
  total_pages: number;
}

const api = useApi();

const emails = ref<SentEmail[]>([]);
const products = ref<Product[]>([]);
const loading = ref(false);
const error = ref<string | null>(null);
const currentPage = ref(1);
const pageSize = ref(20);
const totalCount = ref(0);
const totalPages = ref(0);
const selectedEmail = ref<SentEmail | null>(null);

const filters = ref<{
  fromDate: string | null;
  toDate: string | null;
  productId: number | null;
}>({
  fromDate: null,
  toDate: null,
  productId: null,
});

const hasFilters = computed(() => {
  return filters.value.fromDate || filters.value.toDate || filters.value.productId;
});

const fetchProducts = async () => {
  try {
    const response = await api.get<Product[]>('/sent-emails/products');
    products.value = response;
  } catch (e) {
    console.error("Failed to fetch products:", e);
  }
};

const buildQueryParams = () => {
  const params = new URLSearchParams();
  params.set('page', currentPage.value.toString());
  params.set('page_size', pageSize.value.toString());
  
  if (filters.value.fromDate) {
    params.set('from_date', filters.value.fromDate);
  }
  if (filters.value.toDate) {
    params.set('to_date', filters.value.toDate);
  }
  if (filters.value.productId) {
    params.set('product_id', filters.value.productId.toString());
  }
  
  return params.toString();
};

const fetchEmails = async () => {
  loading.value = true;
  error.value = null;
  try {
    const response = await api.get<PaginatedResponse>(`/sent-emails?${buildQueryParams()}`);
    emails.value = response.data;
    totalCount.value = response.total_count;
    totalPages.value = response.total_pages;
  } catch (e) {
    console.error("Failed to fetch sent emails:", e);
    error.value = "Kunde inte ladda skickade mejl";
  } finally {
    loading.value = false;
  }
};

const applyFilters = () => {
  currentPage.value = 1;
  fetchEmails();
};

const clearFilters = () => {
  filters.value = {
    fromDate: null,
    toDate: null,
    productId: null,
  };
  currentPage.value = 1;
  fetchEmails();
};

const changePage = (page: number) => {
  if (page < 1 || page > totalPages.value) return;
  currentPage.value = page;
  fetchEmails();
};

const openDetails = (email: SentEmail) => {
  selectedEmail.value = email;
};

const closeDetails = () => {
  selectedEmail.value = null;
};

const formatDate = (dateStr: string) => {
  const date = new Date(dateStr);
  return date.toLocaleString('sv-SE', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });
};

const truncate = (str: string, length: number) => {
  if (str.length <= length) return str;
  return str.slice(0, length) + '...';
};

onMounted(() => {
  fetchProducts();
  fetchEmails();
});
</script>
