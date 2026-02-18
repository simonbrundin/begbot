import { defineComponent, ref, watchEffect, mergeProps, unref, useSSRContext } from 'vue';
import { ssrRenderAttrs, ssrRenderAttr, ssrInterpolate, ssrIncludeBooleanAttr } from 'vue/server-renderer';
import { d as useSupabaseUser, e as useRouter, u as useNuxtApp } from './server.mjs';
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

const useSupabaseClient = () => {
  return useNuxtApp().$supabase.client;
};
const _sfc_main = /* @__PURE__ */ defineComponent({
  __name: "login",
  __ssrInlineRender: true,
  setup(__props) {
    useSupabaseClient();
    const user = useSupabaseUser();
    const router = useRouter();
    const email = ref("");
    const password = ref("");
    const error = ref("");
    const loading = ref(false);
    watchEffect(() => {
      if (user.value) {
        router.push("/");
      }
    });
    return (_ctx, _push, _parent, _attrs) => {
      _push(`<div${ssrRenderAttrs(mergeProps({ class: "min-h-screen flex items-center justify-center bg-slate-900" }, _attrs))}><div class="max-w-md w-full"><div class="card p-8"><div class="text-center mb-8"><h1 class="text-2xl font-bold text-primary-500">Begbot</h1><p class="text-slate-400 mt-2">Logga in f\xF6r att forts\xE4tta</p></div><form class="space-y-4"><div><label class="label">E-post</label><input${ssrRenderAttr("value", unref(email))} type="email" class="input" placeholder="du@example.com" required></div><div><label class="label">L\xF6senord</label><input${ssrRenderAttr("value", unref(password))} type="password" class="input" placeholder="\u2022\u2022\u2022\u2022\u2022\u2022\u2022\u2022" required></div>`);
      if (unref(error)) {
        _push(`<div class="p-3 bg-red-900/50 text-red-400 rounded-lg text-sm border border-red-800">${ssrInterpolate(unref(error))}</div>`);
      } else {
        _push(`<!---->`);
      }
      _push(`<button type="submit" class="btn btn-primary w-full"${ssrIncludeBooleanAttr(unref(loading)) ? " disabled" : ""}>`);
      if (unref(loading)) {
        _push(`<span>Loggar in...</span>`);
      } else {
        _push(`<span>Logga in</span>`);
      }
      _push(`</button></form><p class="text-center text-sm text-slate-500 mt-6"> Anv\xE4nd dina Supabase-uppgifter </p></div></div></div>`);
    };
  }
});
const _sfc_setup = _sfc_main.setup;
_sfc_main.setup = (props, ctx) => {
  const ssrContext = useSSRContext();
  (ssrContext.modules || (ssrContext.modules = /* @__PURE__ */ new Set())).add("pages/login.vue");
  return _sfc_setup ? _sfc_setup(props, ctx) : void 0;
};

export { _sfc_main as default };
//# sourceMappingURL=login-Ciy6tU9x.mjs.map
