<template>
  <div class="flex flex-column gap-4">
    <h1 class="text-3xl font-bold text-900 m-0">{{ $t('forms.title') }}</h1>

    <div v-if="loading.lists" class="flex justify-content-center p-8">
      <PvProgressSpinner style="width:2rem;height:2rem" />
    </div>

    <PvMessage v-else-if="publicLists.length === 0" severity="info" :closable="false">
      {{ $t('forms.noPublicLists') }}
    </PvMessage>

    <div v-else class="grid align-items-start">
      <!-- Left panel: controls -->
      <div class="col-12 md:col-4" data-cy="lists">
        <PvCard>
          <template #title>{{ $t('forms.publicLists') }}</template>
          <template #subtitle>{{ $t('forms.selectHelp') }}</template>
          <template #content>
            <div class="flex flex-column gap-3">
              <div v-for="(l, i) in publicLists" :key="l.id" class="flex align-items-center gap-2">
                <PvCheckbox v-model="checked" :value="i" :input-id="`list-${l.id}`" />
                <label :for="`list-${l.id}`" class="cursor-pointer">{{ l.name }}</label>
              </div>
            </div>

            <template v-if="serverConfig.public_subscription?.enabled">
              <PvDivider />
              <p class="font-semibold text-sm text-900 mb-2">{{ $t('forms.publicSubPage') }}</p>
              <a :href="`${serverConfig.root_url}/subscription/form`" target="_blank" rel="noopener noreferer"
                class="text-primary text-sm flex align-items-center gap-1 no-underline hover:underline" data-cy="url">
                <i class="pi pi-external-link" />
                {{ serverConfig.root_url }}/subscription/form
              </a>
            </template>

            <template v-if="redirectURLs.length > 0">
              <PvDivider />
              <p class="font-semibold text-sm text-900 mb-1">{{ $t('forms.redirectURL') }}</p>
              <p class="text-sm text-500 mb-3">{{ $t('forms.redirectURLHelp') }}</p>
              <div class="flex flex-column gap-3" data-cy="redirect-urls">
                <div class="flex align-items-center gap-2">
                  <PvRadioButton v-model="selectedRedirectURL" value="" input-id="redirect-none" />
                  <label for="redirect-none" class="cursor-pointer">{{ $t('globals.terms.none') }}</label>
                </div>
                <div v-for="url in redirectURLs" :key="url" class="flex align-items-center gap-2">
                  <PvRadioButton v-model="selectedRedirectURL" :value="url" :input-id="`redirect-${url}`" />
                  <label :for="`redirect-${url}`" class="cursor-pointer text-sm">{{ url }}</label>
                </div>
              </div>
            </template>
          </template>
        </PvCard>
      </div>

      <!-- Right panel: generated HTML -->
      <div class="col-12 md:col-8" data-cy="form">
        <PvCard>
          <template #title>{{ $t('forms.formHTML') }}</template>
          <template #subtitle>{{ $t('forms.formHTMLHelp') }}</template>
          <template #content>
            <div v-if="checked.length === 0"
              class="flex flex-column align-items-center justify-content-center gap-3 py-6 text-500">
              <i class="pi pi-code" style="font-size:2.5rem" />
              <span class="text-sm">{{ $t('forms.selectHelp') }}</span>
            </div>
            <code-editor v-else lang="html" v-model="html" disabled />
          </template>
        </PvCard>
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
