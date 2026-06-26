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
          <PvInputText v-model="data['security.oidc']['provider_url']" name="oidc.provider_url"
            placeholder="https://login.yoursite.com" :disabled="!data['security.oidc']['enabled']" :maxlength="300"
            required type="url" pattern="https?://.*" class="w-full" />
          <div class="quick-links mt-2" :class="{ disabled: !data['security.oidc']['enabled'] }">
            <a href="#" @click.prevent="() => setProvider('google')">Google</a>
            <a href="#" @click.prevent="() => setProvider('microsoft')">Microsoft</a>
            <a href="#" @click.prevent="() => setProvider('apple')">Apple</a>
          </div>
        </div>

        <div class="field">
          <label class="block mb-1 text-sm font-medium">{{ $t('settings.security.OIDCName') }}</label>
          <PvInputText v-model="data['security.oidc']['provider_name']" name="oidc.provider_name"
            :disabled="!data['security.oidc']['enabled']" :maxlength="200" class="w-full" />
        </div>

        <div class="field">
          <label class="block mb-1 text-sm font-medium">{{ $t('settings.security.OIDCClientID') }}</label>
          <PvInputText v-model="data['security.oidc']['client_id']" name="oidc.client_id" ref="clientIdEl"
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
        <p v-if="data['security.oidc']['enabled'] && !isURLOk" class="text-red-500 text-sm mt-1">
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
              <PvRadioButton v-model="selectedProvider" value="altcha" name="captcha_provider" input-id="captcha-altcha" />
              <label for="captcha-altcha">ALTCHA</label>
            </div>
            <div class="flex items-center gap-2">
              <PvRadioButton v-model="selectedProvider" value="hcaptcha" name="captcha_provider" input-id="captcha-hcaptcha" />
              <label for="captcha-hcaptcha">hCaptcha (deprecated)</label>
            </div>
          </div>
        </div>

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
    </div>

    <hr />

    <p class="settings-section-label">{{ $t('settings.security.trustedURLs') }} / CORS</p>
    <div class="field">
      <PvTextarea v-model="trustedURLs" name="trusted_urls" rows="5"
        placeholder="https://example.com" class="w-full" />
      <small class="block mt-1 text-color-secondary">{{ $t('settings.security.trustedURLsHelp') }}</small>
    </div>
  </div>
</template>

<script setup lang="ts">
import {
  ref, computed, nextTick, onMounted,
} from 'vue';
import { storeToRefs } from 'pinia';
import { useI18n } from 'vue-i18n';
import { useMainStore } from '../../store';
import { useGlobal } from '../../composables/useGlobal';
import CopyText from '../../components/CopyText.vue';
import { getRoles } from '../../api/generated/endpoints/roles/roles';

const OIDC_PROVIDERS: Record<string, string> = {
  google: 'https://accounts.google.com',
  github: 'https://token.actions.githubusercontent.com',
  microsoft: 'https://login.microsoftonline.com/{TENANT_HERE}/v2.0',
  apple: 'https://appleid.apple.com',
};

const props = defineProps<{ form?: any }>();
const { $can } = useGlobal();
const { listUserRoles, listListRoles } = getRoles();
const { t } = useI18n();
const store = useMainStore();
const { serverConfig, userRoles, listRoles } = storeToRefs(store);
const clientIdEl = ref<any>(null);
const data = props.form;

const listRoleOptions = computed(() => [
  { id: null, name: `— ${t('globals.terms.none')} —` },
  ...listRoles.value,
]);

const trustedURLs = computed({
  get() {
    const domains = data['security.trusted_urls'];
    return domains && Array.isArray(domains) ? domains.join('\n') : '';
  },
  set(value: string) {
    data['security.trusted_urls'] = value.split('\n');
  },
});

const captchaEnabled = computed({
  get() {
    return data['security.captcha'].altcha.enabled || data['security.captcha'].hcaptcha.enabled;
  },
  set(value: boolean) {
    data['security.captcha'].altcha.enabled = !!value;
    data['security.captcha'].hcaptcha.enabled = false;
  },
});

const selectedProvider = computed({
  get() {
    return data['security.captcha'].hcaptcha.enabled ? 'hcaptcha' : 'altcha';
  },
  set(value: string) {
    data['security.captcha'].hcaptcha.enabled = value === 'hcaptcha';
    data['security.captcha'].altcha.enabled = value === 'altcha';
  },
});

const isURLOk = computed(() => {
  try {
    const u = new URL(serverConfig.value.root_url);
    return u.hostname !== 'localhost' && u.hostname !== '127.0.0.1';
  } catch (e) {
    return false;
  }
});

onMounted(() => {
  if ($can('roles:get')) {
    listUserRoles().then((res: any) => { store.setModelResponse({ model: 'userRoles', data: res }); });
    listListRoles().then((res: any) => { store.setModelResponse({ model: 'listRoles', data: res }); });
  }
});

function setProvider(provider: string) {
  data['security.oidc'].provider_url = OIDC_PROVIDERS[provider];
  data['security.oidc'].provider_name = provider.charAt(0).toUpperCase() + provider.slice(1);
  nextTick(() => { clientIdEl.value?.$el?.focus(); });
}
</script>
