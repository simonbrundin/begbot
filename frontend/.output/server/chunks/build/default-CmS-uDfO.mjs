import { defineComponent, mergeProps, unref, withCtx, createVNode, useSSRContext } from 'vue';
import { ssrRenderAttrs, ssrRenderComponent, ssrRenderSlot } from 'vue/server-renderer';
import { _ as __nuxt_component_0 } from './nuxt-link-C9A8ylPO.mjs';
import { u as useLoadingStore } from './loading-qsg6mAmB.mjs';
import '../nitro/nitro.mjs';
import 'node:http';
import 'node:https';
import 'node:events';
import 'node:buffer';
import 'node:fs';
import 'node:path';
import 'node:crypto';
import 'node:url';
import './server.mjs';
import '../routes/renderer.mjs';
import 'vue-bundle-renderer/runtime';
import 'unhead/server';
import 'devalue';
import 'unhead/utils';
import 'unhead/plugins';
import 'vue-router';
import '@supabase/ssr';

const _sfc_main$1 = /* @__PURE__ */ defineComponent({
  __name: "LoadingSpinner",
  __ssrInlineRender: true,
  props: {
    visible: { type: Boolean }
  },
  setup(__props) {
    return (_ctx, _push, _parent, _attrs) => {
      if (__props.visible) {
        _push(`<div${ssrRenderAttrs(mergeProps({
          class: "fixed inset-0 bg-slate-900/60 backdrop-blur-sm flex items-center justify-center z-50",
          "aria-label": "Laddar...",
          role: "status"
        }, _attrs))}><div class="relative"><div class="w-16 h-16 rounded-full border-4 border-slate-700"></div><div class="absolute inset-0 w-16 h-16 rounded-full border-4 border-emerald-400 border-t-transparent animate-spin"></div><div class="absolute inset-0 flex items-center justify-center"><div class="w-2 h-2 rounded-full bg-emerald-400 animate-pulse"></div></div></div></div>`);
      } else {
        _push(`<!---->`);
      }
    };
  }
});
const _sfc_setup$1 = _sfc_main$1.setup;
_sfc_main$1.setup = (props, ctx) => {
  const ssrContext = useSSRContext();
  (ssrContext.modules || (ssrContext.modules = /* @__PURE__ */ new Set())).add("components/LoadingSpinner.vue");
  return _sfc_setup$1 ? _sfc_setup$1(props, ctx) : void 0;
};
const _sfc_main = /* @__PURE__ */ defineComponent({
  __name: "default",
  __ssrInlineRender: true,
  setup(__props) {
    const loadingStore = useLoadingStore();
    return (_ctx, _push, _parent, _attrs) => {
      const _component_LoadingSpinner = _sfc_main$1;
      const _component_NuxtLink = __nuxt_component_0;
      _push(`<div${ssrRenderAttrs(mergeProps({ class: "min-h-screen bg-slate-900 text-slate-100" }, _attrs))}>`);
      _push(ssrRenderComponent(_component_LoadingSpinner, {
        visible: unref(loadingStore).showSpinner
      }, null, _parent));
      _push(`<aside class="fixed left-0 top-0 h-full w-64 bg-slate-800 border-r border-slate-700 p-4"><div class="mb-8"><h1 class="text-xl font-bold text-emerald-400">Begbot</h1></div><nav class="space-y-1">`);
      _push(ssrRenderComponent(_component_NuxtLink, {
        to: "/",
        class: "flex items-center gap-3 px-4 py-2 rounded-lg hover:bg-slate-700 text-slate-300 hover:text-white",
        "active-class": "bg-slate-700 text-white"
      }, {
        default: withCtx((_, _push2, _parent2, _scopeId) => {
          if (_push2) {
            _push2(`<span${_scopeId}>\xD6versikt</span>`);
          } else {
            return [
              createVNode("span", null, "\xD6versikt")
            ];
          }
        }),
        _: 1
      }, _parent));
      _push(ssrRenderComponent(_component_NuxtLink, {
        to: "/products",
        class: "flex items-center gap-3 px-4 py-2 rounded-lg hover:bg-slate-700 text-slate-300 hover:text-white",
        "active-class": "bg-slate-700 text-white"
      }, {
        default: withCtx((_, _push2, _parent2, _scopeId) => {
          if (_push2) {
            _push2(`<span${_scopeId}>Produkter</span>`);
          } else {
            return [
              createVNode("span", null, "Produkter")
            ];
          }
        }),
        _: 1
      }, _parent));
      _push(ssrRenderComponent(_component_NuxtLink, {
        to: "/listings",
        class: "flex items-center gap-3 px-4 py-2 rounded-lg hover:bg-slate-700 text-slate-300 hover:text-white",
        "active-class": "bg-slate-700 text-white"
      }, {
        default: withCtx((_, _push2, _parent2, _scopeId) => {
          if (_push2) {
            _push2(`<span${_scopeId}>Mina annonser</span>`);
          } else {
            return [
              createVNode("span", null, "Mina annonser")
            ];
          }
        }),
        _: 1
      }, _parent));
      _push(ssrRenderComponent(_component_NuxtLink, {
        to: "/transactions",
        class: "flex items-center gap-3 px-4 py-2 rounded-lg hover:bg-slate-700 text-slate-300 hover:text-white",
        "active-class": "bg-slate-700 text-white"
      }, {
        default: withCtx((_, _push2, _parent2, _scopeId) => {
          if (_push2) {
            _push2(`<span${_scopeId}>Transaktioner</span>`);
          } else {
            return [
              createVNode("span", null, "Transaktioner")
            ];
          }
        }),
        _: 1
      }, _parent));
      _push(ssrRenderComponent(_component_NuxtLink, {
        to: "/analytics",
        class: "flex items-center gap-3 px-4 py-2 rounded-lg hover:bg-slate-700 text-slate-300 hover:text-white",
        "active-class": "bg-slate-700 text-white"
      }, {
        default: withCtx((_, _push2, _parent2, _scopeId) => {
          if (_push2) {
            _push2(`<span${_scopeId}>Marknadsanalys</span>`);
          } else {
            return [
              createVNode("span", null, "Marknadsanalys")
            ];
          }
        }),
        _: 1
      }, _parent));
      _push(ssrRenderComponent(_component_NuxtLink, {
        to: "/scraping",
        class: "flex items-center gap-3 px-4 py-2 rounded-lg hover:bg-slate-700 text-slate-300 hover:text-white",
        "active-class": "bg-slate-700 text-white"
      }, {
        default: withCtx((_, _push2, _parent2, _scopeId) => {
          if (_push2) {
            _push2(`<span${_scopeId}>Scraping</span>`);
          } else {
            return [
              createVNode("span", null, "Scraping")
            ];
          }
        }),
        _: 1
      }, _parent));
      _push(ssrRenderComponent(_component_NuxtLink, {
        to: "/ads",
        class: "flex items-center gap-3 px-4 py-2 rounded-lg hover:bg-slate-700 text-slate-300 hover:text-white",
        "active-class": "bg-slate-700 text-white"
      }, {
        default: withCtx((_, _push2, _parent2, _scopeId) => {
          if (_push2) {
            _push2(`<span${_scopeId}>Hittade annonser</span>`);
          } else {
            return [
              createVNode("span", null, "Hittade annonser")
            ];
          }
        }),
        _: 1
      }, _parent));
      _push(`</nav></aside><main class="ml-64 p-8">`);
      ssrRenderSlot(_ctx.$slots, "default", {}, null, _push, _parent);
      _push(`</main></div>`);
    };
  }
});
const _sfc_setup = _sfc_main.setup;
_sfc_main.setup = (props, ctx) => {
  const ssrContext = useSSRContext();
  (ssrContext.modules || (ssrContext.modules = /* @__PURE__ */ new Set())).add("layouts/default.vue");
  return _sfc_setup ? _sfc_setup(props, ctx) : void 0;
};

export { _sfc_main as default };
//# sourceMappingURL=default-CmS-uDfO.mjs.map
