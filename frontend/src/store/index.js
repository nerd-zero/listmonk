import { defineStore } from 'pinia';
import { models } from '../constants';

// eslint-disable-next-line import/prefer-default-export
export const useMainStore = defineStore('main', {
  state: () => ({
    // Model data keyed by model name (lists, campaigns, etc.)
    ...Object.keys(models).reduce((obj, cur) => ({ ...obj, [cur]: [] }), {}),

    // Per-model loading flags driven by API interceptors.
    loading: Object.keys(models).reduce((obj, cur) => ({ ...obj, [cur]: false }), {}),

    // Incremented by the topbar refresh button; views watch this to re-fetch.
    refreshTick: 0,
  }),

  actions: {
    setModelResponse({ model, data }) {
      this[model] = data;
    },

    setLoading({ model, status }) {
      this.loading[model] = status;
    },

    refresh() {
      this.refreshTick += 1;
    },
  },
});
