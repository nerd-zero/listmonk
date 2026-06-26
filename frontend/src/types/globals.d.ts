import type Utils from '../utils';

declare module 'vue' {
  interface ComponentCustomProperties {
    $utils: Utils;
    $can: (...perms: string[]) => boolean;
    $canList: (id: number, perm: string) => boolean;
    $events: {
      $on(event: string, handler: (...args: any[]) => void): void;
      $off(event: string, handler?: (...args: any[]) => void): void;
      $emit(event: string, ...args: any[]): void;
    };
    $api: Record<string, (...args: any[]) => any>;
  }
}

export {};
