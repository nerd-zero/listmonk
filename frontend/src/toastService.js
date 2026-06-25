// Singleton wrapper so non-component code (utils.js, api/index.js) can trigger
// PrimeVue toasts without needing access to the component instance.
// App.vue calls setInstance() once the useToast() composable is available.

let toastInstance = null;
let confirmInstance = null;

export function setToastInstance(instance) {
  toastInstance = instance;
}

export function setConfirmInstance(instance) {
  confirmInstance = instance;
}

// Map Buefy type strings to PrimeVue severity strings.
function toSeverity(typ) {
  const map = {
    'is-success': 'success',
    'is-danger': 'error',
    'is-warning': 'warn',
    'is-info': 'info',
  };
  return map[typ] || 'success';
}

export function showToast(msg, typ, duration) {
  if (toastInstance) {
    toastInstance.add({
      severity: toSeverity(typ),
      detail: msg,
      life: duration || 3000,
    });
  }
}

export function showConfirm(msg, onConfirm, onCancel) {
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

export function showPrompt(msg, onConfirm, onCancel) {
  // PrimeVue doesn't have a built-in prompt dialog; use browser prompt as fallback.
  // eslint-disable-next-line no-alert
  const value = window.prompt(msg);
  if (value !== null) {
    onConfirm?.(value);
  } else {
    onCancel?.();
  }
}
