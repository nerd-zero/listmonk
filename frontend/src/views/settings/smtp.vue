<template>
  <div>
    <div class="smtp-list">
      <div class="smtp-card" v-for="(item, n) in form.smtp" :key="n">
        <!-- Card header -->
        <div class="smtp-card-header">
          <div class="flex items-center gap-2">
            <PvToggleSwitch v-model="item.enabled" name="enabled" data-cy="btn-enable-smtp" />
            <span class="smtp-card-title">{{ item.name || `SMTP #${n + 1}` }}</span>
          </div>
          <a v-if="form.smtp.length > 1" href="#" class="delete-link"
            @click.prevent="$utils.confirm(null, () => removeSMTP(n))" data-cy="btn-delete-smtp">
            <i class="pi pi-trash" /> {{ $t('globals.buttons.delete') }}
          </a>
        </div>

        <div :class="{ disabled: !item.enabled }">
          <!-- Host / Port -->
          <div class="smtp-section">
            <div class="grid">
              <div class="col-9">
                <div class="field">
                  <label class="block mb-1 text-sm font-medium">{{ $t('settings.mailserver.host') }}</label>
                  <PvInputText v-model="item.host" name="host" placeholder="smtp.yourmailserver.net"
                    :maxlength="200" class="w-full" />
                  <small class="block mt-1 text-color-secondary">{{ $t('settings.mailserver.hostHelp') }}</small>
                </div>
              </div>
              <div class="col-3">
                <div class="field">
                  <label class="block mb-1 text-sm font-medium">{{ $t('settings.mailserver.port') }}</label>
                  <PvInputNumber v-model="item.port" name="port" placeholder="25" :min="1" :max="65535" class="w-full" />
                </div>
              </div>
            </div>

            <div class="quick-links">
              <a href="#" @click.prevent="() => fillSettings(n, 'gmail')">Gmail</a>
              <a href="#" @click.prevent="() => fillSettings(n, 'ses')">Amazon SES</a>
              <a href="#" @click.prevent="() => fillSettings(n, 'azure')">Azure ACS</a>
              <a href="#" @click.prevent="() => fillSettings(n, 'mailgun')">Mailgun</a>
              <a href="#" @click.prevent="() => fillSettings(n, 'mailjet')">Mailjet</a>
              <a href="#" @click.prevent="() => fillSettings(n, 'sendgrid')">Sendgrid</a>
              <a href="#" @click.prevent="() => fillSettings(n, 'postmark')">Postmark</a>
              <a href="#" @click.prevent="() => fillSettings(n, 'forwardemail')">Forward Email</a>
              <a href="#" @click.prevent="() => fillSettings(n, 'lettermint')">Lettermint</a>
            </div>
          </div>

          <!-- Authentication -->
          <div class="smtp-section">
            <p class="smtp-section-label">{{ $t('settings.mailserver.authProtocol') }}</p>
            <div class="grid">
              <div class="col-3">
                <div class="field">
                  <PvSelect v-model="item.auth_protocol" name="auth_protocol"
                    :options="[{ label: 'LOGIN', value: 'login' }, { label: 'CRAM', value: 'cram' }, { label: 'PLAIN', value: 'plain' }, { label: 'None', value: 'none' }]"
                    option-label="label" option-value="value" class="w-full" />
                </div>
              </div>
              <div class="col-4">
                <div class="field">
                  <PvInputText v-model="item.username" :class="`smtp-username-${n}`"
                    :disabled="item.auth_protocol === 'none'" name="username"
                    :placeholder="$t('settings.mailserver.username')" :maxlength="200" class="w-full" />
                </div>
              </div>
              <div class="col-5">
                <div class="field">
                  <PvPassword v-model="item.password" :disabled="item.auth_protocol === 'none'" name="password"
                    :input-class="`password-${n}`"
                    :placeholder="$t('settings.mailserver.passwordHelp')" :maxlength="200" :feedback="false"
                    class="w-full" />
                  <small class="block mt-1 text-color-secondary">{{ $t('settings.mailserver.passwordHelp') }}</small>
                </div>
              </div>
            </div>
          </div>

          <!-- TLS -->
          <div class="smtp-section">
            <p class="smtp-section-label">TLS</p>
            <div class="grid">
              <div class="col-4">
                <div class="field">
                  <label class="block mb-1 text-sm font-medium">{{ $t('settings.smtp.heloHost') }}</label>
                  <PvInputText v-model="item.hello_hostname" name="hello_hostname" placeholder="" :maxlength="200"
                    class="w-full" />
                  <small class="block mt-1 text-color-secondary">{{ $t('settings.smtp.heloHostHelp') }}</small>
                </div>
              </div>
              <div class="col-4">
                <div class="field">
                  <label class="block mb-1 text-sm font-medium">{{ $t('settings.mailserver.tls') }}</label>
                  <PvSelect v-model="item.tls_type" name="items.tls_type"
                    :options="[{ label: $t('globals.states.off'), value: 'none' }, { label: 'STARTTLS', value: 'STARTTLS' }, { label: 'SSL/TLS', value: 'TLS' }]"
                    option-label="label" option-value="value" class="w-full" />
                </div>
              </div>
              <div class="col-4">
                <div class="field">
                  <label class="block mb-1 text-sm font-medium invisible">-</label>
                  <div class="flex items-center gap-2 mt-1">
                    <PvToggleSwitch v-model="item.tls_skip_verify" :disabled="item.tls_type === 'none'"
                      name="item.tls_skip_verify" />
                    <span class="text-sm">{{ $t('settings.mailserver.skipTLS') }}</span>
                  </div>
                  <small class="block mt-1 text-color-secondary">{{ $t('settings.mailserver.skipTLSHelp') }}</small>
                </div>
              </div>
            </div>
          </div>

          <!-- Limits -->
          <div class="smtp-section">
            <p class="smtp-section-label">{{ $t('settings.mailserver.maxConns') }} &amp; Timeouts</p>
            <div class="grid">
              <div class="col-3">
                <div class="field">
                  <label class="block mb-1 text-sm font-medium">{{ $t('settings.mailserver.maxConns') }}</label>
                  <PvInputNumber v-model="item.max_conns" name="max_conns" placeholder="25" :min="1" :max="65535"
                    class="w-full" />
                  <small class="block mt-1 text-color-secondary">{{ $t('settings.mailserver.maxConnsHelp') }}</small>
                </div>
              </div>
              <div class="col-3">
                <div class="field">
                  <label class="block mb-1 text-sm font-medium">{{ $t('settings.smtp.retries') }}</label>
                  <PvInputNumber v-model="item.max_msg_retries" name="max_msg_retries" placeholder="2" :min="1"
                    :max="1000" class="w-full" />
                </div>
              </div>
              <div class="col-2">
                <div class="field">
                  <label class="block mb-1 text-sm font-medium">{{ $t('settings.smtp.retryDelay') }}</label>
                  <PvInputText v-model="item.msg_retry_delay" name="msg_retry_delay" placeholder="0s"
                    :pattern="regDuration" :maxlength="10" class="w-full" />
                </div>
              </div>
              <div class="col-2">
                <div class="field">
                  <label class="block mb-1 text-sm font-medium">{{ $t('settings.mailserver.idleTimeout') }}</label>
                  <PvInputText v-model="item.idle_timeout" name="idle_timeout" placeholder="15s"
                    :pattern="regDuration" :maxlength="10" class="w-full" />
                </div>
              </div>
              <div class="col-2">
                <div class="field">
                  <label class="block mb-1 text-sm font-medium">{{ $t('settings.mailserver.waitTimeout') }}</label>
                  <PvInputText v-model="item.wait_timeout" name="wait_timeout" placeholder="5s"
                    :pattern="regDuration" :maxlength="10" class="w-full" />
                </div>
              </div>
            </div>
          </div>

          <!-- Identity -->
          <div class="smtp-section">
            <p class="smtp-section-label">{{ $t('globals.fields.name') }} &amp; Addresses</p>
            <div class="grid">
              <div class="col-5">
                <div class="field">
                  <label class="block mb-1 text-sm font-medium">{{ $t('globals.fields.name') }}</label>
                  <PvInputText v-model="item.name" name="name" placeholder="email-primary" :maxlength="100"
                    class="w-full" />
                  <small class="block mt-1 text-color-secondary">{{ $t('settings.mailserver.nameHelp') }}</small>
                </div>
              </div>
              <div class="col-7">
                <div class="field">
                  <label class="block mb-1 text-sm font-medium">{{ $t('settings.smtp.fromAddresses') }}</label>
                  <PvAutoComplete v-model="item.from_addresses" name="from_addresses" multiple
                    :placeholder="'user@example.com, anothersite.com'" class="w-full" />
                  <small class="block mt-1 text-color-secondary">{{ $t('settings.smtp.fromAddressesHelp') }}</small>
                </div>
              </div>
            </div>

            <div class="field">
              <p v-if="item.email_headers.length === 0 && !item.showHeaders">
                <a href="#" class="settings-link" @click.prevent="() => showSMTPHeaders(n)">
                  <i class="pi pi-plus" /> {{ $t('settings.smtp.setCustomHeaders') }}
                </a>
              </p>
              <div v-if="item.email_headers.length > 0 || item.showHeaders">
                <PvTextarea v-model="item.strEmailHeaders" name="email_headers" class="w-full"
                  placeholder="[{&quot;X-Custom&quot;: &quot;value&quot;}, {&quot;X-Custom2&quot;: &quot;value&quot;}]" />
                <small class="block mt-1 text-color-secondary">{{ $t('settings.smtp.customHeadersHelp') }}</small>
              </div>
            </div>
          </div>

          <!-- Test -->
          <div class="smtp-section smtp-section--test">
            <form @submit.prevent="() => doSMTPTest(item, n)">
              <div class="smtp-test-row">
                <template v-if="smtpTestItem === n">
                  <div class="smtp-test-from">
                    <span class="text-sm font-medium">{{ $t('settings.general.fromEmail') }}</span>
                    <span class="text-sm text-color-secondary">{{ settings['app.from_email'] }}</span>
                  </div>
                  <div class="field" style="flex:1">
                    <label class="block mb-1 text-sm font-medium">{{ $t('settings.smtp.toEmail') }}</label>
                    <PvInputText type="email" required v-model="testEmail" :ref="'testEmailTo'"
                      placeholder="email@site.com" :class="`test-email-${n}`" class="w-full" />
                  </div>
                  <PvButton severity="primary" type="submit" :label="$t('settings.smtp.sendTest')" />
                </template>
                <a href="#" v-else class="settings-link" @click.prevent="showTestForm(n)">
                  <i class="pi pi-send" /> {{ $t('settings.smtp.testConnection') }}
                </a>
              </div>
              <div v-if="errMsg && smtpTestItem === n" class="field mt-3">
                <PvTextarea v-model="errMsg" class="w-full text-red-500 text-sm" readonly />
              </div>
            </form>
          </div>
        </div>
      </div>
    </div>

    <PvButton @click="addSMTP" icon="pi pi-plus" severity="primary" :label="$t('globals.buttons.addNew')" />
  </div>
</template>

<script>
import { mapState } from 'pinia';
import { useMainStore } from '../../store';
import { regDuration } from '../../constants';

const smtpTemplates = {
  gmail: { host: 'smtp.gmail.com', port: 465, auth_protocol: 'login', tls_type: 'TLS' },
  ses: { host: 'email-smtp.YOUR-REGION.amazonaws.com', port: 465, auth_protocol: 'login', tls_type: 'TLS' },
  azure: { host: 'smtp.azurecomm.net', port: 587, auth_protocol: 'login', tls_type: 'STARTTLS' },
  mailjet: { host: 'in-v3.mailjet.com', port: 465, auth_protocol: 'cram', tls_type: 'TLS' },
  mailgun: { host: 'smtp.mailgun.org', port: 465, auth_protocol: 'login', tls_type: 'TLS' },
  sendgrid: { host: 'smtp.sendgrid.net', port: 465, auth_protocol: 'login', tls_type: 'TLS' },
  forwardemail: { host: 'smtp.forwardemail.net', port: 465, auth_protocol: 'login', tls_type: 'TLS' },
  postmark: { host: 'smtp.postmarkapp.com', port: 587, auth_protocol: 'cram', tls_type: 'STARTTLS' },
  lettermint: { host: 'smtp.lettermint.co', port: 465, auth_protocol: 'login', tls_type: 'TLS' },
};

export default {
  props: {
    form: { type: Object, default: () => {} },
  },

  data() {
    return { data: this.form, regDuration, smtpTestItem: null, testEmail: '', errMsg: '' };
  },

  methods: {
    addSMTP() {
      this.data.smtp.push({
        name: '', enabled: true, host: '', hello_hostname: '', port: 587,
        auth_protocol: 'none', username: '', password: '', email_headers: [],
        from_addresses: [], max_conns: 10, max_msg_retries: 2, msg_retry_delay: '0s',
        idle_timeout: '15s', wait_timeout: '5s', tls_type: 'STARTTLS', tls_skip_verify: false,
      });
      this.$nextTick(() => {
        const items = document.querySelectorAll('.smtp-list input[name="host"]');
        items[items.length - 1].focus();
      });
    },

    removeSMTP(i) { this.data.smtp.splice(i, 1); },

    showSMTPHeaders(i) {
      const s = this.data.smtp[i];
      s.showHeaders = true;
      this.data.smtp.splice(i, 1, s);
    },

    doSMTPTest(item, n) {
      if (!this.isTestEnabled(item)) {
        this.$utils.toast(this.$t('settings.smtp.testEnterEmail'), 'is-danger');
        this.$nextTick(() => {
          const i = document.querySelector(`.password-${n}`);
          this.data.smtp[n].password = '';
          i.focus();
          i.select();
        });
        return;
      }
      this.errMsg = '';
      this.$api.testSMTP({ ...item, email: this.testEmail }).then(() => {
        this.$utils.toast(this.$t('campaigns.testSent'));
      }).catch((err) => {
        if (err.response?.data?.message) { this.errMsg = err.response.data.message; }
      });
    },

    showTestForm(n) {
      this.smtpTestItem = n;
      this.testItem = this.form.smtp[n];
      this.errMsg = '';
      this.$nextTick(() => { document.querySelector(`.test-email-${n}`).focus(); });
    },

    isTestEnabled(item) {
      if (!item.host || !item.port) return false;
      if (item.auth_protocol !== 'none' && item.password.includes('•')) return false;
      return true;
    },

    fillSettings(n, key) {
      this.data.smtp.splice(n, 1, {
        ...this.data.smtp[n], ...smtpTemplates[key],
        username: '', password: '', hello_hostname: '', tls_skip_verify: false,
      });
      this.$nextTick(() => { document.querySelector(`.smtp-username-${n}`).focus(); });
    },
  },

  computed: {
    ...mapState(useMainStore, ['settings']),
  },
};
</script>

<style scoped lang="scss">
.smtp-list { display: flex; flex-direction: column; gap: 1.25rem; margin-bottom: 1.25rem; }

.smtp-card {
  background: var(--lm-surface);
  border: 1px solid var(--lm-border);
  border-radius: 10px;
  overflow: hidden;
}

.smtp-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.85rem 1.25rem;
  background: var(--lm-bg-subtle);
  border-bottom: 1px solid var(--lm-border);
}

.smtp-card-title { font-weight: 600; font-size: 0.95rem; color: var(--lm-text); }

.smtp-section {
  padding: 1rem 1.25rem;
  & + & { border-top: 1px solid var(--lm-border); }
  &--test { background: var(--lm-bg-subtle); }
}

.smtp-section-label {
  font-size: 0.7rem;
  font-weight: 700;
  color: var(--lm-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.06em;
  margin-bottom: 0.75rem;
}

.delete-link {
  font-size: 0.8rem;
  color: var(--p-red-500);
  text-decoration: none;
  &:hover { text-decoration: underline; }
}

.smtp-test-row {
  display: flex;
  align-items: flex-end;
  gap: 1rem;
  flex-wrap: wrap;
}

.smtp-test-from {
  display: flex;
  flex-direction: column;
  gap: 0.2rem;
}

:deep(.p-password) { width: 100%; }
:deep(.p-password-input) { width: 100%; }
</style>
