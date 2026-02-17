<template>
  <div>
    <div class="flex justify-between items-center mb-6">
      <h1 class="page-header">Sökord</h1>
      <div class="flex gap-2">
        <NuxtLink to="/scraping/history" class="btn btn-secondary">
          Historik
        </NuxtLink>
        <button
          @click="fetchAds"
          :disabled="isFetching"
          class="btn btn-primary"
        >
          {{ isFetching ? "Hämtar..." : "Hämta annonser" }}
        </button>
        <button @click="showAddModal = true" class="btn btn-secondary">
          Lägg till sökord
        </button>
      </div>
    </div>

    <div v-if="isFetching || fetchStatus" class="card p-4 mb-6">
      <div class="flex justify-between items-center mb-2">
        <span class="font-medium text-slate-100">{{ fetchStatusText }}</span>
        <div class="flex items-center gap-3">
          <span class="text-sm text-slate-400">{{ fetchProgress }}%</span>
          <button
            v-if="canCancelJob"
            @click="cancelJob"
            :disabled="isCancelling"
            class="btn btn-danger text-xs py-1 px-3"
          >
            {{ isCancelling ? 'Avbryter...' : 'Avbryt' }}
          </button>
        </div>
      </div>
      <div class="w-full bg-slate-700 rounded-full h-2.5">
        <div
          class="bg-primary-500 h-2.5 rounded-full transition-all duration-300"
          :style="{ width: fetchProgress + '%' }"
        ></div>
      </div>
      <p v-if="fetchStatus?.current_query" class="text-sm text-slate-400 mt-2">
        Bearbetar: {{ fetchStatus.current_query }}
      </p>
      <p
        v-if="fetchStatus?.ads_found !== undefined && fetchStatus.ads_found > 0"
        class="text-sm text-emerald-400 mt-2"
      >
        Hittade {{ fetchStatus.ads_found }} annonser
      </p>
      <p v-if="fetchStatus?.error" class="text-sm text-red-400 mt-2">
        Fel: {{ fetchStatus.error }}
      </p>
    </div>

    <ScraperLogConsole
      v-if="currentJobId"
      :logs="scraperLogs"
      :is-connected="isLogConnected"
      :error="logError"
      @clear="clearScraperLogs"
    />

    <div class="card overflow-hidden mt-6" v-if="hasSearchTerms">
      <table class="table">
        <thead>
          <tr>
            <th>Beskrivning</th>
            <th>URL</th>
            <th>Marknadsplats</th>
            <th>Status</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="term in searchTerms" :key="term.id">
            <td class="font-medium text-slate-100">{{ term.description }}</td>
            <td class="text-sm text-slate-400 truncate max-w-xs">
              {{ term.url }}
            </td>
            <td>{{ marketplaceName(term.marketplace_id) }}</td>
            <td>
              <button
                @click="toggleActive(term)"
                :class="term.is_active ? 'badge badge-success' : 'badge'"
              >
                {{ term.is_active ? "Aktiv" : "Inaktiv" }}
              </button>
            </td>
            <td>
              <button
                @click="deleteTerm(term.id)"
                class="text-red-400 hover:text-red-300 text-sm"
              >
                Ta bort
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div
      v-if="!hasSearchTerms"
      class="text-center py-12 text-slate-500"
    >
      Inga sökord konfigurerade. Lägg till ditt första för att börja skrapa!
    </div>

    <div
      v-if="showAddModal"
      class="fixed inset-0 bg-black/70 flex items-center justify-center z-50"
    >
      <div
        class="bg-slate-800 rounded-lg p-6 w-full max-w-lg border border-slate-700"
      >
        <h2 class="text-xl font-bold text-slate-100 mb-4">Lägg till sökord</h2>

        <form @submit.prevent="saveTerm" class="space-y-4">
          <div>
            <label class="label">Beskrivning</label>
            <input
              v-model="form.description"
              type="text"
              class="input"
              placeholder="t.ex., iPhone 15 Pro"
              required
            />
          </div>
          <div>
            <label class="label">Sök-URL</label>
            <input
              v-model="form.url"
              type="url"
              class="input"
              placeholder="https://www.blocket.se/..."
              required
            />
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

          <div class="flex justify-end gap-2 pt-4">
            <button
              type="button"
              @click="showAddModal = false"
              class="btn btn-secondary"
            >
              Avbryt
            </button>
            <button type="submit" class="btn btn-primary">Lägg till</button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { SearchTerm, Marketplace } from "~/types/database";
import { useScraperLogs } from "~/composables/useScraperLogs";

interface FetchStatus {
  status: string;
  progress: number;
  total_queries: number;
  completed_queries: number;
  current_query: string;
  ads_found: number;
  error: string;
}

const api = useApi();

const searchTerms = ref<SearchTerm[]>([]);
const hasSearchTerms = computed(() => searchTerms.value.length > 0);
const marketplaces = ref<Marketplace[]>([]);
const loading = ref(false);
const showAddModal = ref(false);

const isFetching = ref(false);
const isCancelling = ref(false);
const currentJobId = ref<string | null>(null);
const fetchStatus = ref<FetchStatus | null>(null);
let pollInterval: ReturnType<typeof setInterval> | null = null;

const { logs: scraperLogs, isConnected: isLogConnected, error: logError, clearLogs: clearScraperLogs } = useScraperLogs(currentJobId);

const form = ref({
  description: "",
  url: "",
  marketplace_id: null as number | null,
  is_active: true,
});

const fetchProgress = computed(() => {
  if (!fetchStatus.value) return 0;
  return fetchStatus.value.progress;
});

const fetchStatusText = computed(() => {
  if (!fetchStatus.value) return "Väntar";
  switch (fetchStatus.value.status) {
    case "pending":
      return "Startar...";
    case "running":
      return "Hämtar annonser...";
    case "completed":
      return "Klar!";
    case "failed":
      return "Misslyckades";
    case "cancelled":
      return "Avbrutet";
    default:
      return "Väntar";
  }
});

const canCancelJob = computed(() => {
  return fetchStatus.value &&
    (fetchStatus.value.status === "pending" || fetchStatus.value.status === "running");
});

const fetchAds = async () => {
  if (isFetching.value) return;

  isFetching.value = true;
  fetchStatus.value = null;

  try {
    const response = await api.post<{ job_id: string; status: string }>('/fetch-ads', {});
    currentJobId.value = response.job_id;

    pollInterval = setInterval(async () => {
      if (!currentJobId.value) {
        stopPolling();
        return;
      }

      try {
        const status = await api.get<FetchStatus>(`/fetch-ads/status/${currentJobId.value}`);
        fetchStatus.value = status;

        if (status && (status.status === "completed" || status.status === "failed" || status.status === "cancelled")) {
          stopPolling();
          isFetching.value = false;
          isCancelling.value = false;
          setTimeout(() => {
            currentJobId.value = null;
            fetchStatus.value = null;
          }, 5000);
          if (status.status === "completed") {
            await fetchData();
          }
        }
      } catch (e) {
        console.error("Failed to fetch status:", e);
        stopPolling();
        isFetching.value = false;
      }
    }, 1000);
  } catch (e) {
    console.error("Failed to start fetch:", e);
    isFetching.value = false;
    currentJobId.value = null;
    fetchStatus.value = {
      status: "failed",
      progress: 0,
      total_queries: 0,
      completed_queries: 0,
      current_query: "",
      ads_found: 0,
      error: "Kunde inte starta hämtning",
    };
  }
};

const stopPolling = () => {
  if (pollInterval) {
    clearInterval(pollInterval);
    pollInterval = null;
  }
};

const cancelJob = async () => {
  console.log("Cancel button clicked, jobId:", currentJobId.value);
  if (!currentJobId.value || isCancelling.value) {
    console.log("Cannot cancel - no jobId or already cancelling");
    return;
  }

  isCancelling.value = true;
  try {
    console.log("Calling cancel API...");
    await api.post(`/fetch-ads/cancel/${currentJobId.value}`, {});
    console.log("Cancel API called successfully");
  } catch (e) {
    console.error("Failed to cancel job:", e);
    isCancelling.value = false;
  }
};

const fetchData = async () => {
  loading.value = true;
  try {
    const [termsRes, marketsRes] = await Promise.all([
      api.get<SearchTerm[]>('/search-terms'),
      api.get<Marketplace[]>('/marketplaces'),
    ]);
    searchTerms.value = [...termsRes];
    marketplaces.value = [...marketsRes];
  } catch (e) {
    console.error("Failed to fetch data:", e);
  } finally {
    loading.value = false;
  }
};

const marketplaceName = (id: number | null) => {
  if (!id) return "Unknown";
  return marketplaces.value.find((m) => m.id === id)?.name || "Unknown";
};

const saveTerm = async () => {
  try {
    await api.post('/search-terms', form.value);
    showAddModal.value = false;
    form.value = {
      description: "",
      url: "",
      marketplace_id: null,
      is_active: true,
    };
    await fetchData();
  } catch (e) {
    console.error("Failed to save search term:", e);
  }
};

const toggleActive = async (term: SearchTerm) => {
  try {
    await api.put(`/search-terms/${term.id}`, { is_active: !term.is_active });
    await fetchData();
  } catch (e) {
    console.error("Failed to toggle active:", e);
  }
};

const deleteTerm = async (id: number) => {
  if (!confirm("Ta bort detta sökord?")) return;
  try {
    await api.delete(`/search-terms/${id}`);
    await fetchData();
  } catch (e) {
    console.error("Failed to delete term:", e);
  }
};

onMounted(fetchData);

onUnmounted(() => {
  stopPolling();
});
</script>
