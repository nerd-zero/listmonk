// Singleton wrapper so non-component code (utils.ts, api/index.ts) can trigger
// PrimeVue toasts without needing access to the component instance.
// App.vue calls setInstance() once the useToast() composable is available.

interface ToastInstance {
  add(opts: { severity: string; detail: string; life: number }): void;
}

interface ConfirmInstance {
  require(opts: {
    message: string;
    header: string;
    icon: string;
    accept: (() => void) | undefined;
    reject: (() => void) | undefined;
    rejectProps: Record<string, unknown>;
  }): void;
}

let toastInstance: ToastInstance | null = null;
let confirmInstance: ConfirmInstance | null = null;

export function setToastInstance(instance: ToastInstance): void {
  toastInstance = instance;
}

export function setConfirmInstance(instance: ConfirmInstance): void {
  confirmInstance = instance;
}

// Map Buefy type strings to PrimeVue severity strings.
function toSeverity(typ: string): string {
  const map: Record<string, string> = {
    'is-success': 'success',
    'is-danger': 'error',
    'is-warning': 'warn',
    'is-info': 'info',
  };
  return map[typ] || 'success';
}

export function showToast(msg: string, typ: string, duration: number): void {
  if (toastInstance) {
    toastInstance.add({
      severity: toSeverity(typ),
      detail: msg,
      life: duration || 3000,
    });
  }
}

export function showConfirm(
  msg: string,
  onConfirm?: () => void,
  onCancel?: () => void,
): void {
  if (confirmInstance) {
    confirmInstance.require({
      message: msg,
      header: 'Confirm',
      icon: 'pi pi-exclamation-triangle',
      accept: onConfirm,
      reject: onCancel,
      rejectProps: { severity: 'secondary', outlined: true },
    });
    return;
  }
  // Fallback if confirm service not yet wired.
  // eslint-disable-next-line no-alert
  if (window.confirm(msg)) {
    onConfirm?.();
  } else {
    onCancel?.();
  }
}

export function showPrompt(
  msg: string,
  onConfirm?: (value: string) => void,
  onCancel?: () => void,
): void {
  // PrimeVue doesn't have a built-in prompt dialog; use browser prompt as fallback.
  // eslint-disable-next-line no-alert
  const value = window.prompt(msg);
  if (value !== null) {
    onConfirm?.(value);
  } else {
    onCancel?.();
  }
}
