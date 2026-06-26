import { describe, it, expect, beforeEach } from 'vitest';
import { createPinia, setActivePinia } from 'pinia';
import { useMainStore } from './index';
import { models } from '../constants';

beforeEach(() => {
  setActivePinia(createPinia());
});

describe('useMainStore — initial state', () => {
  it('initialises every model key to an empty array', () => {
    const store = useMainStore();
    for (const key of Object.keys(models)) {
      expect((store as any)[key]).toEqual([]);
    }
  });

  it('initialises every loading flag to false', () => {
    const store = useMainStore();
    for (const key of Object.keys(models)) {
      expect(store.loading[key as keyof typeof models]).toBe(false);
    }
  });

  it('initialises refreshTick to 0', () => {
    expect(useMainStore().refreshTick).toBe(0);
  });
});

describe('useMainStore — setModelResponse', () => {
  it('stores data under the given model key', () => {
    const store = useMainStore();
    const payload = [{ id: 1, name: 'Newsletter' }];
    store.setModelResponse({ model: 'lists', data: payload });
    expect((store as any).lists).toEqual(payload);
  });

  it('replaces existing model data', () => {
    const store = useMainStore();
    store.setModelResponse({ model: 'campaigns', data: [{ id: 1 }] });
    store.setModelResponse({ model: 'campaigns', data: [{ id: 2 }, { id: 3 }] });
    expect((store as any).campaigns).toEqual([{ id: 2 }, { id: 3 }]);
  });

  it('does not affect other model keys', () => {
    const store = useMainStore();
    store.setModelResponse({ model: 'lists', data: [{ id: 1 }] });
    expect((store as any).campaigns).toEqual([]);
  });

  it('accepts a plain object as data', () => {
    const store = useMainStore();
    const config = { version: '2.0', rootUrl: 'https://example.com' };
    store.setModelResponse({ model: 'serverConfig', data: config });
    expect((store as any).serverConfig).toEqual(config);
  });
});

describe('useMainStore — setLoading', () => {
  it('sets the loading flag for the given model to true', () => {
    const store = useMainStore();
    store.setLoading({ model: 'subscribers', status: true });
    expect(store.loading.subscribers).toBe(true);
  });

  it('clears the loading flag back to false', () => {
    const store = useMainStore();
    store.setLoading({ model: 'subscribers', status: true });
    store.setLoading({ model: 'subscribers', status: false });
    expect(store.loading.subscribers).toBe(false);
  });

  it('does not affect loading flags for other models', () => {
    const store = useMainStore();
    store.setLoading({ model: 'campaigns', status: true });
    expect(store.loading.lists).toBe(false);
    expect(store.loading.subscribers).toBe(false);
  });
});

describe('useMainStore — refresh', () => {
  it('increments refreshTick by 1 each call', () => {
    const store = useMainStore();
    store.refresh();
    expect(store.refreshTick).toBe(1);
    store.refresh();
    expect(store.refreshTick).toBe(2);
  });
});
