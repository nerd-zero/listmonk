<template>
  <div>
    <div class="items mail-servers">
      <div class="block box" v-for="(item, n) in form.smtp" :key="n">
        <div class="grid">
          <div class="col-2">
            <div class="field">
              <div class="flex items-center gap-2">
                <PvToggleSwitch v-model="item.enabled" name="enabled" data-cy="btn-enable-smtp" />
                <span>{{ $t('globals.buttons.enabled') }}</span>
              </div>
            </div>
            <div class="field" v-if="form.smtp.length > 1">
              <a @click.prevent="$utils.confirm(null, () => removeSMTP(n))" href="#" data-cy="btn-delete-smtp">
                <i class="pi pi-trash" />
                {{ $t('globals.buttons.delete') }}
              </a>
            </div>
          </div><!-- first column -->

          <div class="col" :class="{ disabled: !item.enabled }">
            <div class="grid">
              <div class="col-9">
                <div class="field">
                  <label class="block mb-1 text-sm font-medium">{{ $t('settings.mailserver.host') }}</label>
                  <PvInputText v-model="item.host" name="host" placeholder="smtp.yourmailserver.net" :maxlength="200" />
                  <small class="block mt-1 text-color-secondary">{{ $t('settings.mailserver.hostHelp') }}</small>
                </div>
              </div>
              <div class="col">
                <div class="field">
                  <label class="block mb-1 text-sm font-medium">{{ $t('settings.mailserver.port') }}</label>
                  <PvInputNumber v-model="item.port" name="port" placeholder="25" :min="1" :max="65535" />
                  <small class="block mt-1 text-color-secondary">{{ $t('settings.mailserver.portHelp') }}</small>
                </div>
              </div>
            </div><!-- host -->

            <div class="grid">
              <div class="col-3">
                <div class="field">
                  <label class="block mb-1 text-sm font-medium">{{ $t('settings.mailserver.authProtocol') }}</label>
                  <PvSelect v-model="item.auth_protocol" name="auth_protocol"
                    :options="[{ label: 'LOGIN', value: 'login' }, { label: 'CRAM', value: 'cram' }, { label: 'PLAIN', value: 'plain' }, { label: 'None', value: 'none' }]"
                    option-label="label" option-value="value" />
                </div>
              </div>
              <div class="col">
                <div class="grid">
                  <div class="col">
                    <div class="field">
                      <label class="block mb-1 text-sm font-medium">{{ $t('settings.mailserver.username') }}</label>
                      <PvInputText v-model="item.username" :class="`smtp-username-${n}`"
                        :disabled="item.auth_protocol === 'none'" name="username" placeholder="mysmtp" :maxlength="200" />
                    </div>
                  </div>
                  <div class="col">
                    <div class="field">
                      <label class="block mb-1 text-sm font-medium">{{ $t('settings.mailserver.password') }}</label>
                      <PvPassword v-model="item.password" :disabled="item.auth_protocol === 'none'" name="password"
                        :input-class="`password-${n}`"
                        :placeholder="$t('settings.mailserver.passwordHelp')" :maxlength="200" :feedback="false" />
                      <small class="block mt-1 text-color-secondary">{{ $t('settings.mailserver.passwordHelp') }}</small>
                    </div>
                  </div>
                </div>
              </div>
            </div><!-- auth -->
            <div class="spaced-links is-size-7">
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
            <hr />

            <div class="grid">
              <div class="col-6">
                <div class="field">
                  <label class="block mb-1 text-sm font-medium">{{ $t('settings.smtp.heloHost') }}</label>
                  <PvInputText v-model="item.hello_hostname" name="hello_hostname" placeholder="" :maxlength="200" />
                  <small class="block mt-1 text-color-secondary">{{ $t('settings.smtp.heloHostHelp') }}</small>
                </div>
              </div>
              <div class="col">
                <div class="grid">
                  <div class="col">
                    <div class="field">
                      <label class="block mb-1 text-sm font-medium">{{ $t('settings.mailserver.tls') }}</label>
                      <PvSelect v-model="item.tls_type" name="items.tls_type"
                        :options="[{ label: $t('globals.states.off'), value: 'none' }, { label: 'STARTTLS', value: 'STARTTLS' }, { label: 'SSL/TLS', value: 'TLS' }]"
                        option-label="label" option-value="value" />
                      <small class="block mt-1 text-color-secondary">{{ $t('settings.mailserver.tlsHelp') }}</small>
                    </div>
                  </div>
                  <div class="col">
                    <div class="field">
                      <small class="block mt-1 text-color-secondary">{{ $t('settings.mailserver.skipTLSHelp') }}</small>
                      <div class="flex items-center gap-2">
                        <PvToggleSwitch v-model="item.tls_skip_verify" :disabled="item.tls_type === 'none'"
                          name="item.tls_skip_verify" />
                        <span>{{ $t('settings.mailserver.skipTLS') }}</span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div><!-- TLS -->
            <hr />

            <div class="grid">
              <div class="col-4">
                <div class="field">
                  <label class="block mb-1 text-sm font-medium">{{ $t('settings.mailserver.maxConns') }}</label>
                  <PvInputNumber v-model="item.max_conns" name="max_conns" placeholder="25" :min="1" :max="65535" />
                  <small class="block mt-1 text-color-secondary">{{ $t('settings.mailserver.maxConnsHelp') }}</small>
                </div>
              </div>
              <div class="col-4">
                <div class="field">
                  <label class="block mb-1 text-sm font-medium">{{ $t('settings.mailserver.idleTimeout') }}</label>
                  <PvInputText v-model="item.idle_timeout" name="idle_timeout" placeholder="15s" :pattern="regDuration"
                    :maxlength="10" />
                  <small class="block mt-1 text-color-secondary">{{ $t('settings.mailserver.idleTimeoutHelp') }}</small>
                </div>
              </div>
              <div class="col-4">
                <div class="field">
                  <label class="block mb-1 text-sm font-medium">{{ $t('settings.mailserver.waitTimeout') }}</label>
                  <PvInputText v-model="item.wait_timeout" name="wait_timeout" placeholder="5s" :pattern="regDuration"
                    :maxlength="10" />
                  <small class="block mt-1 text-color-secondary">{{ $t('settings.mailserver.waitTimeoutHelp') }}</small>
                </div>
              </div>
            </div>

            <div class="grid">
              <div class="col-4">
                <div class="field">
                  <label class="block mb-1 text-sm font-medium">{{ $t('settings.smtp.retries') }}</label>
                  <PvInputNumber v-model="item.max_msg_retries" name="max_msg_retries" placeholder="2" :min="1" :max="1000" />
                  <small class="block mt-1 text-color-secondary">{{ $t('settings.smtp.retriesHelp') }}</small>
                </div>
              </div>
              <div class="col-4">
                <div class="field">
                  <label class="block mb-1 text-sm font-medium">{{ $t('settings.smtp.retryDelay') }}</label>
                  <PvInputText v-model="item.msg_retry_delay" name="msg_retry_delay" placeholder="0s" :pattern="regDuration"
                    :maxlength="10" />
                  <small class="block mt-1 text-color-secondary">{{ $t('settings.smtp.retryDelayHelp') }}</small>
                </div>
              </div>
            </div>

            <hr />
            <div class="grid">
              <div class="col-6">
                <div class="field">
                  <label class="block mb-1 text-sm font-medium">{{ $t('globals.fields.name') }}</label>
                  <PvInputText v-model="item.name" name="name" placeholder="email-primary" :maxlength="100" />
                  <small class="block mt-1 text-color-secondary">{{ $t('settings.mailserver.nameHelp') }}</small>
                </div>
              </div>
              <div class="col-6">
                <div class="field">
                  <label class="block mb-1 text-sm font-medium">{{ $t('settings.smtp.fromAddresses') }}</label>
                  <PvAutoComplete v-model="item.from_addresses" name="from_addresses" multiple
                    :placeholder="'user@example.com, anothersite.com'" />
                  <small class="block mt-1 text-color-secondary">{{ $t('settings.smtp.fromAddressesHelp') }}</small>
                </div>
              </div>
            </div>

            <div class="grid">
              <div class="col">
                <p v-if="item.email_headers.length === 0 && !item.showHeaders">
                  <a href="#" @click.prevent="() => showSMTPHeaders(n)">
                    <i class="pi pi-plus" />{{ $t('settings.smtp.setCustomHeaders') }}</a>
                </p>
                <div class="field" v-if="item.email_headers.length > 0 || item.showHeaders">
                  <PvTextarea v-model="item.strEmailHeaders" name="email_headers"
                    placeholder="[{&quot;X-Custom&quot;: &quot;value&quot;}, {&quot;X-Custom2&quot;: &quot;value&quot;}]" />
                  <small class="block mt-1 text-color-secondary">{{ $t('settings.smtp.customHeadersHelp') }}</small>
                </div>
              </div>
            </div>
            <hr />

            <form @submit.prevent="() => doSMTPTest(item, n)">
              <div class="grid">
                <template v-if="smtpTestItem === n">
                  <div class="col-5">
                    <strong>{{ $t('settings.general.fromEmail') }}</strong>
                    <br />
                    {{ settings['app.from_email'] }}
                  </div>
                  <div class="col-4">
                    <div class="field">
                      <label class="block mb-1 text-sm font-medium">{{ $t('settings.smtp.toEmail') }}</label>
                      <PvInputText type="email" required v-model="testEmail" :ref="'testEmailTo'"
                        placeholder="email@site.com" :class="`test-email-${n}`" />
                    </div>
                  </div>
                </template>
                <div class="col has-text-right">
                  <PvButton v-if="smtpTestItem === n" severity="primary" @click.prevent="() => doSMTPTest(item, n)"
                    :label="$t('settings.smtp.sendTest')" />
                  <a href="#" v-else class="is-primary" @click.prevent="showTestForm(n)">
                    <i class="pi pi-send" /> {{ $t('settings.smtp.testConnection') }}
                  </a>
                </div>
                <div class="grid">
                  <div class="col" />
                </div>
              </div>
              <div v-if="errMsg && smtpTestItem === n">
                <div class="field mt-4">
                  <PvTextarea v-model="errMsg" class="has-text-danger is-size-6" readonly />
                </div>
              </div>
            </form><!-- smtp test -->
          </div>
        </div><!-- second container column -->
      </div><!-- block -->
    </div><!-- mail-servers -->

    <PvButton @click="addSMTP" icon="pi pi-plus" severity="primary" :label="$t('globals.buttons.addNew')" />
  </div>
</template>

<script>
import { mapState } from 'vuex';
import { regDuration } from '../../constants';

const smtpTemplates = {
  gmail: {
    host: 'smtp.gmail.com', port: 465, auth_protocol: 'login', tls_type: 'TLS',
  },
  ses: {
    host: 'email-smtp.YOUR-REGION.amazonaws.com', port: 465, auth_protocol: 'login', tls_type: 'TLS',
  },
  azure: {
    host: 'smtp.azurecomm.net', port: 587, auth_protocol: 'login', tls_type: 'STARTTLS',
  },
  mailjet: {
    host: 'in-v3.mailjet.com', port: 465, auth_protocol: 'cram', tls_type: 'TLS',
  },
  mailgun: {
    host: 'smtp.mailgun.org', port: 465, auth_protocol: 'login', tls_type: 'TLS',
  },
  sendgrid: {
    host: 'smtp.sendgrid.net', port: 465, auth_protocol: 'login', tls_type: 'TLS',
  },
  forwardemail: {
    host: 'smtp.forwardemail.net', port: 465, auth_protocol: 'login', tls_type: 'TLS',
  },
  postmark: {
    host: 'smtp.postmarkapp.com', port: 587, auth_protocol: 'cram', tls_type: 'STARTTLS',
  },
  lettermint: {
    host: 'smtp.lettermint.co', port: 465, auth_protocol: 'login', tls_type: 'TLS',
  },
};

export default {
  props: {
    form: {
      type: Object, default: () => { },
    },
  },

  data() {
    return {
      data: this.form,
      regDuration,
      // Index of the SMTP block item in the array to show the
      // test form in.
      smtpTestItem: null,
      testEmail: '',
      errMsg: '',
    };
  },

  methods: {
    addSMTP() {
      this.data.smtp.push({
        name: '',
        enabled: true,
        host: '',
        hello_hostname: '',
        port: 587,
        auth_protocol: 'none',
        username: '',
        password: '',
        email_headers: [],
        from_addresses: [],
        max_conns: 10,
        max_msg_retries: 2,
        msg_retry_delay: '0s',
        idle_timeout: '15s',
        wait_timeout: '5s',
        tls_type: 'STARTTLS',
        tls_skip_verify: false,
      });

      this.$nextTick(() => {
        const items = document.querySelectorAll('.mail-servers input[name="host"]');
        items[items.length - 1].focus();
      });
    },

    removeSMTP(i) {
      this.data.smtp.splice(i, 1);
    },

    showSMTPHeaders(i) {
      const s = this.data.smtp[i];
      s.showHeaders = true;
      this.data.smtp.splice(i, 1, s);
    },

    testConnection() {
      let em = this.settings['app.from_email'].replace('>', '').split('<');
      if (em.length > 1) {
        em = `<${em[em.length - 1]}>`;
      }
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
        if (err.response?.data?.message) {
          this.errMsg = err.response.data.message;
        }
      });
    },

    showTestForm(n) {
      this.smtpTestItem = n;
      this.testItem = this.form.smtp[n];
      this.errMsg = '';

      this.$nextTick(() => {
        document.querySelector(`.test-email-${n}`).focus();
      });
    },

    isTestEnabled(item) {
      if (!item.host || !item.port) {
        return false;
      }
      if (item.auth_protocol !== 'none' && item.password.includes('•')) {
        return false;
      }

      return true;
    },

    validateFromAddress(v) {
      // Accept an e-mail address (user@example.com) or a domain (example.com).
      return /^[^\s@]+(\.[^\s@]+)+$|^[^\s@]+@[^\s@]+(\.[^\s@]+)+$/.test(v);
    },

    fillSettings(n, key) {
      this.data.smtp.splice(n, 1, {
        ...this.data.smtp[n],
        ...smtpTemplates[key],
        username: '',
        password: '',
        hello_hostname: '',
        tls_skip_verify: false,
      });

      this.$nextTick(() => {
        document.querySelector(`.smtp-username-${n}`).focus();
      });
    },
  },

  computed: {
    ...mapState(['settings']),
  },
};
</script>
