import { defineComponent, ref, watch, useSSRContext } from 'vue';
import { ssrRenderAttrs, ssrRenderClass, ssrInterpolate, ssrRenderList } from 'vue/server-renderer';
import { u as useApi } from './useApi-C1L14LOp.mjs';
import { _ as _export_sfc } from './server.mjs';
import './loading--Nv5gQJt.mjs';
import 'pinia';
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
import 'vue-router';
import '@supabase/ssr';

const _sfc_main = /* @__PURE__ */ defineComponent({
  __name: "conversations",
  __ssrInlineRender: true,
  setup(__props) {
    const api = useApi();
    const conversations2 = ref([]);
    const loading = ref(true);
    const showNeedsReview = ref(true);
    const fetchConversations = async () => {
      loading.value = true;
      try {
        const endpoint = showNeedsReview.value ? "/conversations?needs_review=true" : "/conversations";
        conversations2.value = await api.get(endpoint);
      } catch (error) {
        console.error("Failed to fetch conversations:", error);
      } finally {
        loading.value = false;
      }
    };
    const formatCurrency = (amount) => {
      return new Intl.NumberFormat("sv-SE", { style: "currency", currency: "SEK" }).format(amount / 100);
    };
    const formatDate = (dateStr) => {
      const date = new Date(dateStr);
      return new Intl.DateTimeFormat("sv-SE", {
        year: "numeric",
        month: "short",
        day: "numeric",
        hour: "2-digit",
        minute: "2-digit"
      }).format(date);
    };
    const statusClass = (status) => {
      const classes = {
        active: "px-2 py-1 bg-green-500/20 text-green-400 text-xs rounded",
        closed: "px-2 py-1 bg-gray-500/20 text-gray-400 text-xs rounded",
        archived: "px-2 py-1 bg-slate-500/20 text-slate-400 text-xs rounded"
      };
      return classes[status] || "px-2 py-1 bg-slate-500/20 text-slate-400 text-xs rounded";
    };
    watch(showNeedsReview, () => {
      fetchConversations();
    });
    return (_ctx, _push, _parent, _attrs) => {
      _push(`<div${ssrRenderAttrs(_attrs)} data-v-0a179bab><div class="flex justify-between items-center mb-6" data-v-0a179bab><h1 class="page-header" data-v-0a179bab>Meddelanden</h1><div class="flex gap-2" data-v-0a179bab><button class="${ssrRenderClass(showNeedsReview.value ? "btn btn-primary" : "btn btn-secondary")}" data-v-0a179bab>${ssrInterpolate(showNeedsReview.value ? "Beh\xF6ver granskning" : "Alla konversationer")}</button></div></div>`);
      if (loading.value) {
        _push(`<div class="text-center py-12 text-slate-500" data-v-0a179bab> Laddar konversationer... </div>`);
      } else if (conversations2.value.length === 0) {
        _push(`<div class="text-center py-12 text-slate-500" data-v-0a179bab>`);
        if (showNeedsReview.value) {
          _push(`<p data-v-0a179bab>Inga konversationer beh\xF6ver granskning just nu.</p>`);
        } else {
          _push(`<p data-v-0a179bab>Inga konversationer hittades.</p>`);
        }
        _push(`</div>`);
      } else {
        _push(`<div class="space-y-4" data-v-0a179bab><!--[-->`);
        ssrRenderList(conversations2.value, (conv) => {
          _push(`<div class="card p-4 hover:border-primary-500 cursor-pointer transition-colors" data-v-0a179bab><div class="flex justify-between items-start mb-2" data-v-0a179bab><div class="flex-1" data-v-0a179bab><h3 class="text-lg font-semibold text-slate-100" data-v-0a179bab>${ssrInterpolate(conv.listing_title)}</h3><p class="text-sm text-slate-400" data-v-0a179bab>${ssrInterpolate(conv.marketplace_name)}</p></div><div class="text-right" data-v-0a179bab>`);
          if (conv.listing_price) {
            _push(`<p class="text-lg font-bold text-primary-500" data-v-0a179bab>${ssrInterpolate(formatCurrency(conv.listing_price))}</p>`);
          } else {
            _push(`<!---->`);
          }
          if (conv.pending_count > 0) {
            _push(`<span class="inline-block mt-1 px-2 py-1 bg-yellow-500/20 text-yellow-400 text-xs rounded" data-v-0a179bab>${ssrInterpolate(conv.pending_count)} att granska </span>`);
          } else {
            _push(`<!---->`);
          }
          _push(`</div></div><div class="flex justify-between items-center text-sm text-slate-400 mt-2" data-v-0a179bab><span class="${ssrRenderClass(statusClass(conv.status))}" data-v-0a179bab>${ssrInterpolate(conv.status)}</span><span data-v-0a179bab>${ssrInterpolate(formatDate(conv.updated_at))}</span></div></div>`);
        });
        _push(`<!--]--></div>`);
      }
      _push(`</div>`);
    };
  }
});
const _sfc_setup = _sfc_main.setup;
_sfc_main.setup = (props, ctx) => {
  const ssrContext = useSSRContext();
  (ssrContext.modules || (ssrContext.modules = /* @__PURE__ */ new Set())).add("pages/conversations.vue");
  return _sfc_setup ? _sfc_setup(props, ctx) : void 0;
};
const conversations = /* @__PURE__ */ _export_sfc(_sfc_main, [["__scopeId", "data-v-0a179bab"]]);

export { conversations as default };
//# sourceMappingURL=conversations-DcOErtXf.mjs.map
