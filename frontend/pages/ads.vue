<script setup lang="ts">
import type { ListingWithDetails } from "~/types/database";

const config = useRuntimeConfig();

const {
  data: listings,
  error,
  pending,
} = await useAsyncData("ads-listings", async () => {
  try {
    const response = await fetch(`${config.public.apiBase}/api/listings`);

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    const data = await response.json();

    if (!data || !Array.isArray(data)) {
      throw new Error("Invalid response from API");
    }

     return data.filter(
       (item: any) => item.Listing && !item.Listing.is_my_listing
     );
  } catch (e: any) {
    console.error("Failed to fetch listings:", e);
    throw new Error(e.message || "Kunde inte hämta annonser");
  }
});

const formatCurrency = (price: number | null) => {
  if (!price) return "-";
  return `${price.toLocaleString("sv-SE")} kr`;
};

const formatPriceAsSEK = (price: number | null) => {
  if (!price) return "-";
  return `${price.toLocaleString("sv-SE")} kr`;
};

const formatValuationAsSEK = (sek: number | null) => {
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

    <div v-if="errorMessage" class="text-center py-12 text-red-400">
      {{ errorMessage }}
    </div>

    <div v-else-if="pending" class="text-center py-12 text-slate-500">
      Laddar...
    </div>

    <template v-else>
      <div class="grid grid-cols-1 gap-4">
        <div
          v-for="item in listings"
          :key="item.Listing?.id"
          class="card overflow-hidden"
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
                      ? formatCurrency(item.Listing.shipping_cost)
                      : "Okänt"
                  }}
                </p>
                <p class="text-sm text-slate-400">
                  Värdering:
                  {{
                    item.Listing.valuation
                      ? formatValuationAsSEK(item.Listing.valuation)
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
        v-if="listings?.length === 0"
        class="text-center py-12 text-slate-500"
      >
        Inga annonser hittades.
      </div>
    </template>
  </div>
</template>
