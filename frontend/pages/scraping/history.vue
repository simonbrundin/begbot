<template>
  <div>
    <div class="flex justify-between items-center mb-6">
      <h1 class="page-header">Scrapinghistorik</h1>
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

    <div v-else-if="runs.length === 0" class="text-center py-12 text-slate-500">
      Ingen scrapinghistorik finns ännu. Starta en sökning för att se historik här.
    </div>

    <div v-else>
      <div class="card overflow-hidden">
        <table class="table">
          <thead>
            <tr>
              <th>Startad</th>
              <th>Status</th>
              <th>Annonser hittade</th>
              <th>Sparade annonser</th>
              <th>Bra köp</th>
              <th>Felmeddelande</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="run in runs" :key="run.id">
              <td class="text-sm text-slate-400">
                {{ formatDate(run.started_at) }}
              </td>
              <td>
                <span :class="statusClass(run.status)">
                  {{ statusText(run.status) }}
                </span>
              </td>
              <td>{{ run.total_ads_found }}</td>
              <td>
                <span :class="run.total_listings_saved > 0 ? 'text-emerald-400' : 'text-slate-400'">
                  {{ run.total_listings_saved }}
                </span>
              </td>
              <td>
                <span :class="run.total_good_buys > 0 ? 'text-emerald-400' : 'text-slate-400'">
                  {{ run.total_good_buys }}
                </span>
              </td>
              <td class="text-sm text-red-400">
                {{ run.error_message || '-' }}
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
import type { ScrapingRun } from "~/types/database";

interface PaginatedResponse {
  data: ScrapingRun[];
  total_count: number;
  page: number;
  page_size: number;
  total_pages: number;
}

const api = useApi();

const runs = ref<ScrapingRun[]>([]);
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
    const response = await api.get<PaginatedResponse>(`/scraping-runs?page=${currentPage.value}&page_size=${pageSize.value}`);
    runs.value = response.data;
    totalCount.value = response.total_count;
    totalPages.value = response.total_pages;
  } catch (e) {
    console.error("Failed to fetch scraping runs:", e);
    error.value = "Kunde inte ladda scrapinghistorik";
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

const statusText = (status: string) => {
  switch (status) {
    case 'running':
      return 'Pågående';
    case 'completed':
      return 'Slutförd';
    case 'failed':
      return 'Misslyckad';
    case 'cancelled':
      return 'Avbruten';
    default:
      return status;
  }
};

const statusClass = (status: string) => {
  switch (status) {
    case 'running':
      return 'text-blue-400';
    case 'completed':
      return 'text-emerald-400';
    case 'failed':
      return 'text-red-400';
    case 'cancelled':
      return 'text-yellow-400';
    default:
      return 'text-slate-400';
  }
};

onMounted(fetchHistory);
</script>
