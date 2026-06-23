<template>
  <div class="items">
    <div class="grid">
      <div class="col-3">
        <div class="field">
          <div class="flex items-center gap-2">
            <PvToggleSwitch v-model="data['security.oidc']['enabled']" name="security.oidc" />
            <span>{{ $t('settings.security.enableOIDC') }}</span>
          </div>
          <small class="block mt-1 text-color-secondary">{{ $t('settings.security.OIDCHelp') }}</small>
        </div>
      </div>
      <div class="col-9">
        <div class="field">
          <label class="block mb-1 text-sm font-medium">{{ $t('settings.security.OIDCURL') }}</label>
          <div>
            <PvInputText v-model="data['security.oidc']['provider_url']" name="oidc.provider_url"
              placeholder="https://login.yoursite.com" :disabled="!data['security.oidc']['enabled']" :maxlength="300"
              required type="url" pattern="https?://.*" class="w-full" />

            <div class="spaced-links is-size-7 mt-2" :class="{ 'disabled': !data['security.oidc']['enabled'] }">
              <a href="#" @click.prevent="() => setProvider('google')">Google</a>
              <a href="#" @click.prevent="() => setProvider('microsoft')">Microsoft</a>
              <a href="#" @click.prevent="() => setProvider('apple')">Apple</a>
            </div>
          </div>
        </div>

        <div class="field">
          <label class="block mb-1 text-sm font-medium">{{ $t('settings.security.OIDCName') }}</label>
          <PvInputText v-model="data['security.oidc']['provider_name']" name="oidc.provider_name" ref="provider_name"
            :disabled="!data['security.oidc']['enabled']" :maxlength="200" class="w-full" />
        </div>

        <div class="field">
          <label class="block mb-1 text-sm font-medium">{{ $t('settings.security.OIDCClientID') }}</label>
          <PvInputText v-model="data['security.oidc']['client_id']" name="oidc.client_id" ref="client_id"
            :disabled="!data['security.oidc']['enabled']" :maxlength="200" required class="w-full" />
        </div>

        <div class="field">
          <label class="block mb-1 text-sm font-medium">{{ $t('settings.security.OIDCClientSecret') }}</label>
          <PvPassword v-model="data['security.oidc']['client_secret']" name="oidc.client_secret"
            :disabled="!data['security.oidc']['enabled']" :maxlength="200" required :feedback="false" class="w-full" />
        </div>

        <hr />

        <div class="field">
          <div class="flex items-center gap-2">
            <PvToggleSwitch v-model="data['security.oidc']['auto_create_users']" :disabled="!data['security.oidc']['enabled']"
              name="oidc.auto_create_users" />
            <span>{{ $t('settings.security.OIDCAutoCreateUsers') }}</span>
          </div>
          <small class="block mt-1 text-color-secondary">{{ $t('settings.security.OIDCAutoCreateUsersHelp') }}</small>
        </div>

        <div class="field">
          <label class="block mb-1 text-sm font-medium">{{ $t('settings.security.OIDCDefaultUserRole') }}</label>
          <PvSelect v-model="data['security.oidc']['default_user_role_id']"
            :disabled="!data['security.oidc']['enabled'] || !data['security.oidc']['auto_create_users']"
            name="oidc.default_user_role_id"
            :options="userRoles" option-label="name" option-value="id" class="w-full" />
          <small class="block mt-1 text-color-secondary">{{ $t('settings.security.OIDCDefaultRoleHelp') }}</small>
        </div>

        <div class="field">
          <label class="block mb-1 text-sm font-medium">{{ $t('settings.security.OIDCDefaultListRole') }}</label>
          <PvSelect v-model="data['security.oidc']['default_list_role_id']"
            :disabled="!data['security.oidc']['enabled'] || !data['security.oidc']['auto_create_users']"
            name="oidc.default_list_role_id"
            :options="listRoleOptions" option-label="name" option-value="id" class="w-full" />
          <small class="block mt-1 text-color-secondary">{{ $t('settings.security.OIDCDefaultRoleHelp') }}</small>
        </div>

        <hr />

        <div class="field">
          <label class="block mb-1 text-sm font-medium">{{ $t('settings.security.OIDCRedirectURL') }}</label>
          <code><copy-text :text="`${serverConfig.root_url}/auth/oidc`" /></code>
        </div>
        <p v-if="data['security.oidc']['enabled'] && !isURLOk" class="has-text-danger">
          <i class="pi pi-exclamation-triangle" />
          {{ $t('settings.security.OIDCRedirectWarning') }}
        </p>
      </div>
    </div>

    <hr />
    <div class="grid">
      <div class="col-3">
        <div class="field">
          <div class="flex items-center gap-2">
            <PvToggleSwitch v-model="captchaEnabled" name="security.captcha" />
            <span>{{ $t('settings.security.enableCaptcha') }}</span>
          </div>
          <small class="block mt-1 text-color-secondary">{{ $t('settings.security.enableCaptchaHelp') }}</small>
        </div>
      </div>
      <div class="col-9" v-if="captchaEnabled">
        <div class="field">
          <div class="flex items-center gap-4">
            <div class="flex items-center gap-2">
              <input type="radio" v-model="selectedProvider" value="altcha" name="captcha_provider" id="captcha-altcha" />
              <label for="captcha-altcha">ALTCHA</label>
            </div>
            <div class="flex items-center gap-2">
              <input type="radio" v-model="selectedProvider" value="hcaptcha" name="captcha_provider" id="captcha-hcaptcha" />
              <label for="captcha-hcaptcha">hCaptcha (deprecated)</label>
            </div>
          </div>
        </div>

        <!-- captcha settings -->
        <div v-if="selectedProvider === 'altcha'">
          <div class="field">
            <label class="block mb-1 text-sm font-medium">{{ $t('settings.security.altchaComplexity') }}</label>
            <PvInputNumber v-model="data['security.captcha']['altcha']['complexity']" name="altcha_complexity"
              :min="1000" :max="1000000" required class="w-full" />
            <small class="block mt-1 text-color-secondary">{{ $t('settings.security.altchaComplexityHelp') }}</small>
          </div>
        </div>
        <div v-if="selectedProvider === 'hcaptcha'">
          <div class="field">
            <label class="block mb-1 text-sm font-medium">{{ $t('settings.security.captchaKey') }}</label>
            <PvInputText v-model="data['security.captcha']['hcaptcha']['key']" name="hcaptcha_key" :maxlength="200"
              required class="w-full" />
            <small class="block mt-1 text-color-secondary">{{ $t('settings.security.captchaKeyHelp') }}</small>
          </div>
          <div class="field">
            <label class="block mb-1 text-sm font-medium">{{ $t('settings.security.captchaSecret') }}</label>
            <PvPassword v-model="data['security.captcha']['hcaptcha']['secret']" name="hcaptcha_secret"
              :maxlength="200" required :feedback="false" class="w-full" />
          </div>
        </div>
      </div>
    </div><!-- captcha -->

    <hr />

    <!-- CORS -->
    <div class="grid">
      <div class="col-12">
        <h3 class="is-size-6"><strong>{{ $t('settings.security.trustedURLs') }} / CORS</strong></h3><br />
        <div class="field">
          <PvTextarea v-model="trustedURLs" name="trusted_urls" rows="5"
            placeholder="https://example.com" class="w-full" />
          <small class="block mt-1 text-color-secondary">{{ $t('settings.security.trustedURLsHelp') }}</small>
        </div>
      </div>
    </div><!-- cors -->
  </div>
</template>

<script>
import { mapState } from 'vuex';
import CopyText from '../../components/CopyText.vue';

const OIDC_PROVIDERS = {
  google: 'https://accounts.google.com',
  github: 'https://token.actions.githubusercontent.com',
  microsoft: 'https://login.microsoftonline.com/{TENANT_HERE}/v2.0',
  apple: 'https://appleid.apple.com',
};

export default {
  components: {
    CopyText,
  },

  props: {
    form: {
      type: Object, default: () => { },
    },
  },

  computed: {
    ...mapState(['serverConfig', 'userRoles', 'listRoles']),

    listRoleOptions() {
      return [{ id: null, name: `— ${this.$t('globals.terms.none')} —` }, ...this.listRoles];
    },

    trustedURLs: {
      get() {
        // Convert array to newline-separated string.
        const domains = this.data['security.trusted_urls'];
        return domains && Array.isArray(domains) ? domains.join('\n') : '';
      },
      set(value) {
        this.data['security.trusted_urls'] = value.split('\n');
      },
    },

    captchaEnabled: {
      get() {
        return this.data['security.captcha'].altcha.enabled || this.data['security.captcha'].hcaptcha.enabled;
      },
      set(value) {
        this.data['security.captcha'].altcha.enabled = !!value;
        this.data['security.captcha'].hcaptcha.enabled = false;
      },
    },

    selectedProvider: {
      get() {
        if (this.data['security.captcha'].hcaptcha.enabled) {
          return 'hcaptcha';
        }

        return 'altcha';
      },
      set(value) {
        this.data['security.captcha'].hcaptcha.enabled = value === 'hcaptcha';
        this.data['security.captcha'].altcha.enabled = value === 'altcha';
      },
    },

    version() {
      return import.meta.env.VUE_APP_VERSION;
    },

    isMobile() {
      return this.windowWidth <= 768;
    },

    isURLOk() {
      try {
        const u = new URL(this.serverConfig.root_url);
        return u.hostname !== 'localhost' && u.hostname !== '127.0.0.1';
      } catch (e) {
        return false;
      }
    },
  },

  mounted() {
    if (this.$can('roles:get')) {
      this.$api.getUserRoles();
      this.$api.getListRoles();
    }
  },

  methods: {
    setProvider(provider) {
      this.data['security.oidc'].provider_url = OIDC_PROVIDERS[provider];
      this.data['security.oidc'].provider_name = provider.charAt(0).toUpperCase() + provider.slice(1);

      this.$nextTick(() => {
        this.$refs.client_id.$el.focus();
      });
    },
  },

  data() {
    return {
      data: this.form,
    };
  },
};
</script>
