import { defineComponent, ref, computed, unref, useSSRContext } from 'vue';
import { ssrRenderAttrs, ssrInterpolate, ssrRenderClass, ssrRenderList, ssrRenderAttr, ssrIncludeBooleanAttr, ssrLooseContain, ssrLooseEqual } from 'vue/server-renderer';
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
  __name: "transactions",
  __ssrInlineRender: true,
  setup(__props) {
    useApi();
    const transactions = ref([]);
    const transactionTypes = ref([]);
    ref(false);
    const showAddModal = ref(false);
    const form = ref({
      date: (/* @__PURE__ */ new Date()).toISOString().split("T")[0],
      transaction_type: null,
      amount: 0
    });
    const totalIncome = computed(
      () => transactions.value.filter((t) => t.amount > 0).reduce((sum, t) => sum + t.amount, 0)
    );
    const totalExpenses = computed(
      () => Math.abs(transactions.value.filter((t) => t.amount < 0).reduce((sum, t) => sum + t.amount, 0))
    );
    const netAmount = computed(
      () => transactions.value.reduce((sum, t) => sum + t.amount, 0)
    );
    const formatCurrency = (cents) => `${(cents / 100).toFixed(2)} kr`;
    const formatDate = (dateStr) => new Date(dateStr).toLocaleDateString("sv-SE");
    const typeClass = (typeId) => {
      var _a, _b;
      if (!typeId) return "bg-slate-700 text-slate-300";
      const type = transactionTypes.value.find((t) => t.id === typeId);
      if (((_a = type == null ? void 0 : type.name) == null ? void 0 : _a.toLowerCase().includes("income")) || ((_b = type == null ? void 0 : type.name) == null ? void 0 : _b.toLowerCase().includes("sell"))) {
        return "badge badge-success";
      }
      return "badge badge-danger";
    };
    const transactionTypeName = (typeId) => {
      var _a;
      if (!typeId) return "Unknown";
      return ((_a = transactionTypes.value.find((t) => t.id === typeId)) == null ? void 0 : _a.name) || "Unknown";
    };
    return (_ctx, _push, _parent, _attrs) => {
      _push(`<div${ssrRenderAttrs(_attrs)}><div class="flex justify-between items-center mb-6"><h1 class="page-header">Transaktioner</h1><button class="btn btn-primary"> L\xE4gg till transaktion </button></div><div class="grid grid-cols-3 gap-4 mb-6"><div class="stat-card"><p class="stat-label">Total inkomst</p><p class="stat-value text-emerald-400">${ssrInterpolate(formatCurrency(unref(totalIncome)))}</p></div><div class="stat-card"><p class="stat-label">Total utgift</p><p class="stat-value text-red-400">${ssrInterpolate(formatCurrency(unref(totalExpenses)))}</p></div><div class="stat-card"><p class="stat-label">Netto</p><p class="${ssrRenderClass(unref(netAmount) >= 0 ? "stat-value text-emerald-400" : "stat-value text-red-400")}">${ssrInterpolate(formatCurrency(unref(netAmount)))}</p></div></div><div class="card overflow-hidden"><table class="table"><thead><tr><th>Datum</th><th>Typ</th><th>Belopp</th><th></th></tr></thead><tbody><!--[-->`);
      ssrRenderList(unref(transactions), (tx) => {
        _push(`<tr><td>${ssrInterpolate(formatDate(tx.date))}</td><td><span class="${ssrRenderClass([typeClass(tx.transaction_type), "badge"])}">${ssrInterpolate(transactionTypeName(tx.transaction_type))}</span></td><td class="${ssrRenderClass(tx.amount >= 0 ? "text-emerald-400" : "text-red-400")}">${ssrInterpolate(formatCurrency(tx.amount))}</td><td><button class="text-red-400 hover:text-red-300 text-sm"> Ta bort </button></td></tr>`);
      });
      _push(`<!--]--></tbody></table></div>`);
      if (unref(showAddModal)) {
        _push(`<div class="fixed inset-0 bg-black/70 flex items-center justify-center z-50"><div class="bg-slate-800 rounded-lg p-6 w-full max-w-md border border-slate-700"><h2 class="text-xl font-bold text-slate-100 mb-4">L\xE4gg till transaktion</h2><form class="space-y-4"><div><label class="label">Datum</label><input${ssrRenderAttr("value", unref(form).date)} type="date" class="input" required></div><div><label class="label">Typ</label><select class="input" required><option value=""${ssrIncludeBooleanAttr(Array.isArray(unref(form).transaction_type) ? ssrLooseContain(unref(form).transaction_type, "") : ssrLooseEqual(unref(form).transaction_type, "")) ? " selected" : ""}>V\xE4lj typ...</option><!--[-->`);
        ssrRenderList(unref(transactionTypes), (t) => {
          _push(`<option${ssrRenderAttr("value", t.id)}${ssrIncludeBooleanAttr(Array.isArray(unref(form).transaction_type) ? ssrLooseContain(unref(form).transaction_type, t.id) : ssrLooseEqual(unref(form).transaction_type, t.id)) ? " selected" : ""}>${ssrInterpolate(t.name)}</option>`);
        });
        _push(`<!--]--></select></div><div><label class="label">Belopp (\xF6re, negativt f\xF6r utgift)</label><input${ssrRenderAttr("value", unref(form).amount)} type="number" class="input" required></div><div class="flex justify-end gap-2 pt-4"><button type="button" class="btn btn-secondary"> Avbryt </button><button type="submit" class="btn btn-primary">L\xE4gg till</button></div></form></div></div>`);
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
  (ssrContext.modules || (ssrContext.modules = /* @__PURE__ */ new Set())).add("pages/transactions.vue");
  return _sfc_setup ? _sfc_setup(props, ctx) : void 0;
};

export { _sfc_main as default };
//# sourceMappingURL=transactions-jFXfCnRD.mjs.map
