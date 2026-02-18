import { defineComponent, computed, ref, useSSRContext } from 'vue';
import { ssrRenderAttrs, ssrInterpolate, ssrRenderList, ssrRenderClass, ssrIncludeBooleanAttr } from 'vue/server-renderer';
import { useRoute } from 'vue-router';
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
import '@supabase/ssr';

const _sfc_main = /* @__PURE__ */ defineComponent({
  __name: "[id]",
  __ssrInlineRender: true,
  setup(__props) {
    const route = useRoute();
    useApi();
    computed(() => parseInt(route.params.id));
    const conversation = ref(null);
    const messages = ref([]);
    const loading = ref(true);
    const generatingMessage = ref(false);
    const editingMessage = ref(null);
    const showAddIncomingModal = ref(false);
    const editForm = ref({
      content: ""
    });
    const incomingForm = ref({
      content: ""
    });
    const canGenerateReply = computed(() => {
      const lastMessage = messages.value[messages.value.length - 1];
      return (lastMessage == null ? void 0 : lastMessage.direction) === "incoming" && !hasPendingOutgoing.value;
    });
    const hasPendingOutgoing = computed(() => {
      return messages.value.some((m) => m.direction === "outgoing" && m.status === "pending");
    });
    const messageClass = (msg) => {
      const baseClass = "p-4 rounded-lg border";
      if (msg.direction === "outgoing") {
        return `${baseClass} bg-primary-900/20 border-primary-700`;
      } else {
        return `${baseClass} bg-slate-700/50 border-slate-600`;
      }
    };
    const messageStatusClass = (status) => {
      const classes = {
        pending: "px-2 py-1 bg-yellow-500/20 text-yellow-400 text-xs rounded",
        approved: "px-2 py-1 bg-green-500/20 text-green-400 text-xs rounded",
        sent: "px-2 py-1 bg-blue-500/20 text-blue-400 text-xs rounded",
        rejected: "px-2 py-1 bg-red-500/20 text-red-400 text-xs rounded",
        received: "px-2 py-1 bg-slate-500/20 text-slate-400 text-xs rounded"
      };
      return classes[status] || "px-2 py-1 bg-slate-500/20 text-slate-400 text-xs rounded";
    };
    const statusLabel = (status) => {
      const labels = {
        pending: "V\xE4ntar p\xE5 granskning",
        approved: "Godk\xE4nd",
        sent: "Skickad",
        rejected: "Nekad",
        received: "Mottagen"
      };
      return labels[status] || status;
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
    return (_ctx, _push, _parent, _attrs) => {
      _push(`<div${ssrRenderAttrs(_attrs)} data-v-a19735fe><div class="mb-6" data-v-a19735fe><button class="text-primary-400 hover:text-primary-300 mb-4" data-v-a19735fe> \u2190 Tillbaka till konversationer </button><h1 class="page-header" data-v-a19735fe>Konversation</h1>`);
      if (conversation.value) {
        _push(`<p class="text-slate-400 mt-2" data-v-a19735fe>${ssrInterpolate(conversation.value.listing_title)} - ${ssrInterpolate(conversation.value.marketplace_name)}</p>`);
      } else {
        _push(`<!---->`);
      }
      _push(`</div>`);
      if (loading.value) {
        _push(`<div class="text-center py-12 text-slate-500" data-v-a19735fe> Laddar meddelanden... </div>`);
      } else {
        _push(`<div class="space-y-6" data-v-a19735fe><div class="card p-6" data-v-a19735fe><h2 class="text-xl font-semibold text-slate-100 mb-4" data-v-a19735fe>Meddelandehistorik</h2>`);
        if (messages.value.length === 0) {
          _push(`<div class="text-center py-8 text-slate-500" data-v-a19735fe> Inga meddelanden \xE4n </div>`);
        } else {
          _push(`<div class="space-y-4" data-v-a19735fe><!--[-->`);
          ssrRenderList(messages.value, (msg) => {
            _push(`<div class="${ssrRenderClass(messageClass(msg))}" data-v-a19735fe><div class="flex justify-between items-start mb-2" data-v-a19735fe><span class="font-medium" data-v-a19735fe>${ssrInterpolate(msg.direction === "outgoing" ? "Du" : "S\xE4ljare")}</span><div class="flex items-center gap-2" data-v-a19735fe><span class="${ssrRenderClass(messageStatusClass(msg.status))}" data-v-a19735fe>${ssrInterpolate(statusLabel(msg.status))}</span><span class="text-xs text-slate-500" data-v-a19735fe>${ssrInterpolate(formatDate(msg.created_at))}</span></div></div><p class="text-slate-200" data-v-a19735fe>${ssrInterpolate(msg.content)}</p>`);
            if (msg.status === "pending" && msg.direction === "outgoing") {
              _push(`<div class="mt-4 flex gap-2" data-v-a19735fe><button class="btn btn-secondary text-sm" data-v-a19735fe> Redigera </button><button class="btn btn-primary text-sm" data-v-a19735fe> Godk\xE4nn och skicka </button><button class="btn btn-danger text-sm" data-v-a19735fe> Neka </button></div>`);
            } else {
              _push(`<!---->`);
            }
            _push(`</div>`);
          });
          _push(`<!--]--></div>`);
        }
        _push(`</div><div class="flex gap-4" data-v-a19735fe>`);
        if (messages.value.length === 0) {
          _push(`<button class="btn btn-primary"${ssrIncludeBooleanAttr(generatingMessage.value) ? " disabled" : ""} data-v-a19735fe>${ssrInterpolate(generatingMessage.value ? "Genererar..." : "Generera f\xF6rsta meddelande")}</button>`);
        } else if (canGenerateReply.value) {
          _push(`<button class="btn btn-primary"${ssrIncludeBooleanAttr(generatingMessage.value) ? " disabled" : ""} data-v-a19735fe>${ssrInterpolate(generatingMessage.value ? "Genererar..." : "Generera svar")}</button>`);
        } else {
          _push(`<!---->`);
        }
        _push(`<button class="btn btn-secondary" data-v-a19735fe> L\xE4gg till inkommande meddelande </button></div></div>`);
      }
      if (editingMessage.value) {
        _push(`<div class="fixed inset-0 bg-black/70 flex items-center justify-center z-50" data-v-a19735fe><div class="bg-slate-800 rounded-lg p-6 w-full max-w-2xl border border-slate-700" data-v-a19735fe><h2 class="text-xl font-bold text-slate-100 mb-4" data-v-a19735fe>Redigera meddelande</h2><form data-v-a19735fe><div class="mb-4" data-v-a19735fe><label class="label" data-v-a19735fe>Meddelande</label><textarea class="input h-32" required data-v-a19735fe>${ssrInterpolate(editForm.value.content)}</textarea></div><div class="flex justify-end gap-2" data-v-a19735fe><button type="button" class="btn btn-secondary" data-v-a19735fe> Avbryt </button><button type="submit" class="btn btn-primary" data-v-a19735fe> Spara </button></div></form></div></div>`);
      } else {
        _push(`<!---->`);
      }
      if (showAddIncomingModal.value) {
        _push(`<div class="fixed inset-0 bg-black/70 flex items-center justify-center z-50" data-v-a19735fe><div class="bg-slate-800 rounded-lg p-6 w-full max-w-2xl border border-slate-700" data-v-a19735fe><h2 class="text-xl font-bold text-slate-100 mb-4" data-v-a19735fe>L\xE4gg till inkommande meddelande</h2><form data-v-a19735fe><div class="mb-4" data-v-a19735fe><label class="label" data-v-a19735fe>Meddelande fr\xE5n s\xE4ljare</label><textarea class="input h-32" required data-v-a19735fe>${ssrInterpolate(incomingForm.value.content)}</textarea></div><div class="flex justify-end gap-2" data-v-a19735fe><button type="button" class="btn btn-secondary" data-v-a19735fe> Avbryt </button><button type="submit" class="btn btn-primary" data-v-a19735fe> L\xE4gg till </button></div></form></div></div>`);
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
  (ssrContext.modules || (ssrContext.modules = /* @__PURE__ */ new Set())).add("pages/conversations/[id].vue");
  return _sfc_setup ? _sfc_setup(props, ctx) : void 0;
};
const _id_ = /* @__PURE__ */ _export_sfc(_sfc_main, [["__scopeId", "data-v-a19735fe"]]);

export { _id_ as default };
//# sourceMappingURL=_id_-Cq6ZD4fT.mjs.map
