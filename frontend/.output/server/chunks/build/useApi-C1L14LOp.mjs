import { b as useRuntimeConfig } from './server.mjs';
import { u as useLoadingStore } from './loading--Nv5gQJt.mjs';

const useApi = () => {
  const config = useRuntimeConfig();
  const loadingStore = useLoadingStore();
  const apiBase = config.public.apiBase;
  const fetch = async (endpoint, options) => {
    const url = `${apiBase}/api${endpoint.startsWith("/") ? endpoint : `/${endpoint}`}`;
    loadingStore.startLoading();
    try {
      const result = await $fetch(url, {
        method: (options == null ? void 0 : options.method) || "GET",
        body: options == null ? void 0 : options.body
      });
      return result;
    } finally {
      loadingStore.stopLoading();
    }
  };
  const get = (endpoint) => {
    return fetch(endpoint, { method: "GET" });
  };
  const post = (endpoint, body) => {
    return fetch(endpoint, { method: "POST", body });
  };
  const put = (endpoint, body) => {
    return fetch(endpoint, { method: "PUT", body });
  };
  const patch = (endpoint, body) => {
    return fetch(endpoint, { method: "PATCH", body });
  };
  const del = (endpoint) => {
    return fetch(endpoint, { method: "DELETE" });
  };
  return {
    fetch,
    get,
    post,
    put,
    patch,
    delete: del
  };
};

export { useApi as u };
//# sourceMappingURL=useApi-C1L14LOp.mjs.map
