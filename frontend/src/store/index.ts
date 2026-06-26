import { defineStore } from 'pinia';
import { models, type ModelName } from '../constants';

type LoadingState = Record<ModelName, boolean>;
type ModelState = Record<ModelName, unknown>;

interface MainStoreState extends ModelState {
  loading: LoadingState;
  refreshTick: number;
}

// eslint-disable-next-line import/prefer-default-export
export const useMainStore = defineStore('main', {
  state: (): MainStoreState => ({
    // Model data keyed by model name (lists, campaigns, etc.)
    ...(Object.keys(models).reduce((obj, cur) => ({ ...obj, [cur]: [] }), {}) as ModelState),

    // Per-model loading flags driven by API interceptors.
    loading: Object.keys(models).reduce(
      (obj, cur) => ({ ...obj, [cur]: false }),
      {} as LoadingState,
    ),

    // Incremented by the topbar refresh button; views watch this to re-fetch.
    refreshTick: 0,
  }),

  actions: {
    setModelResponse({ model, data }: { model: ModelName; data: unknown }) {
      (this as unknown as Record<string, unknown>)[model] = data;
    },

    setLoading({ model, status }: { model: ModelName; status: boolean }) {
      this.loading[model] = status;
    },

    refresh() {
      this.refreshTick += 1;
    },
  },
});
