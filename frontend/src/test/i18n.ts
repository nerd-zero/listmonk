import { createI18n } from 'vue-i18n';

// Minimal i18n instance for unit tests.
// Returns key strings verbatim so assertions can match on i18n keys.
export const testI18n = createI18n({
  legacy: false,
  locale: 'en',
  messages: { en: {} },
  missing: (_locale, key) => key,
});
