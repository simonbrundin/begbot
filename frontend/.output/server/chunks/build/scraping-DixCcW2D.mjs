import { defineComponent, ref, computed, unref, watch, nextTick, mergeProps, useSSRContext } from 'vue';
import { ssrRenderAttrs, ssrIncludeBooleanAttr, ssrInterpolate, ssrRenderStyle, ssrRenderComponent, ssrRenderList, ssrRenderClass, ssrRenderAttr, ssrLooseContain, ssrLooseEqual } from 'vue/server-renderer';
import { b as useRuntimeConfig } from './server.mjs';
import { u as useApi } from './useApi-EIa4-qJb.mjs';
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

const _sfc_main$1 = /* @__PURE__ */ defineComponent({
  __name: "ScraperLogConsole",
  __ssrInlineRender: true,
  props: {
    logs: {},
    isConnected: { type: Boolean },
    error: {}
  },
  emits: ["clear"],
  setup(__props) {
    const props = __props;
    const logContainer = ref();
    watch(() => props.logs.length, () => {
      nextTick(() => {
        if (logContainer.value) {
          logContainer.value.scrollTop = logContainer.value.scrollHeight;
        }
      });
    });
    const formatTime = (timestamp) => {
      const date = new Date(timestamp);
      return date.toLocaleTimeString("sv-SE", {
        hour: "2-digit",
        minute: "2-digit",
        second: "2-digit"
      });
    };
    const getLevelClass = (level) => {
      switch (level) {
        case "error":
          return "text-red-400";
        case "warning":
          return "text-amber-400";
        case "info":
        default:
          return "text-blue-400";
      }
    };
    return (_ctx, _push, _parent, _attrs) => {
      _push(`<div${ssrRenderAttrs(mergeProps({ class: "card p-4 mt-4" }, _attrs))}><div class="flex justify-between items-center mb-3"><h3 class="text-sm font-medium text-slate-300">Scraper-logg</h3><div class="flex items-center gap-3">`);
      if (__props.isConnected) {
        _push(`<span class="flex items-center gap-1 text-xs text-emerald-400"><span class="w-2 h-2 bg-emerald-400 rounded-full animate-pulse"></span> Live </span>`);
      } else {
        _push(`<!---->`);
      }
      _push(`<button class="text-xs text-slate-400 hover:text-slate-300"${ssrIncludeBooleanAttr(__props.logs.length === 0) ? " disabled" : ""}> Rensa </button></div></div><div class="bg-slate-950 rounded-lg p-3 font-mono text-xs h-64 overflow-y-auto space-y-1">`);
      if (__props.logs.length === 0) {
        _push(`<div class="text-slate-600 italic"> V\xE4ntar p\xE5 loggmeddelanden... </div>`);
      } else {
        _push(`<!---->`);
      }
      _push(`<!--[-->`);
      ssrRenderList(__props.logs, (log, index) => {
        _push(`<div class="flex gap-2"><span class="text-slate-500 shrink-0">${ssrInterpolate(formatTime(log.timestamp))}</span><span class="${ssrRenderClass([getLevelClass(log.level), "shrink-0 w-14"])}"> [${ssrInterpolate(log.level.toUpperCase())}] </span><span class="text-slate-300 break-all">${ssrInterpolate(log.message)}</span></div>`);
      });
      _push(`<!--]--></div>`);
      if (__props.error) {
        _push(`<div class="mt-2 text-sm text-red-400">${ssrInterpolate(__props.error)}</div>`);
      } else {
        _push(`<!---->`);
      }
      _push(`</div>`);
    };
  }
});
const _sfc_setup$1 = _sfc_main$1.setup;
_sfc_main$1.setup = (props, ctx) => {
  const ssrContext = useSSRContext();
  (ssrContext.modules || (ssrContext.modules = /* @__PURE__ */ new Set())).add("components/ScraperLogConsole.vue");
  return _sfc_setup$1 ? _sfc_setup$1(props, ctx) : void 0;
};
function useScraperLogs(jobId) {
  const logs = ref([]);
  const isConnected = ref(false);
  const error = ref(null);
  let eventSource = null;
  let reconnectAttempts = 0;
  const MAX_RECONNECT_ATTEMPTS = 3;
  const connect = () => {
    if (!jobId.value) return;
    logs.value = [];
    error.value = null;
    reconnectAttempts = 0;
    const config = useRuntimeConfig();
    const apiBase = config.public.apiBase || "http://localhost:8081";
    const url = `${apiBase}/api/fetch-ads/logs/${jobId.value}`;
    console.log("Connecting to SSE:", url);
    eventSource = new EventSource(url);
    eventSource.onopen = () => {
      console.log("SSE connection opened");
      isConnected.value = true;
      error.value = null;
      reconnectAttempts = 0;
    };
    eventSource.onmessage = (event) => {
      console.log("SSE message received:", event.data);
      try {
        const log = JSON.parse(event.data);
        logs.value.push(log);
      } catch (e) {
        console.error("Failed to parse log entry:", e);
      }
    };
    eventSource.onerror = (err) => {
      console.error("SSE error:", err);
      isConnected.value = false;
      if (reconnectAttempts < MAX_RECONNECT_ATTEMPTS) {
        reconnectAttempts++;
        console.log(`Reconnecting... attempt ${reconnectAttempts}`);
        setTimeout(() => {
          disconnect();
          connect();
        }, 1e3 * reconnectAttempts);
      } else {
        error.value = "Anslutningen br\xF6ts. F\xF6rs\xF6k uppdatera sidan.";
      }
    };
  };
  const disconnect = () => {
    if (eventSource) {
      eventSource.close();
      eventSource = null;
    }
    isConnected.value = false;
  };
  const clearLogs = () => {
    logs.value = [];
  };
  watch(jobId, (newJobId, oldJobId) => {
    console.log("JobId changed:", newJobId, "old:", oldJobId);
    if (newJobId && newJobId !== oldJobId) {
      disconnect();
      connect();
    } else if (!newJobId) {
      disconnect();
      logs.value = [];
    }
  });
  return {
    logs,
    isConnected,
    error,
    connect,
    disconnect,
    clearLogs
  };
}
const _sfc_main = /* @__PURE__ */ defineComponent({
  __name: "scraping",
  __ssrInlineRender: true,
  setup(__props) {
    useApi();
    const searchTerms = ref([]);
    const hasSearchTerms = computed(() => searchTerms.value.length > 0);
    const marketplaces = ref([]);
    ref(false);
    const showAddModal = ref(false);
    const isFetching = ref(false);
    const currentJobId = ref(null);
    const fetchStatus = ref(null);
    const { logs: scraperLogs, isConnected: isLogConnected, error: logError, clearLogs: clearScraperLogs } = useScraperLogs(currentJobId);
    const form = ref({
      description: "",
      url: "",
      marketplace_id: null,
      is_active: true
    });
    const fetchProgress = computed(() => {
      if (!fetchStatus.value) return 0;
      return fetchStatus.value.progress;
    });
    const fetchStatusText = computed(() => {
      if (!fetchStatus.value) return "V\xE4ntar";
      switch (fetchStatus.value.status) {
        case "pending":
          return "Startar...";
        case "running":
          return "H\xE4mtar annonser...";
        case "completed":
          return "Klar!";
        case "failed":
          return "Misslyckades";
        default:
          return "V\xE4ntar";
      }
    });
    const marketplaceName = (id) => {
      var _a;
      if (!id) return "Unknown";
      return ((_a = marketplaces.value.find((m) => m.id === id)) == null ? void 0 : _a.name) || "Unknown";
    };
    return (_ctx, _push, _parent, _attrs) => {
      var _a, _b, _c;
      const _component_ScraperLogConsole = _sfc_main$1;
      _push(`<div${ssrRenderAttrs(_attrs)}><div class="flex justify-between items-center mb-6"><h1 class="page-header">S\xF6kord</h1><div class="flex gap-2"><button${ssrIncludeBooleanAttr(unref(isFetching)) ? " disabled" : ""} class="btn btn-primary">${ssrInterpolate(unref(isFetching) ? "H\xE4mtar..." : "H\xE4mta annonser")}</button><button class="btn btn-secondary"> L\xE4gg till s\xF6kord </button></div></div>`);
      if (unref(isFetching) || unref(fetchStatus)) {
        _push(`<div class="card p-4 mb-6"><div class="flex justify-between items-center mb-2"><span class="font-medium text-slate-100">${ssrInterpolate(unref(fetchStatusText))}</span><span class="text-sm text-slate-400">${ssrInterpolate(unref(fetchProgress))}%</span></div><div class="w-full bg-slate-700 rounded-full h-2.5"><div class="bg-primary-500 h-2.5 rounded-full transition-all duration-300" style="${ssrRenderStyle({ width: unref(fetchProgress) + "%" })}"></div></div>`);
        if ((_a = unref(fetchStatus)) == null ? void 0 : _a.current_query) {
          _push(`<p class="text-sm text-slate-400 mt-2"> Bearbetar: ${ssrInterpolate(unref(fetchStatus).current_query)}</p>`);
        } else {
          _push(`<!---->`);
        }
        if (((_b = unref(fetchStatus)) == null ? void 0 : _b.ads_found) !== void 0 && unref(fetchStatus).ads_found > 0) {
          _push(`<p class="text-sm text-emerald-400 mt-2"> Hittade ${ssrInterpolate(unref(fetchStatus).ads_found)} annonser </p>`);
        } else {
          _push(`<!---->`);
        }
        if ((_c = unref(fetchStatus)) == null ? void 0 : _c.error) {
          _push(`<p class="text-sm text-red-400 mt-2"> Fel: ${ssrInterpolate(unref(fetchStatus).error)}</p>`);
        } else {
          _push(`<!---->`);
        }
        _push(`</div>`);
      } else {
        _push(`<!---->`);
      }
      if (unref(currentJobId)) {
        _push(ssrRenderComponent(_component_ScraperLogConsole, {
          logs: unref(scraperLogs),
          "is-connected": unref(isLogConnected),
          error: unref(logError),
          onClear: unref(clearScraperLogs)
        }, null, _parent));
      } else {
        _push(`<!---->`);
      }
      if (unref(hasSearchTerms)) {
        _push(`<div class="card overflow-hidden mt-6"><table class="table"><thead><tr><th>Beskrivning</th><th>URL</th><th>Marknadsplats</th><th>Status</th><th></th></tr></thead><tbody><!--[-->`);
        ssrRenderList(unref(searchTerms), (term) => {
          _push(`<tr><td class="font-medium text-slate-100">${ssrInterpolate(term.description)}</td><td class="text-sm text-slate-400 truncate max-w-xs">${ssrInterpolate(term.url)}</td><td>${ssrInterpolate(marketplaceName(term.marketplace_id))}</td><td><button class="${ssrRenderClass(term.is_active ? "badge badge-success" : "badge")}">${ssrInterpolate(term.is_active ? "Aktiv" : "Inaktiv")}</button></td><td><button class="text-red-400 hover:text-red-300 text-sm"> Ta bort </button></td></tr>`);
        });
        _push(`<!--]--></tbody></table></div>`);
      } else {
        _push(`<!---->`);
      }
      if (!unref(hasSearchTerms)) {
        _push(`<div class="text-center py-12 text-slate-500"> Inga s\xF6kord konfigurerade. L\xE4gg till ditt f\xF6rsta f\xF6r att b\xF6rja skrapa! </div>`);
      } else {
        _push(`<!---->`);
      }
      if (unref(showAddModal)) {
        _push(`<div class="fixed inset-0 bg-black/70 flex items-center justify-center z-50"><div class="bg-slate-800 rounded-lg p-6 w-full max-w-lg border border-slate-700"><h2 class="text-xl font-bold text-slate-100 mb-4">L\xE4gg till s\xF6kord</h2><form class="space-y-4"><div><label class="label">Beskrivning</label><input${ssrRenderAttr("value", unref(form).description)} type="text" class="input" placeholder="t.ex., iPhone 15 Pro" required></div><div><label class="label">S\xF6k-URL</label><input${ssrRenderAttr("value", unref(form).url)} type="url" class="input" placeholder="https://www.blocket.se/..." required></div><div><label class="label">Marknadsplats</label><select class="input"><option value=""${ssrIncludeBooleanAttr(Array.isArray(unref(form).marketplace_id) ? ssrLooseContain(unref(form).marketplace_id, "") : ssrLooseEqual(unref(form).marketplace_id, "")) ? " selected" : ""}>V\xE4lj marknadsplats...</option><!--[-->`);
        ssrRenderList(unref(marketplaces), (m) => {
          _push(`<option${ssrRenderAttr("value", m.id)}${ssrIncludeBooleanAttr(Array.isArray(unref(form).marketplace_id) ? ssrLooseContain(unref(form).marketplace_id, m.id) : ssrLooseEqual(unref(form).marketplace_id, m.id)) ? " selected" : ""}>${ssrInterpolate(m.name)}</option>`);
        });
        _push(`<!--]--></select></div><div class="flex justify-end gap-2 pt-4"><button type="button" class="btn btn-secondary"> Avbryt </button><button type="submit" class="btn btn-primary">L\xE4gg till</button></div></form></div></div>`);
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
  (ssrContext.modules || (ssrContext.modules = /* @__PURE__ */ new Set())).add("pages/scraping.vue");
  return _sfc_setup ? _sfc_setup(props, ctx) : void 0;
};

export { _sfc_main as default };
//# sourceMappingURL=scraping-DixCcW2D.mjs.map
