<template>
  <div>
    <div class="flex justify-between items-center mb-6">
      <h1 class="page-header">Schemalagda körningar</h1>
      <button @click="showAddModal = true" class="btn btn-secondary">
        Lägg till cron-jobb
      </button>
    </div>

    <div class="card overflow-hidden" v-if="hasCronJobs">
      <table class="table">
        <thead>
          <tr>
            <th>Namn</th>
            <th>Cron-expression</th>
            <th>Söktermer</th>
            <th>Status</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="job in cronJobs" :key="job.id">
            <td class="font-medium text-slate-100">{{ job.name }}</td>
            <td class="font-mono text-sm text-slate-400">{{ job.cron_expression }}</td>
            <td class="text-sm text-slate-400">
              <span v-if="!job.search_term_ids || job.search_term_ids.length === 0">Alla</span>
              <span v-else>{{ job.search_term_ids.join(', ') }}</span>
            </td>
            <td>
              <button
                @click="toggleActive(job)"
                :class="job.is_active ? 'badge badge-success' : 'badge'"
                :disabled="isJobRunning(job.id)"
              >
                <span v-if="isJobRunning(job.id)" class="flex items-center gap-1">
                  <span class="w-2 h-2 bg-yellow-400 rounded-full animate-pulse"></span>
                  Körs...
                </span>
                <span v-else>{{ job.is_active ? "Aktiv" : "Inaktiv" }}</span>
              </button>
            </td>
            <td>
              <div class="flex gap-2">
                <button
                  v-if="isJobRunning(job.id)"
                  @click="cancelJob(job.id)"
                  class="text-yellow-400 hover:text-yellow-300 text-sm"
                >
                  Avbryt
                </button>
                <button
                  v-else
                  @click="editJob(job)"
                  class="text-blue-400 hover:text-blue-300 text-sm"
                >
                  Redigera
                </button>
                <button
                  @click="deleteJob(job.id)"
                  class="text-red-400 hover:text-red-300 text-sm"
                >
                  Ta bort
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="!hasCronJobs && !loading" class="text-center py-12 text-slate-500">
      Inga schemalagda körningar konfigurerade. Skapa ditt första cron-jobb!
    </div>

    <div v-if="loading" class="text-center py-12 text-slate-500">
      Laddar...
    </div>

    <div
      v-if="showAddModal || editingJob"
      class="fixed inset-0 bg-black/70 flex items-center justify-center z-50"
    >
      <div
        class="bg-slate-800 rounded-lg p-6 w-full max-w-lg border border-slate-700"
      >
        <h2 class="text-xl font-bold text-slate-100 mb-4">
          {{ editingJob ? 'Redigera cron-jobb' : 'Lägg till cron-jobb' }}
        </h2>

        <form @submit.prevent="saveJob" class="space-y-4">
          <div>
            <label class="label">Namn</label>
            <input
              v-model="form.name"
              type="text"
              class="input"
              placeholder="t.ex., Daglig iPhone-skanning"
              required
            />
          </div>
          <div>
            <label class="label">Cron-expression</label>
            <input
              v-model="form.cron_expression"
              type="text"
              class="input font-mono"
              placeholder="0 8 * * *"
              required
            />
            <p class="text-xs text-slate-500 mt-1">
              Format: minut timme dag-i-månad månad veckodag
            </p>
          </div>
          <div>
            <label class="label">Söktermer (valfritt)</label>
            <input
              v-model="form.search_term_ids_text"
              type="text"
              class="input"
              placeholder="1,2,3 (tomt = alla)"
            />
            <p class="text-xs text-slate-500 mt-1">
              Kommaseparerade IDn, eller tomt för alla söktermer
            </p>
          </div>
          <div>
            <label class="flex items-center gap-2 cursor-pointer">
              <input
                v-model="form.is_active"
                type="checkbox"
                class="w-4 h-4 rounded bg-slate-700 border-slate-600"
              />
              <span class="text-sm text-slate-300">Aktiv</span>
            </label>
          </div>

          <div class="flex justify-end gap-2 pt-4">
            <button
              type="button"
              @click="closeModal"
              class="btn btn-secondary"
            >
              Avbryt
            </button>
            <button type="submit" class="btn btn-primary">
              {{ editingJob ? 'Spara' : 'Lägg till' }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useApi } from "~/composables/useApi";

interface CronJob {
  id: number;
  name: string;
  cron_expression: string;
  search_term_ids: number[];
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

const api = useApi();

const cronJobs = ref<CronJob[]>([]);
const hasCronJobs = computed(() => cronJobs.value.length > 0);
const loading = ref(false);
const showAddModal = ref(false);
const editingJob = ref<CronJob | null>(null);
const runningJobIds = ref<number[]>([]);

let statusInterval: ReturnType<typeof setInterval> | null = null;

const isJobRunning = (id: number) => {
  return runningJobIds.value.includes(id);
};

const fetchStatus = async () => {
  try {
    const status = await api.get<{ running_jobs: number[] }>('/cron-jobs/status');
    runningJobIds.value = status.running_jobs || [];
  } catch (e) {
    console.error("Failed to fetch status:", e);
  }
};

const form = ref({
  name: "",
  cron_expression: "",
  search_term_ids_text: "",
  is_active: true,
});

const fetchData = async () => {
  loading.value = true;
  try {
    const jobs = await api.get<CronJob[]>('/cron-jobs');
    cronJobs.value = [...jobs];
  } catch (e) {
    console.error("Failed to fetch cron jobs:", e);
  } finally {
    loading.value = false;
  }
};

const parseSearchTermIds = (text: string): number[] => {
  if (!text.trim()) return [];
  return text.split(',').map(s => parseInt(s.trim(), 10)).filter(n => !isNaN(n));
};

const saveJob = async () => {
  try {
    const payload = {
      name: form.value.name,
      cron_expression: form.value.cron_expression,
      search_term_ids: parseSearchTermIds(form.value.search_term_ids_text),
      is_active: form.value.is_active,
    };

    if (editingJob.value) {
      await api.put(`/cron-jobs/${editingJob.value.id}`, payload);
    } else {
      await api.post('/cron-jobs', payload);
    }

    closeModal();
    await fetchData();
    await fetchStatus();
  } catch (e) {
    console.error("Failed to save cron job:", e);
    alert("Kunde inte spara cron-jobb: " + (e as Error).message);
  }
};

const editJob = (job: CronJob) => {
  editingJob.value = job;
  form.value = {
    name: job.name,
    cron_expression: job.cron_expression,
    search_term_ids_text: job.search_term_ids?.join(', ') || '',
    is_active: job.is_active,
  };
};

const deleteJob = async (id: number) => {
  if (!confirm("Ta bort detta cron-jobb?")) return;
  try {
    await api.delete(`/cron-jobs/${id}`);
    await fetchData();
  } catch (e) {
    console.error("Failed to delete cron job:", e);
  }
};

const cancelJob = async (id: number) => {
  try {
    await api.post('/cron-jobs/cancel', { job_id: id });
    await fetchStatus();
  } catch (e) {
    console.error("Failed to cancel cron job:", e);
  }
};

const toggleActive = async (job: CronJob) => {
  try {
    await api.put(`/cron-jobs/${job.id}`, { is_active: !job.is_active });
    await fetchData();
  } catch (e) {
    console.error("Failed to toggle active:", e);
  }
};

const closeModal = () => {
  showAddModal.value = false;
  editingJob.value = null;
  form.value = {
    name: "",
    cron_expression: "",
    search_term_ids_text: "",
    is_active: true,
  };
};

onMounted(() => {
  fetchData();
  fetchStatus();
  statusInterval = setInterval(fetchStatus, 5000);
});

onUnmounted(() => {
  if (statusInterval) {
    clearInterval(statusInterval);
  }
});
</script>
