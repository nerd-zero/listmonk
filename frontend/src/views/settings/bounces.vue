<template>
  <div>
    <div class="columns mb-6">
      <div class="column is-4">
        <div class="field" data-cy="btn-enable-bounce">
          <div class="flex items-center gap-2">
            <PvToggleSwitch v-model="data['bounce.enabled']" name="bounce.enabled" />
            <span>{{ $t('settings.bounces.enable') }}</span>
          </div>
        </div>
      </div>
      <div class="column">
        <div v-for="typ in bounceTypes" :key="typ" class="columns">
          <div class="column is-2" :class="{ disabled: !data['bounce.enabled'] }" :label="$t('settings.bounces.count')"
            label-position="on-border">
            {{ $t(`bounces.${typ}`) }}
          </div>
          <div class="column is-4" :class="{ disabled: !data['bounce.enabled'] }">
            <div class="field" data-cy="btn-bounce-count">
              <label class="block mb-1 text-sm font-medium">{{ $t('settings.bounces.count') }}</label>
              <PvInputNumber v-model="data['bounce.actions'][typ]['count']" name="bounce.count"
                placeholder="3" :min="1" :max="1000" />
              <small class="block mt-1 text-color-secondary">{{ $t('settings.bounces.countHelp') }}</small>
            </div>
          </div>
          <div class="column is-4" :class="{ disabled: !data['bounce.enabled'] }">
            <div class="field">
              <label class="block mb-1 text-sm font-medium">{{ $t('settings.bounces.action') }}</label>
              <PvSelect v-model="data['bounce.actions'][typ]['action']" name="bounce.action"
                :options="bounceActionOptions" option-label="label" option-value="value" />
            </div>
          </div>
        </div>
      </div>
    </div><!-- columns -->

    <div class="mb-6">
      <div class="field" data-cy="btn-enable-bounce-webhook">
        <div class="flex items-center gap-2">
          <PvToggleSwitch v-model="data['bounce.webhooks_enabled']" :disabled="!data['bounce.enabled']"
            name="webhooks_enabled" data-cy="btn-enable-bounce-webhook" />
          <span>{{ $t('settings.bounces.enableWebhooks') }}</span>
        </div>
        <p class="has-text-grey">
          <a href="https://listmonk.app/docs/bounces" target="_blank" rel="noopener noreferer">{{
            $t('globals.buttons.learnMore') }} &rarr;</a>
        </p>
      </div>
      <div class="box" v-if="data['bounce.webhooks_enabled']">
        <div class="columns">
          <div class="column">
            <div class="field">
              <div class="flex items-center gap-2">
                <PvToggleSwitch v-model="data['bounce.ses_enabled']" name="ses_enabled"
                  data-cy="btn-enable-bounce-ses" />
                <span>{{ $t('settings.bounces.enableSES') }}</span>
              </div>
            </div>
          </div>
        </div>
        <div class="columns">
          <div class="column is-3">
            <div class="field">
              <div class="flex items-center gap-2">
                <PvToggleSwitch v-model="data['bounce.azure'].enabled" name="azure_enabled"
                  data-cy="btn-enable-bounce-azure" />
                <span>{{ $t('settings.bounces.enableAzure') }}</span>
              </div>
            </div>
          </div>
          <div class="column">
            <div class="field">
              <label class="block mb-1 text-sm font-medium">{{ $t('settings.bounces.azureSharedSecret') }}</label>
              <PvPassword v-model="data['bounce.azure'].shared_secret" :feedback="false"
                :disabled="!data['bounce.azure'].enabled" name="azure_shared_secret"
                data-cy="bounce-azure-shared-secret" />
              <small class="block mt-1 text-color-secondary">{{ $t('settings.bounces.azureSharedSecretHelp') }}</small>
            </div>
          </div>
          <div class="column">
            <div class="field">
              <label class="block mb-1 text-sm font-medium">{{ $t('settings.bounces.azureSharedSecretHeader') }}</label>
              <PvInputText v-model="data['bounce.azure'].shared_secret_header" type="text"
                :disabled="!data['bounce.azure'].enabled" name="azure_shared_secret_header"
                data-cy="bounce-azure-shared-secret-header" />
              <small class="block mt-1 text-color-secondary">{{ $t('settings.bounces.azureSharedSecretHeaderHelp') }}</small>
            </div>
          </div>
        </div>
        <div class="columns">
          <div class="column is-3">
            <div class="field">
              <div class="flex items-center gap-2">
                <PvToggleSwitch v-model="data['bounce.sendgrid_enabled']" name="sendgrid_enabled"
                  data-cy="btn-enable-bounce-sendgrid" />
                <span>{{ $t('settings.bounces.enableSendgrid') }}</span>
              </div>
            </div>
          </div>
          <div class="column">
            <div class="field">
              <label class="block mb-1 text-sm font-medium">{{ $t('settings.bounces.sendgridKey') }}</label>
              <PvPassword v-model="data['bounce.sendgrid_key']" :feedback="false"
                :disabled="!data['bounce.sendgrid_enabled']" name="sendgrid_enabled"
                data-cy="btn-enable-bounce-sendgrid" />
              <small class="block mt-1 text-color-secondary">{{ $t('globals.messages.passwordChange') }}</small>
            </div>
          </div>
        </div>
        <div class="columns">
          <div class="column is-3">
            <div class="field">
              <div class="flex items-center gap-2">
                <PvToggleSwitch v-model="data['bounce.postmark'].enabled" name="postmark_enabled"
                  data-cy="btn-enable-bounce-postmark" />
                <span>{{ $t('settings.bounces.enablePostmark') }}</span>
              </div>
            </div>
          </div>
          <div class="column">
            <div class="field">
              <label class="block mb-1 text-sm font-medium">{{ $t('settings.bounces.postmarkUsername') }}</label>
              <PvInputText v-model="data['bounce.postmark'].username" type="text"
                :disabled="!data['bounce.postmark'].enabled" name="postmark_username"
                data-cy="btn-enable-bounce-postmark" />
              <small class="block mt-1 text-color-secondary">{{ $t('settings.bounces.postmarkUsernameHelp') }}</small>
            </div>
          </div>
          <div class="column">
            <div class="field">
              <label class="block mb-1 text-sm font-medium">{{ $t('settings.bounces.postmarkPassword') }}</label>
              <PvPassword v-model="data['bounce.postmark'].password" :feedback="false"
                :disabled="!data['bounce.postmark'].enabled" name="postmark_password"
                data-cy="btn-enable-bounce-postmark" />
              <small class="block mt-1 text-color-secondary">{{ $t('globals.messages.passwordChange') }}</small>
            </div>
          </div>
        </div>
        <div class="columns">
          <div class="column is-3">
            <div class="field">
              <div class="flex items-center gap-2">
                <PvToggleSwitch v-model="data['bounce.forwardemail'].enabled" name="forwardemail_enabled"
                  data-cy="btn-enable-bounce-forwardemail" />
                <span>{{ $t('settings.bounces.enableForwardemail') }}</span>
              </div>
            </div>
          </div>
          <div class="column">
            <div class="field">
              <label class="block mb-1 text-sm font-medium">{{ $t('settings.bounces.forwardemailKey') }}</label>
              <PvPassword v-model="data['bounce.forwardemail'].key" :feedback="false"
                :disabled="!data['bounce.forwardemail'].enabled" name="forwardemail_enabled"
                data-cy="btn-enable-bounce-forwardemail" />
              <small class="block mt-1 text-color-secondary">{{ $t('globals.messages.passwordChange') }}</small>
            </div>
          </div>
        </div>
        <div class="columns">
          <div class="column is-3">
            <div class="field">
              <div class="flex items-center gap-2">
                <PvToggleSwitch v-model="data['bounce.lettermint'].enabled" name="lettermint_enabled"
                  data-cy="btn-enable-bounce-lettermint" />
                <span>{{ $t('settings.bounces.enableLettermint') }}</span>
              </div>
            </div>
          </div>
          <div class="column">
            <div class="field">
              <label class="block mb-1 text-sm font-medium">{{ $t('settings.bounces.lettermintKey') }}</label>
              <PvPassword v-model="data['bounce.lettermint'].key" :feedback="false"
                :disabled="!data['bounce.lettermint'].enabled" name="lettermint_key"
                data-cy="bounce-lettermint-key" />
              <small class="block mt-1 text-color-secondary">{{ $t('globals.messages.passwordChange') }}</small>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- bounce mailbox -->
    <div class="field">
      <div class="flex items-center gap-2">
        <PvToggleSwitch v-if="data['bounce.mailboxes']" v-model="data['bounce.mailboxes'][0].enabled"
          :disabled="!data['bounce.enabled']" name="enabled" data-cy="btn-enable-bounce-mailbox" />
        <span v-if="data['bounce.mailboxes']">{{ $t('settings.bounces.enableMailbox') }}</span>
      </div>
    </div>

    <template v-if="data['bounce.enabled'] && data['bounce.mailboxes'][0].enabled">
      <div class="block box" v-for="(item, n) in data['bounce.mailboxes']" :key="n">
        <div class="columns">
          <div class="column" :class="{ disabled: !item.enabled }">
            <div class="columns">
              <div class="column is-3">
                <div class="field">
                  <label class="block mb-1 text-sm font-medium">{{ $t('settings.bounces.type') }}</label>
                  <PvSelect v-model="item.type" name="type"
                    :options="[{ label: 'POP', value: 'pop' }]" option-label="label" option-value="value" />
                </div>
              </div>
              <div class="column is-6">
                <div class="field">
                  <label class="block mb-1 text-sm font-medium">{{ $t('settings.mailserver.host') }}</label>
                  <PvInputText v-model="item.host" name="host" placeholder="bounce.yourmailserver.net" :maxlength="200" />
                  <small class="block mt-1 text-color-secondary">{{ $t('settings.mailserver.hostHelp') }}</small>
                </div>
              </div>
              <div class="column is-3">
                <div class="field">
                  <label class="block mb-1 text-sm font-medium">{{ $t('settings.mailserver.port') }}</label>
                  <PvInputNumber v-model="item.port" name="port" placeholder="25" :min="1" :max="65535" />
                  <small class="block mt-1 text-color-secondary">{{ $t('settings.mailserver.portHelp') }}</small>
                </div>
              </div>
            </div><!-- host -->

            <div class="columns">
              <div class="column is-3">
                <div class="field">
                  <label class="block mb-1 text-sm font-medium">{{ $t('settings.mailserver.authProtocol') }}</label>
                  <PvSelect v-model="item.auth_protocol" name="auth_protocol"
                    :options="getAuthProtocolOptions(item.type)" option-label="label" option-value="value" />
                </div>
              </div>
              <div class="column">
                <div class="columns">
                  <div class="column">
                    <div class="field">
                      <label class="block mb-1 text-sm font-medium">{{ $t('settings.mailserver.username') }}</label>
                      <PvInputText v-model="item.username" :disabled="item.auth_protocol === 'none'" name="username"
                        placeholder="mysmtp" :maxlength="200" />
                    </div>
                  </div>
                  <div class="column">
                    <div class="field">
                      <label class="block mb-1 text-sm font-medium">{{ $t('settings.mailserver.password') }}</label>
                      <PvPassword v-model="item.password" :disabled="item.auth_protocol === 'none'" name="password"
                        :feedback="false" :placeholder="$t('settings.mailserver.passwordHelp')" :maxlength="200" />
                      <small class="block mt-1 text-color-secondary">{{ $t('settings.mailserver.passwordHelp') }}</small>
                    </div>
                  </div>
                </div>
              </div>
            </div><!-- auth -->

            <div class="columns">
              <div class="column is-6">
                <div class="columns">
                  <div class="column">
                    <div class="field">
                      <div class="flex items-center gap-2">
                        <PvToggleSwitch v-model="item.tls_enabled" name="item.tls_enabled" />
                        <span>{{ $t('settings.mailserver.tls') }}</span>
                      </div>
                      <small class="block mt-1 text-color-secondary">{{ $t('settings.mailserver.tlsHelp') }}</small>
                    </div>
                  </div>
                  <div class="column">
                    <div class="field">
                      <div class="flex items-center gap-2">
                        <PvToggleSwitch v-model="item.tls_skip_verify" :disabled="!item.tls_enabled"
                          name="item.tls_skip_verify" />
                        <span>{{ $t('settings.mailserver.skipTLS') }}</span>
                      </div>
                      <small class="block mt-1 text-color-secondary">{{ $t('settings.mailserver.skipTLSHelp') }}</small>
                    </div>
                  </div>
                </div>
              </div>
              <div class="column" />
              <div class="column is-4">
                <div class="field">
                  <label class="block mb-1 text-sm font-medium">{{ $t('settings.bounces.scanInterval') }}</label>
                  <PvInputText v-model="item.scan_interval" name="scan_interval" placeholder="15m"
                    :pattern="regDuration" :maxlength="10" />
                  <small class="block mt-1 text-color-secondary">{{ $t('settings.bounces.scanIntervalHelp') }}</small>
                </div>
              </div>
            </div><!-- TLS -->
          </div>
        </div><!-- second container column -->
      </div><!-- block -->
    </template>
  </div>
</template>

<script>
import { regDuration } from '../../constants';

export default {
  props: {
    form: {
      type: Object, default: () => { },
    },
  },

  data() {
    return {
      bounceTypes: ['soft', 'hard', 'complaint'],
      data: this.form,
      regDuration,
    };
  },

  computed: {
    bounceActionOptions() {
      return [
        { label: this.$t('globals.terms.none'), value: 'none' },
        { label: this.$t('email.unsub'), value: 'unsubscribe' },
        { label: this.$t('settings.bounces.blocklist'), value: 'blocklist' },
        { label: this.$t('globals.buttons.delete'), value: 'delete' },
      ];
    },
  },

  methods: {
    removeBounceBox(i) {
      this.data['bounce.mailboxes'].splice(i, 1);
    },

    getAuthProtocolOptions(type) {
      const opts = [{ label: 'none', value: 'none' }];
      if (type === 'pop') {
        opts.push({ label: 'userpass', value: 'userpass' });
      } else {
        opts.push({ label: 'cram', value: 'cram' });
        opts.push({ label: 'plain', value: 'plain' });
        opts.push({ label: 'login', value: 'login' });
      }
      return opts;
    },
  },
};
</script>
