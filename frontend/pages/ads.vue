<script setup lang="ts">
import type { ListingWithDetails } from "~/types/database";
import { createVimNavigation } from "~/composables/useVimNavigation";

const api = useApi();

const listings = ref<ListingWithDetails[]>([]);
const potentialListings = ref<ListingWithDetails[]>([]);
const activeTab = ref<'all' | 'good-value'>('all');
const error = ref<Error | null>(null);

const filteredListings = computed(() => {
  if (activeTab.value === 'good-value') {
    return potentialListings.value;
  }
  return listings.value;
});

const vimNav = createVimNavigation(0);

const isVimNavigationFocused = ref(false);

const fetchListings = async () => {
  try {
    const data = await api.get<ListingWithDetails[]>("/listings");

    if (!data || !Array.isArray(data)) {
      throw new Error("Invalid response from API");
    }

    listings.value = data.filter(
      (item) => item.Listing && !item.Listing.is_my_listing
    );
    vimNav.setItemCount(filteredListings.value.length);
  } catch (e: any) {
    console.error("Failed to fetch listings:", e);
    error.value = new Error(e.message || "Kunde inte hämta annonser");
  }
};

const fetchPotentialListings = async () => {
  try {
    const data = await api.get<ListingWithDetails[]>("/listings?good-value=true");

    if (!data || !Array.isArray(data)) {
      throw new Error("Invalid response from API");
    }

    potentialListings.value = data.filter(
      (item) => item.Listing && !item.Listing.is_my_listing
    );
    vimNav.setItemCount(filteredListings.value.length);
  } catch (e: any) {
    console.error("Failed to fetch potential listings:", e);
  }
};

await Promise.all([fetchListings(), fetchPotentialListings()]);

watch(filteredListings, (newList) => {
  vimNav.setItemCount(newList.length);
});

watch(activeTab, () => {
  vimNav.setItemCount(filteredListings.value.length);
  vimNav.clearSelection();
});

const selectedIndex = computed(() => vimNav.selectedIndex.value);

const selectedListing = computed(() => {
  if (selectedIndex.value === null || selectedIndex.value === -1) return null;
  return filteredListings.value[selectedIndex.value];
});

const removeListingFromState = (listingId: number) => {
  listings.value = listings.value.filter(l => l.Listing?.id !== listingId);
  potentialListings.value = potentialListings.value.filter(l => l.Listing?.id !== listingId);
};

const clearSelection = () => {
  vimNav.clearSelection();
  isVimNavigationFocused.value = false;
  vimNav.setFocused(false);
};

const deleteListing = async (listingId: number) => {
  try {
    await api.delete(`/listings/${listingId}`);
    removeListingFromState(listingId);
    clearSelection();
  } catch (e: any) {
    console.error("Failed to delete listing:", e);
    error.value = new Error(e.message || "Kunde inte ta bort annons");
  }
};

watch(selectedIndex, (index) => {
  if (index === null) return;
  setTimeout(() => {
    const elements = document.querySelectorAll('.card');
    const selectedEl = elements[index] as HTMLElement;
    if (selectedEl) {
      selectedEl.scrollIntoView({ behavior: 'smooth', block: 'center' });
    }
  }, 0);
});

onMounted(() => {
  window.addEventListener('keydown', handleKeydown);
});

onUnmounted(() => {
  window.removeEventListener('keydown', handleKeydown);
});

const handleKeydown = (e: KeyboardEvent) => {
  if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) {
    return;
  }

  if (e.key === 'j' || e.key === 'k') {
    if (!isVimNavigationFocused.value) {
      isVimNavigationFocused.value = true;
      vimNav.setFocused(true);
    }
  }

  if (e.key === 'Escape') {
    if (isVimNavigationFocused.value) {
      e.preventDefault();
      vimNav.clearSelection();
      isVimNavigationFocused.value = false;
      vimNav.setFocused(false);
    }
    return;
  }

  if (!isVimNavigationFocused.value) return;

  if (e.key === 'j') {
    e.preventDefault();
    vimNav.moveDown();
  } else if (e.key === 'k') {
    e.preventDefault();
    vimNav.moveUp();
  } else if (e.key === 'd') {
    e.preventDefault();
    const listing = selectedListing.value;
    if (!listing?.Listing?.id) return;
    
    if (confirm(`Ta bort "${listing.Listing.title || 'denna annons'}"?`)) {
      deleteListing(listing.Listing.id);
    }
  }
};

const isSelected = (index: number) => selectedIndex.value === index;

const formatPriceAsSEK = (price: number | null) => {
  if (!price) return "-";
  return `${price.toLocaleString("sv-SE")} kr`;
};

const formatValuationAsSEK = (sek: number | null | undefined) => {
  if (!sek) return "-";
  return `${sek.toLocaleString("sv-SE")} kr`;
};

const statusClass = (status: string) => {
  const classes: Record<string, string> = {
    draft: "badge badge-warning",
    active: "badge badge-success",
    sold: "badge badge-info",
    archived: "badge",
  };
  return classes[status] || classes.draft;
};

const marketplaceName = (id: number | null) => {
  if (!id) return "Unknown";
  return "Blocket";
};

const errorMessage = computed(() => {
  return error.value?.message || null;
});
</script>

<template>
  <div>
    <div class="flex justify-between items-center mb-6">
      <h1 class="page-header">Hittade annonser</h1>
    </div>

    <div class="flex gap-2 mb-6">
      <button
        @click="activeTab = 'all'"
        class="px-4 py-2 rounded-lg transition-colors"
        :class="activeTab === 'all' 
          ? 'bg-primary-600 text-white' 
          : 'bg-slate-700 text-slate-300 hover:bg-slate-600'"
      >
        Alla
      </button>
      <button
        @click="activeTab = 'good-value'"
        class="px-4 py-2 rounded-lg transition-colors"
        :class="activeTab === 'good-value' 
          ? 'bg-primary-600 text-white' 
          : 'bg-slate-700 text-slate-300 hover:bg-slate-600'"
      >
        Prisvärda
      </button>
    </div>

    <div v-if="errorMessage" class="text-center py-12 text-red-400">
      {{ errorMessage }}
    </div>

    <template v-else>
      <div class="grid grid-cols-1 gap-4">
        <div
          v-for="(item, index) in filteredListings"
          :key="item.Listing?.id"
          class="card overflow-hidden transition-all"
          :class="{ 'ring-2 ring-primary-500 ring-offset-2 ring-offset-slate-800': isSelected(index) }"
        >
          <div class="p-4">
            <div
              v-if="item.Listing"
              class="flex justify-between items-start mb-2"
            >
              <div>
                <p v-if="item.Listing.title" class="font-medium text-slate-100">
                  {{ item.Listing.title }}
                </p>
                <p v-else-if="item.Product" class="font-medium text-slate-100">
                  {{ item.Product.brand }} - {{ item.Product.name }}
                </p>
                <p v-else class="text-slate-500">Okänd produkt</p>
                <p class="text-sm text-slate-400">
                  Produkt:
                  <template v-if="item.Product">
                    {{ item.Product.brand }} - {{ item.Product.name }}
                  </template>
                  <template v-else>
                    Okänd produkt
                  </template>
                </p>
                <p class="text-lg font-bold text-primary-500">
                  {{
                    item.Listing.price
                      ? formatPriceAsSEK(item.Listing.price)
                      : "-"
                  }}
                </p>
                <p class="text-sm text-slate-400">
                  Nypris: {{ formatValuationAsSEK(item.Valuations?.find(v => v.valuation_type_id === 4)?.valuation) }}
                </p>
                <p class="text-sm text-slate-400">
                  Frakt:
                  {{
                    item.Listing.shipping_cost !== null &&
                    item.Listing.shipping_cost !== undefined
                      ? formatPriceAsSEK(item.Listing.shipping_cost)
                      : "Okänt"
                  }}
                </p>
                <p class="text-sm text-slate-400">
                  Värdering:
                  {{
                    item.ComputedValuation
                      ? formatValuationAsSEK(item.ComputedValuation)
                      : "-"
                  }}
                </p>
                <div v-if="item.Valuations && item.Valuations.filter(v => v.valuation_type_id !== 4).length > 0" class="mt-1">
                  <p class="text-xs text-slate-500 mb-1">Delvärderingar:</p>
                  <div class="flex flex-wrap gap-2">
                    <span
                      v-for="v in item.Valuations.filter(v => v.valuation_type_id !== 4)"
                      :key="v.id"
                      class="text-xs bg-slate-700 px-2 py-1 rounded"
                    >
                      {{ formatValuationAsSEK(v.valuation) }} - {{ v.valuation_type }}
                    </span>
                  </div>
                </div>
                <p
                  v-if="item.PotentialProfit !== undefined"
                  class="text-sm font-medium"
                  :class="item.PotentialProfit > 0 ? 'text-emerald-400' : 'text-red-400'"
                >
                  Vinst: {{ formatPriceAsSEK(item.PotentialProfit) }}
                  <span v-if="item.DiscountPercent !== undefined" class="text-slate-400 ml-2">
                    ({{ item.DiscountPercent.toFixed(1) }}% rabatt)
                  </span>
                </p>
                <p
                  v-else
                  class="text-sm text-slate-500"
                >
                  Ingen värdering tillgänglig
                </p>
              </div>
              <span :class="statusClass(item.Listing.status)">
                {{ item.Listing.status }}
              </span>
            </div>

            <p
              v-if="item.Listing?.description"
              class="text-sm text-slate-400 mb-2"
            >
              {{ item.Listing.description?.substring(0, 100) }}...
            </p>

            <div
              v-if="item.Listing"
              class="flex justify-between items-center text-sm text-slate-400"
            >
              <span>{{ marketplaceName(item.Listing.marketplace_id) }}</span>
              <a
                :href="item.Listing.link"
                target="_blank"
                class="text-primary-400 hover:text-primary-300"
              >
                Visa
              </a>
            </div>
          </div>
        </div>
      </div>

      <div
        v-if="filteredListings?.length === 0"
        class="text-center py-12 text-slate-500"
      >
        {{ activeTab === 'good-value' ? 'Inga prisvärda annonser hittades.' : 'Inga annonser hittades.' }}
      </div>
    </template>
  </div>
</template>
