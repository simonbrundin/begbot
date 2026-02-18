import { defineComponent, ref, computed, unref, useSSRContext } from 'vue';
import { ssrRenderAttrs, ssrInterpolate, ssrRenderList, ssrRenderClass, ssrRenderStyle } from 'vue/server-renderer';
import { T as TRADE_STATUSES } from './database-D1vXHN9-.mjs';
import { u as useApi } from './useApi-C1L14LOp.mjs';
import './server.mjs';
import '../nitro/nitro.mjs';
import 'node:http';
import 'node:https';
import 'node:events';
import 'node:buffer';
import 'node:fs';
import 'node:path';
import 'node:crypto';
import 'node:url';
import '../routes/renderer.mjs';
import 'vue-bundle-renderer/runtime';
import 'unhead/server';
import 'devalue';
import 'unhead/utils';
import 'pinia';
import 'vue-router';
import '@supabase/ssr';
import './loading--Nv5gQJt.mjs';

const _sfc_main = /* @__PURE__ */ defineComponent({
  __name: "analytics",
  __ssrInlineRender: true,
  setup(__props) {
    useApi();
    const items = ref([]);
    ref(false);
    const soldItems = computed(() => items.value.filter((i) => i.status_id === 5));
    const inStockItems = computed(() => items.value.filter((i) => i.status_id === 3));
    const totalProfit = computed(
      () => soldItems.value.reduce((sum, item) => {
        const sellTotal = (item.sell_price || 0) + (item.sell_shipping_collected || 0);
        const buyTotal = item.buy_price + item.buy_shipping_cost;
        const sellCost = (item.sell_packaging_cost || 0) + (item.sell_postage_cost || 0);
        return sum + (sellTotal - (buyTotal + sellCost));
      }, 0)
    );
    const totalRevenue = computed(
      () => soldItems.value.reduce((sum, item) => sum + (item.sell_price || 0) + (item.sell_shipping_collected || 0), 0)
    );
    const totalCOGS = computed(
      () => soldItems.value.reduce((sum, item) => sum + item.buy_price + item.buy_shipping_cost, 0)
    );
    const totalShipping = computed(
      () => soldItems.value.reduce((sum, item) => sum + (item.sell_packaging_cost || 0) + (item.sell_postage_cost || 0), 0)
    );
    const inventoryValue = computed(
      () => inStockItems.value.reduce((sum, item) => sum + item.buy_price + item.buy_shipping_cost, 0)
    );
    const statusCounts = computed(() => {
      const counts = {};
      items.value.forEach((item) => {
        counts[item.status_id] = (counts[item.status_id] || 0) + 1;
      });
      return counts;
    });
    const formatCurrency = (cents) => `${(cents / 100).toFixed(2)} kr`;
    const formatDate = (dateStr) => new Date(dateStr).toLocaleDateString("sv-SE");
    const calculateProfit = (item) => {
      const sellTotal = (item.sell_price || 0) + (item.sell_shipping_collected || 0);
      const buyTotal = item.buy_price + item.buy_shipping_cost;
      const sellCost = (item.sell_packaging_cost || 0) + (item.sell_postage_cost || 0);
      const profit = sellTotal - (buyTotal + sellCost);
      return `+${formatCurrency(profit)}`;
    };
    const statusColor = (statusId) => {
      const id = typeof statusId === "string" ? parseInt(statusId) : statusId;
      const colors = {
        1: "bg-slate-500",
        2: "bg-sky-500",
        3: "bg-amber-500",
        4: "bg-purple-500",
        5: "bg-emerald-500"
      };
      return colors[id] || "bg-slate-500";
    };
    return (_ctx, _push, _parent, _attrs) => {
      _push(`<div${ssrRenderAttrs(_attrs)}><h1 class="page-header">Analyser</h1><div class="grid grid-cols-4 gap-4 mb-6"><div class="stat-card"><p class="stat-label">Total vinst</p><p class="stat-value text-emerald-400">${ssrInterpolate(formatCurrency(unref(totalProfit)))}</p></div><div class="stat-card"><p class="stat-label">Lager v\xE4rde</p><p class="stat-value">${ssrInterpolate(formatCurrency(unref(inventoryValue)))}</p></div><div class="stat-card"><p class="stat-label">S\xE5lda artiklar</p><p class="stat-value">${ssrInterpolate(unref(soldItems).length)}</p></div><div class="stat-card"><p class="stat-label">Artiklar i lager</p><p class="stat-value">${ssrInterpolate(unref(inStockItems).length)}</p></div></div><div class="grid grid-cols-2 gap-6"><div class="card p-6"><h2 class="section-title">Artiklar per status</h2><div class="space-y-3"><!--[-->`);
      ssrRenderList(unref(statusCounts), (count, statusId) => {
        _push(`<div class="flex items-center"><span class="w-24 text-sm text-slate-400">${ssrInterpolate(unref(TRADE_STATUSES)[parseInt(statusId)])}</span><div class="flex-1 bg-slate-700 rounded-full h-4 mx-2"><div class="${ssrRenderClass([statusColor(statusId), "h-full rounded-full transition-all"])}" style="${ssrRenderStyle({ width: `${count / unref(items).length * 100}%` })}"></div></div><span class="w-8 text-sm font-medium">${ssrInterpolate(count)}</span></div>`);
      });
      _push(`<!--]--></div></div><div class="card p-6"><h2 class="section-title">Vinstf\xF6rdelning (s\xE5lda artiklar)</h2><div class="space-y-3"><div class="flex justify-between"><span class="text-slate-400">Total int\xE4kt</span><span class="font-medium text-slate-100">${ssrInterpolate(formatCurrency(unref(totalRevenue)))}</span></div><div class="flex justify-between"><span class="text-slate-400">Kostnad f\xF6r s\xE5lda varor</span><span class="font-medium text-red-400">-${ssrInterpolate(formatCurrency(unref(totalCOGS)))}</span></div><div class="flex justify-between"><span class="text-slate-400">Fraktkostnader</span><span class="font-medium text-red-400">-${ssrInterpolate(formatCurrency(unref(totalShipping)))}</span></div><div class="border-t border-slate-700 pt-2 flex justify-between"><span class="font-bold text-slate-100">Nettoresultat</span><span class="font-bold text-emerald-400">${ssrInterpolate(formatCurrency(unref(totalProfit)))}</span></div></div></div></div><div class="card p-6 mt-6"><h2 class="section-title">Senaste f\xF6rs\xE4ljningar</h2><table class="table"><thead><tr><th>Produkt</th><th>K\xF6pris</th><th>F\xF6rs\xE4ljningspris</th><th>Vinst</th><th>S\xE5ld datum</th></tr></thead><tbody><!--[-->`);
      ssrRenderList(unref(soldItems).slice(0, 10), (item) => {
        _push(`<tr><td>`);
        if (item.product) {
          _push(`<span class="text-slate-100">${ssrInterpolate(item.product.brand)} ${ssrInterpolate(item.product.name)}</span>`);
        } else {
          _push(`<span class="text-slate-500">Ok\xE4nd</span>`);
        }
        _push(`</td><td>${ssrInterpolate(formatCurrency(item.buy_price + item.buy_shipping_cost))}</td><td>${ssrInterpolate(item.sell_price ? formatCurrency(item.sell_price + (item.sell_shipping_collected || 0)) : "-")}</td><td class="text-emerald-400 font-medium">${ssrInterpolate(calculateProfit(item))}</td><td class="text-sm text-slate-400">${ssrInterpolate(item.sell_date ? formatDate(item.sell_date) : "-")}</td></tr>`);
      });
      _push(`<!--]--></tbody></table></div></div>`);
    };
  }
});
const _sfc_setup = _sfc_main.setup;
_sfc_main.setup = (props, ctx) => {
  const ssrContext = useSSRContext();
  (ssrContext.modules || (ssrContext.modules = /* @__PURE__ */ new Set())).add("pages/analytics.vue");
  return _sfc_setup ? _sfc_setup(props, ctx) : void 0;
};

export { _sfc_main as default };
//# sourceMappingURL=analytics-yESQwxzS.mjs.map
