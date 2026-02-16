import { defineComponent, withAsyncContext, computed, unref, toValue, getCurrentInstance, onServerPrefetch, ref, shallowRef, nextTick, toRef, useSSRContext } from 'vue';
import { ssrRenderAttrs, ssrInterpolate, ssrRenderList, ssrRenderClass, ssrRenderAttr } from 'vue/server-renderer';
import { u as useNuxtApp, a as asyncDataDefaults, b as useRuntimeConfig, c as createError } from './server.mjs';
import { debounce } from 'perfect-debounce';
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

const isDefer = (dedupe) => dedupe === "defer" || dedupe === false;
function useAsyncData(...args) {
  var _a, _b, _c, _d, _e, _f, _g;
  const autoKey = typeof args[args.length - 1] === "string" ? args.pop() : void 0;
  if (_isAutoKeyNeeded(args[0], args[1])) {
    args.unshift(autoKey);
  }
  let [_key, _handler, options = {}] = args;
  const key = computed(() => toValue(_key));
  if (typeof key.value !== "string") {
    throw new TypeError("[nuxt] [useAsyncData] key must be a string.");
  }
  if (typeof _handler !== "function") {
    throw new TypeError("[nuxt] [useAsyncData] handler must be a function.");
  }
  const nuxtApp = useNuxtApp();
  (_a = options.server) != null ? _a : options.server = true;
  (_b = options.default) != null ? _b : options.default = getDefault;
  (_c = options.getCachedData) != null ? _c : options.getCachedData = getDefaultCachedData;
  (_d = options.lazy) != null ? _d : options.lazy = false;
  (_e = options.immediate) != null ? _e : options.immediate = true;
  (_f = options.deep) != null ? _f : options.deep = asyncDataDefaults.deep;
  (_g = options.dedupe) != null ? _g : options.dedupe = "cancel";
  options._functionName || "useAsyncData";
  nuxtApp._asyncData[key.value];
  function createInitialFetch() {
    var _a2;
    const initialFetchOptions = { cause: "initial", dedupe: options.dedupe };
    if (!((_a2 = nuxtApp._asyncData[key.value]) == null ? void 0 : _a2._init)) {
      initialFetchOptions.cachedData = options.getCachedData(key.value, nuxtApp, { cause: "initial" });
      nuxtApp._asyncData[key.value] = createAsyncData(nuxtApp, key.value, _handler, options, initialFetchOptions.cachedData);
    }
    return () => nuxtApp._asyncData[key.value].execute(initialFetchOptions);
  }
  const initialFetch = createInitialFetch();
  const asyncData = nuxtApp._asyncData[key.value];
  asyncData._deps++;
  const fetchOnServer = options.server !== false && nuxtApp.payload.serverRendered;
  if (fetchOnServer && options.immediate) {
    const promise = initialFetch();
    if (getCurrentInstance()) {
      onServerPrefetch(() => promise);
    } else {
      nuxtApp.hook("app:created", async () => {
        await promise;
      });
    }
  }
  const asyncReturn = {
    data: writableComputedRef(() => {
      var _a2;
      return (_a2 = nuxtApp._asyncData[key.value]) == null ? void 0 : _a2.data;
    }),
    pending: writableComputedRef(() => {
      var _a2;
      return (_a2 = nuxtApp._asyncData[key.value]) == null ? void 0 : _a2.pending;
    }),
    status: writableComputedRef(() => {
      var _a2;
      return (_a2 = nuxtApp._asyncData[key.value]) == null ? void 0 : _a2.status;
    }),
    error: writableComputedRef(() => {
      var _a2;
      return (_a2 = nuxtApp._asyncData[key.value]) == null ? void 0 : _a2.error;
    }),
    refresh: (...args2) => {
      var _a2;
      if (!((_a2 = nuxtApp._asyncData[key.value]) == null ? void 0 : _a2._init)) {
        const initialFetch2 = createInitialFetch();
        return initialFetch2();
      }
      return nuxtApp._asyncData[key.value].execute(...args2);
    },
    execute: (...args2) => asyncReturn.refresh(...args2),
    clear: () => {
      const entry = nuxtApp._asyncData[key.value];
      if (entry == null ? void 0 : entry._abortController) {
        try {
          entry._abortController.abort(new DOMException("AsyncData aborted by user.", "AbortError"));
        } finally {
          entry._abortController = void 0;
        }
      }
      clearNuxtDataByKey(nuxtApp, key.value);
    }
  };
  const asyncDataPromise = Promise.resolve(nuxtApp._asyncDataPromises[key.value]).then(() => asyncReturn);
  Object.assign(asyncDataPromise, asyncReturn);
  return asyncDataPromise;
}
function writableComputedRef(getter) {
  return computed({
    get() {
      var _a;
      return (_a = getter()) == null ? void 0 : _a.value;
    },
    set(value) {
      const ref2 = getter();
      if (ref2) {
        ref2.value = value;
      }
    }
  });
}
function _isAutoKeyNeeded(keyOrFetcher, fetcher) {
  if (typeof keyOrFetcher === "string") {
    return false;
  }
  if (typeof keyOrFetcher === "object" && keyOrFetcher !== null) {
    return false;
  }
  if (typeof keyOrFetcher === "function" && typeof fetcher === "function") {
    return false;
  }
  return true;
}
function clearNuxtDataByKey(nuxtApp, key) {
  if (key in nuxtApp.payload.data) {
    nuxtApp.payload.data[key] = void 0;
  }
  if (key in nuxtApp.payload._errors) {
    nuxtApp.payload._errors[key] = asyncDataDefaults.errorValue;
  }
  if (nuxtApp._asyncData[key]) {
    nuxtApp._asyncData[key].data.value = void 0;
    nuxtApp._asyncData[key].error.value = asyncDataDefaults.errorValue;
    {
      nuxtApp._asyncData[key].pending.value = false;
    }
    nuxtApp._asyncData[key].status.value = "idle";
  }
  if (key in nuxtApp._asyncDataPromises) {
    nuxtApp._asyncDataPromises[key] = void 0;
  }
}
function pick(obj, keys) {
  const newObj = {};
  for (const key of keys) {
    newObj[key] = obj[key];
  }
  return newObj;
}
function createAsyncData(nuxtApp, key, _handler, options, initialCachedData) {
  var _a, _b;
  (_b = (_a = nuxtApp.payload._errors)[key]) != null ? _b : _a[key] = asyncDataDefaults.errorValue;
  const hasCustomGetCachedData = options.getCachedData !== getDefaultCachedData;
  const handler = _handler ;
  const _ref = options.deep ? ref : shallowRef;
  const hasCachedData = initialCachedData != null;
  const unsubRefreshAsyncData = nuxtApp.hook("app:data:refresh", async (keys) => {
    if (!keys || keys.includes(key)) {
      await asyncData.execute({ cause: "refresh:hook" });
    }
  });
  const asyncData = {
    data: _ref(hasCachedData ? initialCachedData : options.default()),
    pending: shallowRef(!hasCachedData),
    error: toRef(nuxtApp.payload._errors, key),
    status: shallowRef("idle"),
    execute: (...args) => {
      var _a2, _b2;
      const [_opts, newValue = void 0] = args;
      const opts = _opts && newValue === void 0 && typeof _opts === "object" ? _opts : {};
      if (nuxtApp._asyncDataPromises[key]) {
        if (isDefer((_a2 = opts.dedupe) != null ? _a2 : options.dedupe)) {
          return nuxtApp._asyncDataPromises[key];
        }
      }
      if (opts.cause === "initial" || nuxtApp.isHydrating) {
        const cachedData = "cachedData" in opts ? opts.cachedData : options.getCachedData(key, nuxtApp, { cause: (_b2 = opts.cause) != null ? _b2 : "refresh:manual" });
        if (cachedData != null) {
          nuxtApp.payload.data[key] = asyncData.data.value = cachedData;
          asyncData.error.value = asyncDataDefaults.errorValue;
          asyncData.status.value = "success";
          return Promise.resolve(cachedData);
        }
      }
      {
        asyncData.pending.value = true;
      }
      if (asyncData._abortController) {
        asyncData._abortController.abort(new DOMException("AsyncData request cancelled by deduplication", "AbortError"));
      }
      asyncData._abortController = new AbortController();
      asyncData.status.value = "pending";
      const cleanupController = new AbortController();
      const promise = new Promise(
        (resolve, reject) => {
          var _a3, _b3;
          try {
            const timeout = (_a3 = opts.timeout) != null ? _a3 : options.timeout;
            const mergedSignal = mergeAbortSignals([(_b3 = asyncData._abortController) == null ? void 0 : _b3.signal, opts == null ? void 0 : opts.signal], cleanupController.signal, timeout);
            if (mergedSignal.aborted) {
              const reason = mergedSignal.reason;
              reject(reason instanceof Error ? reason : new DOMException(String(reason != null ? reason : "Aborted"), "AbortError"));
              return;
            }
            mergedSignal.addEventListener("abort", () => {
              const reason = mergedSignal.reason;
              reject(reason instanceof Error ? reason : new DOMException(String(reason != null ? reason : "Aborted"), "AbortError"));
            }, { once: true, signal: cleanupController.signal });
            return Promise.resolve(handler(nuxtApp, { signal: mergedSignal })).then(resolve, reject);
          } catch (err) {
            reject(err);
          }
        }
      ).then(async (_result) => {
        let result = _result;
        if (options.transform) {
          result = await options.transform(_result);
        }
        if (options.pick) {
          result = pick(result, options.pick);
        }
        nuxtApp.payload.data[key] = result;
        asyncData.data.value = result;
        asyncData.error.value = asyncDataDefaults.errorValue;
        asyncData.status.value = "success";
      }).catch((error) => {
        var _a3;
        if (nuxtApp._asyncDataPromises[key] && nuxtApp._asyncDataPromises[key] !== promise) {
          return nuxtApp._asyncDataPromises[key];
        }
        if ((_a3 = asyncData._abortController) == null ? void 0 : _a3.signal.aborted) {
          return nuxtApp._asyncDataPromises[key];
        }
        if (typeof DOMException !== "undefined" && error instanceof DOMException && error.name === "AbortError") {
          asyncData.status.value = "idle";
          return nuxtApp._asyncDataPromises[key];
        }
        asyncData.error.value = createError(error);
        asyncData.data.value = unref(options.default());
        asyncData.status.value = "error";
      }).finally(() => {
        {
          asyncData.pending.value = false;
        }
        cleanupController.abort();
        delete nuxtApp._asyncDataPromises[key];
      });
      nuxtApp._asyncDataPromises[key] = promise;
      return nuxtApp._asyncDataPromises[key];
    },
    _execute: debounce((...args) => asyncData.execute(...args), 0, { leading: true }),
    _default: options.default,
    _deps: 0,
    _init: true,
    _hash: void 0,
    _off: () => {
      var _a2;
      unsubRefreshAsyncData();
      if ((_a2 = nuxtApp._asyncData[key]) == null ? void 0 : _a2._init) {
        nuxtApp._asyncData[key]._init = false;
      }
      if (!hasCustomGetCachedData) {
        nextTick(() => {
          var _a3;
          if (!((_a3 = nuxtApp._asyncData[key]) == null ? void 0 : _a3._init)) {
            clearNuxtDataByKey(nuxtApp, key);
            asyncData.execute = () => Promise.resolve();
            asyncData.data.value = asyncDataDefaults.value;
          }
        });
      }
    }
  };
  return asyncData;
}
const getDefault = () => asyncDataDefaults.value;
const getDefaultCachedData = (key, nuxtApp, ctx) => {
  if (nuxtApp.isHydrating) {
    return nuxtApp.payload.data[key];
  }
  if (ctx.cause !== "refresh:manual" && ctx.cause !== "refresh:hook") {
    return nuxtApp.static.data[key];
  }
};
function mergeAbortSignals(signals, cleanupSignal, timeout) {
  var _a, _b, _c;
  const list = signals.filter((s) => !!s);
  if (typeof timeout === "number" && timeout >= 0) {
    const timeoutSignal = (_a = AbortSignal.timeout) == null ? void 0 : _a.call(AbortSignal, timeout);
    if (timeoutSignal) {
      list.push(timeoutSignal);
    }
  }
  if (AbortSignal.any) {
    return AbortSignal.any(list);
  }
  const controller = new AbortController();
  for (const sig of list) {
    if (sig.aborted) {
      const reason = (_b = sig.reason) != null ? _b : new DOMException("Aborted", "AbortError");
      try {
        controller.abort(reason);
      } catch {
        controller.abort();
      }
      return controller.signal;
    }
  }
  const onAbort = () => {
    var _a2;
    const abortedSignal = list.find((s) => s.aborted);
    const reason = (_a2 = abortedSignal == null ? void 0 : abortedSignal.reason) != null ? _a2 : new DOMException("Aborted", "AbortError");
    try {
      controller.abort(reason);
    } catch {
      controller.abort();
    }
  };
  for (const sig of list) {
    (_c = sig.addEventListener) == null ? void 0 : _c.call(sig, "abort", onAbort, { once: true, signal: cleanupSignal });
  }
  return controller.signal;
}
const _sfc_main = /* @__PURE__ */ defineComponent({
  __name: "ads",
  __ssrInlineRender: true,
  async setup(__props) {
    let __temp, __restore;
    const config = useRuntimeConfig();
    const {
      data: listings,
      error,
      pending
    } = ([__temp, __restore] = withAsyncContext(async () => useAsyncData("ads-listings", async () => {
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
          (item) => item.Listing && !item.Listing.is_my_listing
        );
      } catch (e) {
        console.error("Failed to fetch listings:", e);
        throw new Error(e.message || "Kunde inte h\xE4mta annonser");
      }
    })), __temp = await __temp, __restore(), __temp);
    const formatCurrency = (price) => {
      if (!price) return "-";
      return `${price.toLocaleString("sv-SE")} kr`;
    };
    const formatPriceAsSEK = (price) => {
      if (!price) return "-";
      return `${price.toLocaleString("sv-SE")} kr`;
    };
    const formatValuationAsSEK = (sek) => {
      if (!sek) return "-";
      return `${sek.toLocaleString("sv-SE")} kr`;
    };
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
      if (!id) return "Unknown";
      return "Blocket";
    };
    const errorMessage = computed(() => {
      var _a;
      return ((_a = error.value) == null ? void 0 : _a.message) || null;
    });
    return (_ctx, _push, _parent, _attrs) => {
      var _a;
      _push(`<div${ssrRenderAttrs(_attrs)}><div class="flex justify-between items-center mb-6"><h1 class="page-header">Hittade annonser</h1></div>`);
      if (unref(errorMessage)) {
        _push(`<div class="text-center py-12 text-red-400">${ssrInterpolate(unref(errorMessage))}</div>`);
      } else if (unref(pending)) {
        _push(`<div class="text-center py-12 text-slate-500"> Laddar... </div>`);
      } else {
        _push(`<!--[--><div class="grid grid-cols-1 gap-4"><!--[-->`);
        ssrRenderList(unref(listings), (item) => {
          var _a2, _b, _c, _d;
          _push(`<div class="card overflow-hidden"><div class="p-4">`);
          if (item.Listing) {
            _push(`<div class="flex justify-between items-start mb-2"><div>`);
            if (item.Listing.title) {
              _push(`<p class="font-medium text-slate-100">${ssrInterpolate(item.Listing.title)}</p>`);
            } else if (item.Product) {
              _push(`<p class="font-medium text-slate-100">${ssrInterpolate(item.Product.brand)} - ${ssrInterpolate(item.Product.name)}</p>`);
            } else {
              _push(`<p class="text-slate-500">Ok\xE4nd produkt</p>`);
            }
            _push(`<p class="text-sm text-slate-400"> Produkt: `);
            if (item.Product) {
              _push(`<!--[-->${ssrInterpolate(item.Product.brand)} - ${ssrInterpolate(item.Product.name)}<!--]-->`);
            } else {
              _push(`<!--[--> Ok\xE4nd produkt <!--]-->`);
            }
            _push(`</p><p class="text-lg font-bold text-primary-500">${ssrInterpolate(item.Listing.price ? formatPriceAsSEK(item.Listing.price) : "-")}</p><p class="text-sm text-slate-400"> Nypris: ${ssrInterpolate(formatValuationAsSEK((_b = (_a2 = item.Valuations) == null ? void 0 : _a2.find((v) => v.valuation_type_id === 4)) == null ? void 0 : _b.valuation))}</p><p class="text-sm text-slate-400"> Frakt: ${ssrInterpolate(item.Listing.shipping_cost !== null && item.Listing.shipping_cost !== void 0 ? formatCurrency(item.Listing.shipping_cost) : "Ok\xE4nt")}</p><p class="text-sm text-slate-400"> V\xE4rdering: ${ssrInterpolate(item.Listing.valuation ? formatValuationAsSEK(item.Listing.valuation) : "-")}</p>`);
            if (item.Valuations && item.Valuations.filter((v) => v.valuation_type_id !== 4).length > 0) {
              _push(`<div class="mt-1"><p class="text-xs text-slate-500 mb-1">Delv\xE4rderingar:</p><div class="flex flex-wrap gap-2"><!--[-->`);
              ssrRenderList(item.Valuations.filter((v) => v.valuation_type_id !== 4), (v) => {
                _push(`<span class="text-xs bg-slate-700 px-2 py-1 rounded">${ssrInterpolate(formatValuationAsSEK(v.valuation))} - ${ssrInterpolate(v.valuation_type)}</span>`);
              });
              _push(`<!--]--></div></div>`);
            } else {
              _push(`<!---->`);
            }
            if (item.PotentialProfit !== void 0) {
              _push(`<p class="${ssrRenderClass([item.PotentialProfit > 0 ? "text-emerald-400" : "text-red-400", "text-sm font-medium"])}"> Vinst: ${ssrInterpolate(formatPriceAsSEK(item.PotentialProfit))} `);
              if (item.DiscountPercent !== void 0) {
                _push(`<span class="text-slate-400 ml-2"> (${ssrInterpolate(item.DiscountPercent.toFixed(1))}% rabatt) </span>`);
              } else {
                _push(`<!---->`);
              }
              _push(`</p>`);
            } else {
              _push(`<p class="text-sm text-slate-500"> Ingen v\xE4rdering tillg\xE4nglig </p>`);
            }
            _push(`</div><span class="${ssrRenderClass(statusClass(item.Listing.status))}">${ssrInterpolate(item.Listing.status)}</span></div>`);
          } else {
            _push(`<!---->`);
          }
          if ((_c = item.Listing) == null ? void 0 : _c.description) {
            _push(`<p class="text-sm text-slate-400 mb-2">${ssrInterpolate((_d = item.Listing.description) == null ? void 0 : _d.substring(0, 100))}... </p>`);
          } else {
            _push(`<!---->`);
          }
          if (item.Listing) {
            _push(`<div class="flex justify-between items-center text-sm text-slate-400"><span>${ssrInterpolate(marketplaceName(item.Listing.marketplace_id))}</span><a${ssrRenderAttr("href", item.Listing.link)} target="_blank" class="text-primary-400 hover:text-primary-300"> Visa </a></div>`);
          } else {
            _push(`<!---->`);
          }
          _push(`</div></div>`);
        });
        _push(`<!--]--></div>`);
        if (((_a = unref(listings)) == null ? void 0 : _a.length) === 0) {
          _push(`<div class="text-center py-12 text-slate-500"> Inga annonser hittades. </div>`);
        } else {
          _push(`<!---->`);
        }
        _push(`<!--]-->`);
      }
      _push(`</div>`);
    };
  }
});
const _sfc_setup = _sfc_main.setup;
_sfc_main.setup = (props, ctx) => {
  const ssrContext = useSSRContext();
  (ssrContext.modules || (ssrContext.modules = /* @__PURE__ */ new Set())).add("pages/ads.vue");
  return _sfc_setup ? _sfc_setup(props, ctx) : void 0;
};

export { _sfc_main as default };
//# sourceMappingURL=ads-Ba2aCWAk.mjs.map
