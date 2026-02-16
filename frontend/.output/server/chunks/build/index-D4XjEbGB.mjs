import { defineComponent, ref, computed, unref, useSSRContext } from 'vue';
import { ssrRenderAttrs, ssrInterpolate, ssrIncludeBooleanAttr, ssrLooseContain, ssrLooseEqual, ssrRenderList, ssrRenderAttr, ssrRenderClass } from 'vue/server-renderer';
import { T as TRADE_STATUSES } from './database-D1vXHN9-.mjs';
import { u as useApi } from './useApi-EIa4-qJb.mjs';
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
import 'unhead/plugins';
import 'vue-router';
import '@supabase/ssr';
import './loading-qsg6mAmB.mjs';

const _sfc_main = /* @__PURE__ */ defineComponent({
  __name: "index",
  __ssrInlineRender: true,
  setup(__props) {
    useApi();
    const items = ref([]);
    const products = ref([]);
    ref(false);
    const showAddModal = ref(false);
    const editingItem = ref(null);
    const statusFilter = ref("");
    const searchQuery = ref("");
    const defaultForm = {
      product_id: null,
      status_id: 1,
      buy_price: 0,
      buy_shipping_cost: 0,
      sell_price: null,
      sell_packaging_cost: 0,
      sell_postage_cost: 0,
      sell_shipping_collected: 0,
      storage: null,
      source_link: "",
      buy_date: "",
      sell_date: ""
    };
    const itemForm = ref({ ...defaultForm });
    const inStockCount = computed(() => items.value.filter((i) => i.status_id === 3).length);
    const listedCount = computed(() => items.value.filter((i) => i.status_id === 4).length);
    const soldCount = computed(() => items.value.filter((i) => i.status_id === 5).length);
    const filteredItems = computed(() => {
      return items.value.filter((item) => {
        var _a, _b, _c, _d;
        if (statusFilter.value && item.status_id !== parseInt(statusFilter.value)) return false;
        if (searchQuery.value) {
          const query = searchQuery.value.toLowerCase();
          const productName = ((_b = (_a = item.product) == null ? void 0 : _a.name) == null ? void 0 : _b.toLowerCase()) || "";
          const brand = ((_d = (_c = item.product) == null ? void 0 : _c.brand) == null ? void 0 : _d.toLowerCase()) || "";
          if (!productName.includes(query) && !brand.includes(query)) return false;
        }
        return true;
      });
    });
    const formatCurrency = (cents) => {
      return `${(cents / 100).toFixed(2)} kr`;
    };
    const formatDate = (dateStr) => {
      return new Date(dateStr).toLocaleDateString("sv-SE");
    };
    const calculateProfit = (item) => {
      const sellTotal = (item.sell_price || 0) + (item.sell_shipping_collected || 0);
      const buyTotal = item.buy_price + item.buy_shipping_cost;
      const sellCost = (item.sell_packaging_cost || 0) + (item.sell_postage_cost || 0);
      const profit = sellTotal - (buyTotal + sellCost);
      return profit >= 0 ? `+${formatCurrency(profit)}` : formatCurrency(profit);
    };
    const profitClass = (item) => {
      const sellTotal = (item.sell_price || 0) + (item.sell_shipping_collected || 0);
      const buyTotal = item.buy_price + item.buy_shipping_cost;
      const sellCost = (item.sell_packaging_cost || 0) + (item.sell_postage_cost || 0);
      const profit = sellTotal - (buyTotal + sellCost);
      return profit >= 0 ? "text-emerald-400" : "text-red-400";
    };
    const statusBadgeClass = (statusId) => {
      const classes = {
        1: "badge badge-info",
        2: "badge badge-info",
        3: "badge badge-warning",
        4: "badge badge-info",
        5: "badge badge-success"
      };
      return classes[statusId] || classes[1];
    };
    return (_ctx, _push, _parent, _attrs) => {
      _push(`<div${ssrRenderAttrs(_attrs)}><div class="flex justify-between items-center mb-6"><h1 class="page-header">Lager</h1><button class="btn btn-primary"> L\xE4gg till </button></div><div class="grid grid-cols-4 gap-4 mb-6"><div class="stat-card"><p class="stat-label">Totalt antal</p><p class="stat-value">${ssrInterpolate(unref(items).length)}</p></div><div class="stat-card"><p class="stat-label">I lager</p><p class="stat-value">${ssrInterpolate(unref(inStockCount))}</p></div><div class="stat-card"><p class="stat-label">Utlagda</p><p class="stat-value">${ssrInterpolate(unref(listedCount))}</p></div><div class="stat-card"><p class="stat-label">S\xE5lda</p><p class="stat-value">${ssrInterpolate(unref(soldCount))}</p></div></div><div class="card p-4 mb-6"><div class="flex gap-4"><select class="input w-48"><option value=""${ssrIncludeBooleanAttr(Array.isArray(unref(statusFilter)) ? ssrLooseContain(unref(statusFilter), "") : ssrLooseEqual(unref(statusFilter), "")) ? " selected" : ""}>Alla statusar</option><!--[-->`);
      ssrRenderList(unref(TRADE_STATUSES), (name, id) => {
        _push(`<option${ssrRenderAttr("value", id)}${ssrIncludeBooleanAttr(Array.isArray(unref(statusFilter)) ? ssrLooseContain(unref(statusFilter), id) : ssrLooseEqual(unref(statusFilter), id)) ? " selected" : ""}>${ssrInterpolate(name)}</option>`);
      });
      _push(`<!--]--></select><input${ssrRenderAttr("value", unref(searchQuery))} type="text" class="input flex-1" placeholder="S\xF6k..."></div></div><div class="card overflow-hidden"><table class="table"><thead><tr><th>Produkt</th><th>Status</th><th>K\xF6pris</th><th>F\xF6rs\xE4ljningspris</th><th>Vinst</th><th>Datum</th><th></th></tr></thead><tbody><!--[-->`);
      ssrRenderList(unref(filteredItems), (item) => {
        _push(`<tr><td>`);
        if (item.product) {
          _push(`<div><p class="font-medium text-slate-100">${ssrInterpolate(item.product.brand)} ${ssrInterpolate(item.product.name)}</p><p class="text-sm text-slate-400">${ssrInterpolate(item.product.category)}</p></div>`);
        } else {
          _push(`<span class="text-slate-500">Ok\xE4nd</span>`);
        }
        _push(`</td><td><span class="${ssrRenderClass(statusBadgeClass(item.status_id))}">${ssrInterpolate(unref(TRADE_STATUSES)[item.status_id] || "ok\xE4nd")}</span></td><td>${ssrInterpolate(formatCurrency(item.buy_price))}</td><td>${ssrInterpolate(item.sell_price ? formatCurrency(item.sell_price) : "-")}</td><td class="${ssrRenderClass(profitClass(item))}">${ssrInterpolate(calculateProfit(item))}</td><td class="text-sm text-slate-400">${ssrInterpolate(formatDate(item.created_at))}</td><td><button class="text-primary-400 hover:text-primary-300"> Redigera </button></td></tr>`);
      });
      _push(`<!--]--></tbody></table></div>`);
      if (unref(showAddModal) || unref(editingItem)) {
        _push(`<div class="fixed inset-0 bg-black/70 flex items-center justify-center z-50"><div class="bg-slate-800 rounded-lg p-6 w-full max-w-2xl max-h-[90vh] overflow-y-auto border border-slate-700"><h2 class="text-xl font-bold text-slate-100 mb-4">${ssrInterpolate(unref(editingItem) ? "Redigera" : "L\xE4gg till ny")}</h2><form class="space-y-4"><div class="grid grid-cols-2 gap-4"><div><label class="label">Produkt</label><select class="input"><option value=""${ssrIncludeBooleanAttr(Array.isArray(unref(itemForm).product_id) ? ssrLooseContain(unref(itemForm).product_id, "") : ssrLooseEqual(unref(itemForm).product_id, "")) ? " selected" : ""}>V\xE4lj produkt...</option><!--[-->`);
        ssrRenderList(unref(products), (p) => {
          _push(`<option${ssrRenderAttr("value", p.id)}${ssrIncludeBooleanAttr(Array.isArray(unref(itemForm).product_id) ? ssrLooseContain(unref(itemForm).product_id, p.id) : ssrLooseEqual(unref(itemForm).product_id, p.id)) ? " selected" : ""}>${ssrInterpolate(p.brand)} ${ssrInterpolate(p.name)}</option>`);
        });
        _push(`<!--]--></select></div><div><label class="label">Status</label><select class="input"><!--[-->`);
        ssrRenderList(unref(TRADE_STATUSES), (name, id) => {
          _push(`<option${ssrRenderAttr("value", id)}${ssrIncludeBooleanAttr(Array.isArray(unref(itemForm).status_id) ? ssrLooseContain(unref(itemForm).status_id, id) : ssrLooseEqual(unref(itemForm).status_id, id)) ? " selected" : ""}>${ssrInterpolate(name)}</option>`);
        });
        _push(`<!--]--></select></div><div><label class="label">K\xF6pris (\xF6re)</label><input${ssrRenderAttr("value", unref(itemForm).buy_price)} type="number" class="input"></div><div><label class="label">K\xF6pfrit (\xF6re)</label><input${ssrRenderAttr("value", unref(itemForm).buy_shipping_cost)} type="number" class="input"></div><div><label class="label">F\xF6rs\xE4ljningspris (\xF6re)</label><input${ssrRenderAttr("value", unref(itemForm).sell_price)} type="number" class="input"></div><div><label class="label">F\xF6rpackning (\xF6re)</label><input${ssrRenderAttr("value", unref(itemForm).sell_packaging_cost)} type="number" class="input"></div><div><label class="label">Frakt (\xF6re)</label><input${ssrRenderAttr("value", unref(itemForm).sell_postage_cost)} type="number" class="input"></div><div><label class="label">Frakt mottaget (\xF6re)</label><input${ssrRenderAttr("value", unref(itemForm).sell_shipping_collected)} type="number" class="input"></div><div><label class="label">Lagerplats</label><input${ssrRenderAttr("value", unref(itemForm).storage)} type="number" class="input"></div><div><label class="label">L\xE4nk</label><input${ssrRenderAttr("value", unref(itemForm).source_link)} type="text" class="input"></div><div><label class="label">K\xF6pt datum</label><input${ssrRenderAttr("value", unref(itemForm).buy_date)} type="date" class="input"></div><div><label class="label">S\xE5lt datum</label><input${ssrRenderAttr("value", unref(itemForm).sell_date)} type="date" class="input"></div></div><div class="flex justify-end gap-2 pt-4"><button type="button" class="btn btn-secondary"> Avbryt </button><button type="submit" class="btn btn-primary">${ssrInterpolate(unref(editingItem) ? "Spara" : "L\xE4gg till")}</button></div></form></div></div>`);
      } else {
        _push(`<!---->`);
      }
      _push(`</div>`);
    };
  }
});
const _sfc_setup = _sfc_main.setup;
_sfc_main.setup = (props, ctx) => {
  const ssrContext = useSSRContext();
  (ssrContext.modules || (ssrContext.modules = /* @__PURE__ */ new Set())).add("pages/index.vue");
  return _sfc_setup ? _sfc_setup(props, ctx) : void 0;
};

export { _sfc_main as default };
//# sourceMappingURL=index-D4XjEbGB.mjs.map
