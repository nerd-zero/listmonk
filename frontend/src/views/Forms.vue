<template>
  <div class="forms-page">
    <div class="page-header">
      <h1 class="page-title">{{ $t('forms.title') }}</h1>
    </div>

    <div v-if="loading.lists" class="flex justify-center p-8">
      <PvProgressSpinner style="width:2rem;height:2rem" />
    </div>
    <div v-else-if="publicLists.length === 0" class="empty-notice">
      {{ $t('forms.noPublicLists') }}
    </div>
    <div v-else class="forms-layout">
      <!-- Left panel: controls -->
      <div class="forms-panel table-card">
        <div class="panel-section">
          <h4 class="panel-title">{{ $t('forms.publicLists') }}</h4>
          <p class="panel-help">{{ $t('forms.selectHelp') }}</p>
          <ul class="check-list" data-cy="lists">
            <li v-for="(l, i) in publicLists" :key="l.id" class="check-item">
              <PvCheckbox v-model="checked" :value="i" :input-id="`list-${l.id}`" />
              <label :for="`list-${l.id}`" class="check-label">{{ l.name }}</label>
            </li>
          </ul>
        </div>

        <template v-if="serverConfig.public_subscription.enabled">
          <div class="panel-section">
            <h4 class="panel-title">{{ $t('forms.publicSubPage') }}</h4>
            <a :href="`${serverConfig.root_url}/subscription/form`" target="_blank" rel="noopener noreferer"
              class="sub-url" data-cy="url">
              <i class="pi pi-external-link" />
              {{ serverConfig.root_url }}/subscription/form
            </a>
          </div>
        </template>

        <div class="panel-section">
          <h4 class="panel-title">{{ $t('forms.redirectURL') }}</h4>
          <p class="panel-help">{{ $t('forms.redirectURLHelp') }}</p>
          <ul v-if="redirectURLs.length > 0" class="check-list" data-cy="redirect-urls">
            <li class="check-item">
              <PvRadioButton v-model="selectedRedirectURL" value="" input-id="redirect-none" />
              <label for="redirect-none" class="check-label">{{ $t('globals.terms.none') }}</label>
            </li>
            <li v-for="url in redirectURLs" :key="url" class="check-item">
              <PvRadioButton v-model="selectedRedirectURL" :value="url" :input-id="`redirect-${url}`" />
              <label :for="`redirect-${url}`" class="check-label">{{ url }}</label>
            </li>
          </ul>
        </div>
      </div>

      <!-- Right panel: generated HTML -->
      <div class="forms-output table-card" data-cy="form">
        <div class="panel-section">
          <h4 class="panel-title">{{ $t('forms.formHTML') }}</h4>
          <p class="panel-help">{{ $t('forms.formHTMLHelp') }}</p>
          <div v-if="checked.length === 0" class="output-placeholder">
            <i class="pi pi-code" />
            <span>Select lists to generate the form HTML</span>
          </div>
          <code-editor v-else lang="html" v-model="html" disabled />
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { mapState } from 'pinia';
import { useMainStore } from '../store';
import CodeEditor from '../components/CodeEditor.vue';

export default {
  name: 'ListForm',

  components: {
    'code-editor': CodeEditor,
  },

  data() {
    return {
      checked: [],
      html: '',
      selectedRedirectURL: '',
    };
  },

  methods: {
    escapeAttr(value) {
      return String(value)
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#39;');
    },

    renderHTML() {
      let h = `<form method="post" action="${this.serverConfig.root_url}/subscription/form" class="listmonk-form">\n`
        + '  <div>\n'
        + `    <h3>${this.$t('public.sub')}</h3>\n`
        + '    <input type="hidden" name="nonce" />\n';

      if (this.selectedRedirectURL) {
        h += `    <input type="hidden" name="next" value="${this.escapeAttr(this.selectedRedirectURL)}" />\n`;
      }

      h += '\n'
        + `    <p><input type="email" name="email" required placeholder="${this.$t('subscribers.email')}" /></p>\n`
        + `    <p><input type="text" name="name" placeholder="${this.$t('public.subName')}" /></p>\n\n`;

      this.checked.forEach((i) => {
        const l = this.publicLists[parseInt(i, 10)];

        h += '    <p>\n'
          + `      <input id="${l.uuid.substr(0, 5)}" type="checkbox" name="l" checked value="${l.uuid}" />\n`
          + `      <label for="${l.uuid.substr(0, 5)}">${l.name}</label>\n`;

        if (l.description) {
          h += '      <br />\n'
            + `      <span>${l.description}</span>\n`;
        }

        h += '    </p>\n';
      });

      // Captcha?
      if (this.serverConfig.public_subscription.captcha_enabled) {
        if (this.serverConfig.public_subscription.captcha_provider === 'altcha') {
          h += '\n'
            + `    <altcha-widget challengeurl="${this.serverConfig.root_url}/api/public/captcha/altcha"></altcha-widget>\n`
            + `    <${'script'} type="module" src="${this.serverConfig.root_url}/public/static/altcha.umd.js" async defer></${'script'}>\n`;
        } else if (this.serverConfig.public_subscription.captcha_provider === 'hcaptcha') {
          h += '\n'
            + `    <div class="h-captcha" data-sitekey="${this.serverConfig.public_subscription.captcha_key}"></div>\n`
            + `    <${'script'} src="https://js.hcaptcha.com/1/api.js" async defer></${'script'}>\n`;
        }
      }

      h += '\n'
        + `    <input type="submit" value="${this.$t('public.sub')} " />\n`
        + '  </div>\n'
        + '</form>';

      this.html = h;
    },
  },

  computed: {
    ...mapState(useMainStore, ['loading', 'lists', 'serverConfig']),

    publicLists() {
      if (!this.lists.results) {
        return [];
      }
      return this.lists.results.filter((l) => l.type === 'public');
    },

    redirectURLs() {
      const urls = this.serverConfig.public_subscription
        ? this.serverConfig.public_subscription.redirect_urls
        : [];
      return Array.isArray(urls) ? urls : [];
    },
  },

  watch: {
    checked() {
      this.renderHTML();
    },

    selectedRedirectURL() {
      this.renderHTML();
    },
  },
};
</script>

<style scoped lang="scss">
.forms-page {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}
.page-header { display: flex; align-items: center; }
.page-title {
  font-size: 1.5rem;
  font-weight: 700;
  color: #0f172a;
  margin: 0;
}
.empty-notice {
  padding: 2rem;
  color: #94a3b8;
  text-align: center;
}

.forms-layout {
  display: grid;
  grid-template-columns: 320px 1fr;
  gap: 1.25rem;
  align-items: start;

  @media (max-width: 900px) { grid-template-columns: 1fr; }
}

.table-card {
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
  overflow: hidden;
}
.forms-panel, .forms-output { padding: 0; }

.panel-section {
  padding: 1.25rem 1.5rem;
  & + & { border-top: 1px solid #f1f5f9; }
}
.panel-title {
  font-size: 0.85rem;
  font-weight: 700;
  color: #0f172a;
  margin: 0 0 0.4rem;
}
.panel-help {
  font-size: 0.8rem;
  color: #94a3b8;
  margin: 0 0 0.75rem;
  line-height: 1.4;
}

.check-list { list-style: none; padding: 0; margin: 0; display: flex; flex-direction: column; gap: 0.5rem; }
.check-item { display: flex; align-items: center; gap: 0.6rem; }
.check-label { font-size: 0.875rem; color: #374151; cursor: pointer; }

.sub-url {
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  font-size: 0.8rem;
  color: #3b82f6;
  text-decoration: none;
  word-break: break-all;
  &:hover { text-decoration: underline; }
}

.output-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.75rem;
  padding: 3rem 1rem;
  color: #94a3b8;
  font-size: 0.875rem;
  i { font-size: 2rem; }
}
</style>
