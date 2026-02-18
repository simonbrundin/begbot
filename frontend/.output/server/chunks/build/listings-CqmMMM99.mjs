import { defineComponent, ref, unref, useSSRContext } from 'vue';
import { ssrRenderAttrs, ssrRenderList, ssrInterpolate, ssrRenderClass, ssrRenderAttr, ssrIncludeBooleanAttr, ssrLooseContain, ssrLooseEqual } from 'vue/server-renderer';
import { L as LISTING_STATUSES } from './database-D1vXHN9-.mjs';
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
  __name: "listings",
  __ssrInlineRender: true,
  setup(__props) {
    useApi();
    const listings = ref([]);
    const products = ref([]);
    const marketplaces = ref([]);
    ref(false);
    const showAddModal = ref(false);
    const editingListing = ref(null);
    const defaultForm = {
      product_id: null,
      price: null,
      link: "",
      description: "",
      marketplace_id: null,
      status: "draft",
      is_my_listing: true
    };
    const form = ref({ ...defaultForm });
    const formatCurrency = (cents) => `${(cents / 100).toFixed(2)} kr`;
    const statusClass = (status) => {
      const classes = {
        draft: "badge badge-warning",
        active: "badge badge-success",
        sold: "badge badge-info",
        archived: "badge"
      };
      return classes[status] || classes.draft;
    };
    const marketplaceName = (id) => {
      var _a;
      if (!id) return "Unknown";
      return ((_a = marketplaces.value.find((m) => m.id === id)) == null ? void 0 : _a.name) || "Unknown";
    };
    const getProductName = (productId) => {
      if (!productId) return "Ok\xE4nd produkt";
      const product = products.value.find((p) => p.id === productId);
      if (!product) return "Ok\xE4nd produkt";
      return `${product.brand} ${product.name}`;
    };
    return (_ctx, _push, _parent, _attrs) => {
      _push(`<div${ssrRenderAttrs(_attrs)}><div class="flex justify-between items-center mb-6"><h1 class="page-header">Mina annonser</h1><button class="btn btn-primary"> L\xE4gg till annons </button></div><div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4"><!--[-->`);
      ssrRenderList(unref(listings), (listing) => {
        var _a;
        _push(`<div class="card overflow-hidden"><div class="p-4"><div class="flex justify-between items-start mb-2"><div>`);
        if (listing.product_id) {
          _push(`<p class="font-medium text-slate-100">${ssrInterpolate(getProductName(listing.product_id))}</p>`);
        } else {
          _push(`<p class="text-slate-500">Ok\xE4nd produkt</p>`);
        }
        _push(`<p class="text-lg font-bold text-primary-500">${ssrInterpolate(listing.price ? formatCurrency(listing.price) : "-")}</p>`);
        if (listing.valuation) {
          _push(`<p class="text-sm text-slate-400"> V\xE4rdering: ${ssrInterpolate(formatCurrency(listing.valuation))}</p>`);
        } else {
          _push(`<!---->`);
        }
        _push(`</div><span class="${ssrRenderClass(statusClass(listing.status))}">${ssrInterpolate(listing.status)}</span></div><p class="text-sm text-slate-400 mb-2">${ssrInterpolate((_a = listing.description) == null ? void 0 : _a.substring(0, 100))}... </p><div class="flex justify-between items-center text-sm text-slate-400"><span>${ssrInterpolate(marketplaceName(listing.marketplace_id))}</span><a${ssrRenderAttr("href", listing.link)} target="_blank" class="text-primary-400 hover:text-primary-300"> Visa </a></div></div><div class="px-4 py-3 bg-slate-800/50 border-t border-slate-700 flex justify-end gap-2"><button class="text-sm text-primary-400 hover:text-primary-300"> Redigera </button><button class="text-sm text-red-400 hover:text-red-300"> Ta bort </button></div></div>`);
      });
      _push(`<!--]--></div>`);
      if (unref(listings).length === 0) {
        _push(`<div class="text-center py-12 text-slate-500"> Inga annonser hittades. L\xE4gg till din f\xF6rsta annons! </div>`);
      } else {
        _push(`<!---->`);
      }
      if (unref(showAddModal) || unref(editingListing)) {
        _push(`<div class="fixed inset-0 bg-black/70 flex items-center justify-center z-50"><div class="bg-slate-800 rounded-lg p-6 w-full max-w-2xl max-h-[90vh] overflow-y-auto border border-slate-700"><h2 class="text-xl font-bold text-slate-100 mb-4">${ssrInterpolate(unref(editingListing) ? "Redigera annons" : "L\xE4gg till ny annons")}</h2><form class="space-y-4"><div class="grid grid-cols-2 gap-4"><div><label class="label">Produkt</label><select class="input"><option value=""${ssrIncludeBooleanAttr(Array.isArray(unref(form).product_id) ? ssrLooseContain(unref(form).product_id, "") : ssrLooseEqual(unref(form).product_id, "")) ? " selected" : ""}>V\xE4lj produkt...</option><!--[-->`);
        ssrRenderList(unref(products), (p) => {
          _push(`<option${ssrRenderAttr("value", p.id)}${ssrIncludeBooleanAttr(Array.isArray(unref(form).product_id) ? ssrLooseContain(unref(form).product_id, p.id) : ssrLooseEqual(unref(form).product_id, p.id)) ? " selected" : ""}>${ssrInterpolate(p.brand)} ${ssrInterpolate(p.name)}</option>`);
        });
        _push(`<!--]--></select></div><div><label class="label">Pris (\xF6re)</label><input${ssrRenderAttr("value", unref(form).price)} type="number" class="input"></div><div><label class="label">Status</label><select class="input"><!--[-->`);
        ssrRenderList(unref(LISTING_STATUSES), (status) => {
          _push(`<option${ssrRenderAttr("value", status)}${ssrIncludeBooleanAttr(Array.isArray(unref(form).status) ? ssrLooseContain(unref(form).status, status) : ssrLooseEqual(unref(form).status, status)) ? " selected" : ""}>${ssrInterpolate(status)}</option>`);
        });
        _push(`<!--]--></select></div><div><label class="label">Marknadsplats</label><select class="input"><option value=""${ssrIncludeBooleanAttr(Array.isArray(unref(form).marketplace_id) ? ssrLooseContain(unref(form).marketplace_id, "") : ssrLooseEqual(unref(form).marketplace_id, "")) ? " selected" : ""}>V\xE4lj marknadsplats...</option><!--[-->`);
        ssrRenderList(unref(marketplaces), (m) => {
          _push(`<option${ssrRenderAttr("value", m.id)}${ssrIncludeBooleanAttr(Array.isArray(unref(form).marketplace_id) ? ssrLooseContain(unref(form).marketplace_id, m.id) : ssrLooseEqual(unref(form).marketplace_id, m.id)) ? " selected" : ""}>${ssrInterpolate(m.name)}</option>`);
        });
        _push(`<!--]--></select></div><div class="col-span-2"><label class="label">L\xE4nk</label><input${ssrRenderAttr("value", unref(form).link)} type="text" class="input"></div><div class="col-span-2"><label class="label">Beskrivning</label><textarea class="input" rows="3">${ssrInterpolate(unref(form).description)}</textarea></div></div><div class="flex justify-end gap-2 pt-4"><button type="button" class="btn btn-secondary"> Avbryt </button><button type="submit" class="btn btn-primary">${ssrInterpolate(unref(editingListing) ? "Spara" : "L\xE4gg till")}</button></div></form></div></div>`);
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
  (ssrContext.modules || (ssrContext.modules = /* @__PURE__ */ new Set())).add("pages/listings.vue");
  return _sfc_setup ? _sfc_setup(props, ctx) : void 0;
};

export { _sfc_main as default };
//# sourceMappingURL=listings-CqmMMM99.mjs.map
