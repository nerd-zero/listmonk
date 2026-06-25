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
                  <PvInputText ref="totpCodeInput" v-model="totpCode" maxlength="6"
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
              <PvPassword ref="disablePasswordInput" v-model="disableTOTPPassword"
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

<script lang="ts">
import { defineComponent } from 'vue';
import { mapState } from 'pinia';
import { useMainStore } from '../store';
import CopyText from '../components/CopyText.vue';

export default defineComponent({
  name: 'UserProfile',

  components: { CopyText },

  data() {
    return {
      form: {},
      data: {},
      isTotpVisible: false,
      totpQR: null,
      totpSecret: null,
      totpCode: '',
      showDisableTOTP: false,
      disableTOTPPassword: '',
      twofaEnabled: false,
    };
  },

  methods: {
    onSubmit() {
      const params = {
        name: this.form.name,
        email: this.form.email,
      };

      if (this.data.passwordLogin && this.form.password) {
        if (this.form.password !== this.form.password2) {
          this.$utils.toast(this.$t('users.passwordMismatch'), 'is-danger');
          return;
        }
        params.password = this.form.password;
        params.password2 = this.form.password2;
      }

      this.$api.updateUserProfile(params).then(() => {
        this.form.password = '';
        this.form.password2 = '';
        this.$utils.toast(this.$t('globals.messages.updated', { name: this.data.username }));
      });
    },

    onToggleEnableTotp() {
      this.$api.getTOTPQR(this.data.id).then((data) => {
        this.totpQR = data.qr;
        this.totpSecret = data.secret;
        this.isTotpVisible = true;
        this.$nextTick(() => { this.$refs.totpCodeInput?.focus(); });
      }).catch(() => {
        this.$utils.toast(this.$t('globals.messages.errorFetching'), 'is-danger');
      });
    },

    onCancelTOTPSetup() {
      this.isTotpVisible = false;
      this.totpQR = null;
      this.totpSecret = null;
      this.totpCode = '';
      this.twofaEnabled = this.data.twofaType === 'totp';
      this.showDisableTOTP = false;
      this.disableTOTPPassword = '';
    },

    confirmTOTP() {
      if (!this.totpCode || this.totpCode.length !== 6) {
        this.$utils.toast(this.$t('globals.messages.invalidValue'), 'is-danger');
        return;
      }
      const d = new FormData();
      d.append('secret', this.totpSecret);
      d.append('code', this.totpCode);
      this.$api.enableTOTP(this.data.id, d).then(() => {
        this.$utils.toast(this.$t('users.twoFAEnabled'));
        this.onCancelTOTPSetup();
        this.$api.getUserProfile().then((data) => {
          this.data = { ...data };
          this.twofaEnabled = data.twofaType === 'totp';
        });
      }).catch(() => {
        this.$utils.toast(this.$t('globals.messages.invalidValue'), 'is-danger');
      });
    },

    toggleDisableTOTP() {
      this.showDisableTOTP = true;
      this.$nextTick(() => { this.$refs.disablePasswordInput?.focus(); });
    },

    confirmDisableTOTP() {
      if (!this.disableTOTPPassword) {
        this.$utils.toast(this.$t('globals.messages.invalidFields'), 'is-danger');
        return;
      }
      const formData = new FormData();
      formData.append('password', this.disableTOTPPassword);
      this.$api.disableTOTP(this.data.id, formData).then(() => {
        this.$utils.toast(this.$t('globals.messages.done'));
        this.showDisableTOTP = false;
        this.disableTOTPPassword = '';
        this.$api.getUserProfile().then((data) => {
          this.data = { ...data };
          this.twofaEnabled = data.twofaType === 'totp';
        });
      }).catch(() => {
        this.$utils.toast(this.$t('users.invalidPassword'), 'is-danger');
      });
    },
  },

  mounted() {
    this.$api.getUserProfile().then((data) => {
      this.data = { ...data };
      this.form = { name: data.name, email: data.email };
      this.twofaEnabled = data.twofaType === 'totp';
    });
  },

  computed: {
    ...mapState(useMainStore, ['loading']),
  },
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
