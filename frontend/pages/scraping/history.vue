<template>
  <div>
    <div class="flex justify-between items-center mb-6">
      <h1 class="page-header">Sökhistorik</h1>
      <div class="flex gap-2">
        <NuxtLink to="/scraping/terms" class="btn btn-secondary">
          Sökord
        </NuxtLink>
        <button @click="refreshHistory" class="btn btn-secondary">
          Uppdatera
        </button>
      </div>
    </div>

    <div v-if="loading" class="text-center py-12 text-slate-500">
      Laddar...
    </div>

    <div v-else-if="error" class="card p-4 text-red-400">
      {{ error }}
    </div>

    <div v-else-if="history.length === 0" class="text-center py-12 text-slate-500">
      Ingen sökhistorik finns ännu. Starta en sökning för att se historik här.
    </div>

    <div v-else>
      <div class="card overflow-hidden">
        <table class="table">
          <thead>
            <tr>
              <th>Tidpunkt</th>
              <th>Sökord</th>
              <th>Marknadsplats</th>
              <th>Resultat</th>
              <th>Nya annonser</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in history" :key="item.id">
              <td class="text-sm text-slate-400">
                {{ formatDate(item.searched_at) }}
              </td>
              <td class="font-medium text-slate-100">{{ item.search_term_desc }}</td>
              <td>{{ item.marketplace_name || '-' }}</td>
              <td>{{ item.results_found }}</td>
              <td>
                <span :class="item.new_ads_found > 0 ? 'text-emerald-400' : 'text-slate-400'">
                  {{ item.new_ads_found }}
                </span>
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
  </div>
</template>

<script setup lang="ts">
import type { SearchHistory } from "~/types/database";

interface PaginatedResponse {
  data: SearchHistory[];
  total_count: number;
  page: number;
  page_size: number;
  total_pages: number;
}

const api = useApi();

const history = ref<SearchHistory[]>([]);
const loading = ref(false);
const error = ref<string | null>(null);
const currentPage = ref(1);
const pageSize = ref(20);
const totalCount = ref(0);
const totalPages = ref(0);

const fetchHistory = async () => {
  loading.value = true;
  error.value = null;
  try {
    const response = await api.get<PaginatedResponse>('/search-history', {
      page: currentPage.value,
      page_size: pageSize.value,
    });
    history.value = response.data;
    totalCount.value = response.total_count;
    totalPages.value = response.total_pages;
  } catch (e) {
    console.error("Failed to fetch history:", e);
    error.value = "Kunde inte ladda sökhistorik";
  } finally {
    loading.value = false;
  }
};

const changePage = (page: number) => {
  if (page < 1 || page > totalPages.value) return;
  currentPage.value = page;
  fetchHistory();
};

const refreshHistory = () => {
  fetchHistory();
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

onMounted(fetchHistory);
</script>
