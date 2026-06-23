<template>
  <section class="user-profile section-mini">
    <div v-if="loading.users" class="flex justify-center p-8">
      <PvProgressSpinner />
    </div>

    <h1 class="title">
      @{{ data.username }}
    </h1>
    <PvTag v-if="data.userRole">{{ data.userRole.name }}</PvTag>

    <br /><br /><br />
    <form @submit.prevent="onSubmit">
      <div v-if="data.type !== 'api'" class="field">
        <label class="block mb-1 text-sm font-medium">{{ $t('subscribers.email') }}</label>
        <PvInputText :maxlength="200" v-model="form.email" name="email" :placeholder="$t('subscribers.email')"
          :disabled="!data.passwordLogin" required autofocus />
      </div>

      <div class="field">
        <label class="block mb-1 text-sm font-medium">{{ $t('globals.fields.name') }}</label>
        <PvInputText :maxlength="200" v-model="form.name" name="name" :placeholder="$t('globals.fields.name')" />
      </div>

      <div v-if="data.passwordLogin" class="grid">
        <div class="col-6">
          <div class="field">
            <label class="block mb-1 text-sm font-medium">{{ $t('users.password') }}</label>
            <PvPassword minlength="8" :maxlength="200" v-model="form.password" name="password"
              :placeholder="$t('users.password')" :feedback="false" />
          </div>
        </div>
        <div class="col-6">
          <div class="field">
            <label class="block mb-1 text-sm font-medium">{{ $t('users.passwordRepeat') }}</label>
            <PvPassword minlength="8" :maxlength="200" v-model="form.password2" name="password2" :feedback="false" />
          </div>
        </div>
      </div>

      <div class="field">
        <PvButton severity="primary" icon="pi pi-save" type="submit" data-cy="btn-save"
          :label="$t('globals.buttons.save')" />
      </div>
    </form>

    <br /><br />

    <!-- 2FA -->
    <section v-if="data.passwordLogin" class="twofa-section">
      <!-- TOTP disabled -->
      <div v-if="data.twofaType === 'none'" class="box">
        <div class="grid align-items-center mb-4">
          <div class="col">
            <h3 class="title is-size-5 mb-0">{{ $t('users.twoFA') }}</h3>
          </div>
          <div class="col-auto">
            <div v-if="!isTotpVisible" class="flex items-center gap-2">
              <PvToggleSwitch v-model="twofaEnabled" @change="onToggleEnableTotp" />
            </div>
          </div>
        </div>

        <p>{{ $t('users.twoFANotEnabled') }}</p>
        <br />

        <!-- TOTP setup -->
        <div v-if="isTotpVisible" class="totp-setup">
          <div v-if="totpQR" class="qr-section">
            <p class="has-text-grey">{{ $t('users.totpScanQR') }}</p><br />

            <img :src="'data:image/png;base64,' + totpQR" alt="QR Code" />

            <br /><br />
            <p>
              <strong>{{ $t('users.totpSecret') }}</strong><br />
              <code><copy-text :text="`${totpSecret}`" /></code>
            </p>

            <br /><br />
            <form @submit.prevent="confirmTOTP">
              <div class="field">
                <label class="block mb-1 text-sm font-medium">{{ $t('users.totpCode') }}</label>
                <PvInputText ref="totpCodeInput" v-model="totpCode" maxlength="6" pattern="[0-9]{6}"
                  placeholder="000000" required />
              </div>
              <div class="buttons">
                <PvButton severity="primary" type="submit" :label="$t('globals.buttons.enable')" />
                <PvButton type="button" @click="onCancelTOTPSetup" :label="$t('globals.buttons.cancel')" />
              </div>
            </form>
          </div>
        </div>
      </div>

      <!-- TOTP Enabled -->
      <div v-if="data.twofaType === 'totp'" class="box">
        <div class="grid align-items-center">
          <div class="col">
            <h3 class="title is-size-5">
              <i class="pi pi-check-circle text-green-500" /> {{ $t('users.twoFAEnabled') }}
            </h3>
          </div>
          <div class="col-auto">
            <div v-if="!showDisableTOTP" class="flex items-center gap-2">
              <PvToggleSwitch v-model="twofaEnabled" @change="toggleDisableTOTP" />
            </div>
          </div>
        </div>

        <p>{{ $t('users.twoFAEnabledDesc', { type: data.twofaType.toUpperCase() }) }}</p>

        <!-- Disable TOTP Flow -->
        <form v-if="showDisableTOTP" class="disable-totp mt-5" @submit.prevent="confirmDisableTOTP">
          <div class="field">
            <label class="block mb-1 text-sm font-medium">{{ $t('users.password') }}</label>
            <PvPassword ref="disablePasswordInput" v-model="disableTOTPPassword" minlength="8" required
              :feedback="false" />
          </div>
          <div class="buttons">
            <PvButton severity="danger" type="submit" :label="$t('globals.buttons.disable')" />
            <PvButton type="button" @click="onCancelTOTPSetup" :label="$t('globals.buttons.cancel')" />
          </div>
        </form>
      </div>
    </section>
  </section>
</template>

<script>
import { mapState } from 'vuex';
import CopyText from '../components/CopyText.vue';

export default {
  name: 'UserProfile',

  components: {
    CopyText,
  },

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

        this.$nextTick(() => {
          if (this.$refs.totpCodeInput) {
            this.$refs.totpCodeInput.focus();
          }
        });
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

        // Reload user profile
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

      this.$nextTick(() => {
        if (this.$refs.disablePasswordInput) {
          this.$refs.disablePasswordInput.focus();
        }
      });
    },

    cancelDisableTOTP() {
      this.showDisableTOTP = false;
      this.disableTOTPPassword = '';
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
        // Reload user profile
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
    ...mapState(['loading']),
  },
};
</script>
