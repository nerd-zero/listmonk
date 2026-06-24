const handlers = {};

export default {
  $on(event, handler) {
    if (!handlers[event]) handlers[event] = [];
    handlers[event].push(handler);
  },
  $off(event, handler) {
    if (handlers[event]) {
      handlers[event] = handlers[event].filter((h) => h !== handler);
    }
  },
  $emit(event, ...args) {
    (handlers[event] || []).forEach((h) => h(...args));
  },
};
