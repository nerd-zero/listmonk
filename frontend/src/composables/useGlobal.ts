import { getCurrentInstance } from 'vue';
import type Utils from '../utils';
import type eventBus from '../eventBus';

// eslint-disable-next-line import/no-unresolved
type AnyFn = (...args: unknown[]) => unknown;

export interface GlobalProperties {
  $utils: Utils;
  $api: Record<string, AnyFn>;
  $can: (...perms: string[]) => boolean;
  $canList: (id: number, perm: string) => boolean;
  $events: typeof eventBus;
}

export function useGlobal(): GlobalProperties {
  const instance = getCurrentInstance();
  if (!instance) throw new Error('useGlobal must be used inside setup()');
  return instance.appContext.config.globalProperties as unknown as GlobalProperties;
}
