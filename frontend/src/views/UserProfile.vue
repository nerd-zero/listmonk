<template>
  <section class="user-profile">
    <div v-if="loading.users" class="flex justify-center p-8">
      <PvProgressSpinner />
    </div>

    <div class="page-header">
      <div class="page-header-left">
        <h1 class="page-title">@{{ data.username }}</h1>
        <PvTag v-if="data.userRole" severity="secondary">{{ data.userRole.name }}</PvTag>
      </div>
      <PvButton severity="primary" icon="pi pi-save" type="submit" form="profile-form"
        data-cy="btn-save" :label="$t('globals.buttons.save')" />
    </div>

    <form id="profile-form" @submit.prevent="onSubmit" class="profile-form">
      <div class="settings-card">
        <div v-if="data.type !== 'api'" class="field">
          <label class="block mb-1 text-sm font-medium">{{ $t('subscribers.email') }}</label>
          <PvInputText :maxlength="200" v-model="form.email" name="email"
            :placeholder="$t('subscribers.email')" :disabled="!data.passwordLogin"
            required autofocus class="w-full" />
        </div>

        <div class="field">
          <label class="block mb-1 text-sm font-medium">{{ $t('globals.fields.name') }}</label>
          <PvInputText :maxlength="200" v-model="form.name" name="name"
            :placeholder="$t('globals.fields.name')" class="w-full" />
        </div>

        <div v-if="data.passwordLogin" class="grid">
          <div class="col-6">
            <div class="field">
              <label class="block mb-1 text-sm font-medium">{{ $t('users.password') }}</label>
              <PvPassword minlength="8" :maxlength="200" v-model="form.password" name="password"
                :placeholder="$t('users.password')" :feedback="false" class="w-full" />
            </div>
          </div>
          <div class="col-6">
            <div class="field">
              <label class="block mb-1 text-sm font-medium">{{ $t('users.passwordRepeat') }}</label>
              <PvPassword minlength="8" :maxlength="200" v-model="form.password2" name="password2"
                :feedback="false" class="w-full" />
            </div>
          </div>
        </div>
      </div>

      <!-- 2FA -->
      <section v-if="data.passwordLogin">
        <!-- TOTP disabled -->
        <div v-if="data.twofaType === 'none'" class="settings-card">
          <div class="twofa-header">
            <div>
              <p class="settings-section-label">{{ $t('users.twoFA') }}</p>
              <p class="text-color-secondary text-sm">{{ $t('users.twoFANotEnabled') }}</p>
            </div>
            <PvToggleSwitch v-if="!isTotpVisible" v-model="twofaEnabled" @change="onToggleEnableTotp" />
          </div>

          <div v-if="isTotpVisible" class="totp-setup">
            <div v-if="totpQR">
              <p class="text-color-secondary text-sm mb-3">{{ $t('users.totpScanQR') }}</p>
              <img :src="'data:image/png;base64,' + totpQR" alt="QR Code" class="totp-qr" />
              <p class="mt-3 mb-4">
                <strong>{{ $t('users.totpSecret') }}</strong><br />
                <code><copy-text :text="`${totpSecret}`" /></code>
              </p>
              <form @submit.prevent="confirmTOTP">
                <div class="field">
                  <label class="block mb-1 text-sm font-medium">{{ $t('users.totpCode') }}</label>
                  <PvInputText ref="totpCodeInputEl" v-model="totpCode" maxlength="6"
                    pattern="[0-9]{6}" placeholder="000000" required class="w-full" />
                </div>
                <div class="flex gap-2 mt-3">
                  <PvButton severity="primary" type="submit" :label="$t('globals.buttons.enable')" />
                  <PvButton severity="secondary" outlined type="button" @click="onCancelTOTPSetup"
                    :label="$t('globals.buttons.cancel')" />
                </div>
              </form>
            </div>
          </div>
        </div>

        <!-- TOTP enabled -->
        <div v-if="data.twofaType === 'totp'" class="settings-card">
          <div class="twofa-header">
            <div>
              <p class="settings-section-label">
                <i class="pi pi-check-circle text-green-500 mr-1" />{{ $t('users.twoFAEnabled') }}
              </p>
              <p class="text-color-secondary text-sm">
                {{ $t('users.twoFAEnabledDesc', { type: data.twofaType.toUpperCase() }) }}
              </p>
            </div>
            <PvToggleSwitch v-if="!showDisableTOTP" v-model="twofaEnabled" @change="toggleDisableTOTP" />
          </div>

          <form v-if="showDisableTOTP" class="mt-4" @submit.prevent="confirmDisableTOTP">
            <div class="field">
              <label class="block mb-1 text-sm font-medium">{{ $t('users.password') }}</label>
              <PvPassword ref="disablePasswordInputEl" v-model="disableTOTPPassword"
                minlength="8" required :feedback="false" class="w-full" />
            </div>
            <div class="flex gap-2 mt-3">
              <PvButton severity="danger" type="submit" :label="$t('globals.buttons.disable')" />
              <PvButton severity="secondary" outlined type="button" @click="onCancelTOTPSetup"
                :label="$t('globals.buttons.cancel')" />
            </div>
          </form>
        </div>
      </section>
    </form>
  </section>
</template>

<script setup lang="ts">
import {
  ref, reactive, nextTick, onMounted,
} from 'vue';
import { storeToRefs } from 'pinia';
import { useI18n } from 'vue-i18n';
import { useMainStore } from '../store';
import { useGlobal } from '../composables/useGlobal';
import CopyText from '../components/CopyText.vue';
import { getUsers as usersApi } from '../api/generated/endpoints/users/users';

const { $utils } = useGlobal();
const {
  getUserProfile, updateUserProfile, generateTotpQr, enableTotp, disableTotp,
} = usersApi();
const { t } = useI18n();
const { loading } = storeToRefs(useMainStore());

const form = reactive<any>({});
const data = ref<any>({});
const isTotpVisible = ref(false);
const totpQR = ref<string | null>(null);
const totpSecret = ref<string | null>(null);
const totpCode = ref('');
const showDisableTOTP = ref(false);
const disableTOTPPassword = ref('');
const twofaEnabled = ref(false);
const totpCodeInputEl = ref<any>(null);
const disablePasswordInputEl = ref<any>(null);

function onSubmit() {
  const params: any = { name: form.name, email: form.email };
  if (data.value.passwordLogin && form.password) {
    if (form.password !== form.password2) {
      $utils.toast(t('users.passwordMismatch'), 'is-danger');
      return;
    }
    params.password = form.password;
    params.password2 = form.password2;
  }
  updateUserProfile(params).then(() => {
    form.password = '';
    form.password2 = '';
    $utils.toast(t('globals.messages.updated', { name: data.value.username }));
  });
}

function onToggleEnableTotp() {
  generateTotpQr(data.value.id).then((d: any) => {
    totpQR.value = d.qr;
    totpSecret.value = d.secret;
    isTotpVisible.value = true;
    nextTick(() => { totpCodeInputEl.value?.focus(); });
  }).catch(() => {
    $utils.toast(t('globals.messages.errorFetching'), 'is-danger');
  });
}

function onCancelTOTPSetup() {
  isTotpVisible.value = false;
  totpQR.value = null;
  totpSecret.value = null;
  totpCode.value = '';
  twofaEnabled.value = data.value.twofaType === 'totp';
  showDisableTOTP.value = false;
  disableTOTPPassword.value = '';
}

function confirmTOTP() {
  if (!totpCode.value || totpCode.value.length !== 6) {
    $utils.toast(t('globals.messages.invalidValue'), 'is-danger');
    return;
  }
  enableTotp(data.value.id, { secret: totpSecret.value!, code: totpCode.value }).then(() => {
    $utils.toast(t('users.twoFAEnabled'));
    onCancelTOTPSetup();
    getUserProfile().then((p: any) => {
      data.value = { ...p };
      twofaEnabled.value = p.twofaType === 'totp';
    });
  }).catch(() => {
    $utils.toast(t('globals.messages.invalidValue'), 'is-danger');
  });
}

function toggleDisableTOTP() {
  showDisableTOTP.value = true;
  nextTick(() => { disablePasswordInputEl.value?.focus(); });
}

function confirmDisableTOTP() {
  if (!disableTOTPPassword.value) {
    $utils.toast(t('globals.messages.invalidFields'), 'is-danger');
    return;
  }
  disableTotp(data.value.id, { password: disableTOTPPassword.value }).then(() => {
    $utils.toast(t('globals.messages.done'));
    showDisableTOTP.value = false;
    disableTOTPPassword.value = '';
    getUserProfile().then((p: any) => {
      data.value = { ...p };
      twofaEnabled.value = p.twofaType === 'totp';
    });
  }).catch(() => {
    $utils.toast(t('users.invalidPassword'), 'is-danger');
  });
}

onMounted(() => {
  getUserProfile().then((d: any) => {
    data.value = { ...d };
    Object.assign(form, { name: d.name, email: d.email });
    twofaEnabled.value = d.twofaType === 'totp';
  });
});
</script>

<style scoped lang="scss">
.user-profile {
  max-width: 640px;
}

.page-header-left {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.profile-form {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;

  .settings-card {
    display: flex;
    flex-direction: column;
    gap: 1rem;

    .field { margin-bottom: 0; }
  }
}

.twofa-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
}

.totp-qr {
  display: block;
  width: 180px;
  height: 180px;
  border: 1px solid var(--lm-border);
  border-radius: 8px;
}

:deep(.p-password) { width: 100%; }
:deep(.p-password-input) { width: 100%; }
</style>
