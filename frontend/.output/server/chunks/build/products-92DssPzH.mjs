import { defineComponent, ref, unref, useSSRContext } from 'vue';
import { ssrRenderAttrs, ssrRenderList, ssrInterpolate, ssrRenderClass, ssrRenderAttr } from 'vue/server-renderer';
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
  __name: "products",
  __ssrInlineRender: true,
  setup(__props) {
    useApi();
    const products = ref([]);
    ref(false);
    const showAddModal = ref(false);
    const editingProduct = ref(null);
    const defaultForm = {
      brand: "",
      name: "",
      category: "",
      model_variant: "",
      sell_packaging_cost: 0,
      sell_postage_cost: 0,
      enabled: false
    };
    const form = ref({ ...defaultForm });
    const formatDate = (dateStr) => new Date(dateStr).toLocaleDateString("sv-SE");
    return (_ctx, _push, _parent, _attrs) => {
      _push(`<div${ssrRenderAttrs(_attrs)}><div class="flex justify-between items-center mb-6"><h1 class="page-header">Produkter</h1><button class="btn btn-primary"> L\xE4gg till produkt </button></div><div class="card overflow-hidden"><table class="table"><thead><tr><th>M\xE4rke</th><th>Namn</th><th>Kategori</th><th>Variant</th><th>Aktiverad</th><th>Skapad</th><th></th></tr></thead><tbody><!--[-->`);
      ssrRenderList(unref(products), (product) => {
        _push(`<tr><td class="font-medium text-slate-100">${ssrInterpolate(product.brand || "-")}</td><td>${ssrInterpolate(product.name || "-")}</td><td>${ssrInterpolate(product.category || "-")}</td><td>${ssrInterpolate(product.model_variant || "-")}</td><td><button class="${ssrRenderClass(product.enabled ? "badge badge-success" : "badge")}">${ssrInterpolate(product.enabled ? "Ja" : "Nej")}</button></td><td class="text-sm text-slate-400">${ssrInterpolate(formatDate(product.created_at))}</td><td><button class="text-primary-400 hover:text-primary-300"> Redigera </button></td></tr>`);
      });
      _push(`<!--]--></tbody></table></div>`);
      if (unref(showAddModal) || unref(editingProduct)) {
        _push(`<div class="fixed inset-0 bg-black/70 flex items-center justify-center z-50"><div class="bg-slate-800 rounded-lg p-6 w-full max-w-lg border border-slate-700"><h2 class="text-xl font-bold text-slate-100 mb-4">${ssrInterpolate(unref(editingProduct) ? "Redigera produkt" : "L\xE4gg till ny produkt")}</h2><form class="space-y-4"><div><label class="label">M\xE4rke</label><input${ssrRenderAttr("value", unref(form).brand)} type="text" class="input"></div><div><label class="label">Namn</label><input${ssrRenderAttr("value", unref(form).name)} type="text" class="input"></div><div><label class="label">Kategori</label><input${ssrRenderAttr("value", unref(form).category)} type="text" class="input" placeholder="t.ex., telefon"></div><div><label class="label">Modellvariant</label><input${ssrRenderAttr("value", unref(form).model_variant)} type="text" class="input" placeholder="t.ex., 256GB"></div><div class="grid grid-cols-2 gap-4"><div><label class="label">F\xF6rpackning (\xF6re)</label><input${ssrRenderAttr("value", unref(form).sell_packaging_cost)} type="number" class="input"></div><div><label class="label">Frakt (\xF6re)</label><input${ssrRenderAttr("value", unref(form).sell_postage_cost)} type="number" class="input"></div></div><div class="flex justify-end gap-2 pt-4"><button type="button" class="btn btn-secondary"> Avbryt </button><button type="submit" class="btn btn-primary">${ssrInterpolate(unref(editingProduct) ? "Spara" : "L\xE4gg till")}</button></div></form></div></div>`);
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
  (ssrContext.modules || (ssrContext.modules = /* @__PURE__ */ new Set())).add("pages/products.vue");
  return _sfc_setup ? _sfc_setup(props, ctx) : void 0;
};

export { _sfc_main as default };
//# sourceMappingURL=products-92DssPzH.mjs.map
