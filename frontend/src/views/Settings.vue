<template>
  <form @submit.prevent="onSubmit">
    <section class="settings">
      <div v-if="loading.settings || isLoading" class="flex justify-center p-8">
        <PvProgressSpinner />
      </div>
      <div class="page-header" style="margin-bottom:1.5rem">
        <h1 class="page-title">
          {{ $t('settings.title') }}
          <span style="font-size:0.85rem;font-weight:400;color:#94a3b8">({{ serverConfig.version }})</span>
        </h1>
        <PvButton v-if="$can('settings:manage')" :disabled="!hasFormChanged" severity="primary" icon="pi pi-save"
          type="submit" class="isSaveEnabled" data-cy="btn-save"
          :label="$t('globals.buttons.save')" />
      </div>

      <section class="wrap settings-wrap" v-if="form">
        <PvTabs class="settings-tabs" v-model:value="tab">
          <PvTabList>
            <PvTab value="0">{{ $t('settings.general.name') }}</PvTab><!-- general -->
            <PvTab value="1">{{ $t('settings.performance.name') }}</PvTab><!-- performance -->
            <PvTab value="2">{{ $t('settings.privacy.name') }}</PvTab><!-- privacy -->
            <PvTab value="3">{{ $t('settings.security.name') }}</PvTab><!-- security -->
            <PvTab value="4">{{ $t('settings.media.title') }}</PvTab><!-- media -->
            <PvTab value="5">{{ $t('settings.smtp.name') }}</PvTab><!-- mail servers -->
            <PvTab value="6">{{ $t('settings.bounces.name') }}</PvTab><!-- bounces -->
            <PvTab value="7">{{ $t('settings.messengers.name') }}</PvTab><!-- messengers -->
            <PvTab value="8">{{ $t('settings.appearance.name') }}</PvTab><!-- appearance -->
            <PvTab value="9">{{ $t('settings.scrub.name') }}</PvTab><!-- mail validation -->
          </PvTabList>
          <PvTabPanels>
            <PvTabPanel value="0">
              <general-settings :form="form" :key="key" />
            </PvTabPanel>
            <PvTabPanel value="1">
              <performance-settings :form="form" :key="key" />
            </PvTabPanel>
            <PvTabPanel value="2">
              <privacy-settings :form="form" :key="key" />
            </PvTabPanel>
            <PvTabPanel value="3">
              <security-settings :form="form" :key="key" />
            </PvTabPanel>
            <PvTabPanel value="4">
              <media-settings :form="form" :key="key" />
            </PvTabPanel>
            <PvTabPanel value="5">
              <smtp-settings :form="form" :key="key" />
            </PvTabPanel>
            <PvTabPanel value="6">
              <bounce-settings :form="form" :key="key" />
            </PvTabPanel>
            <PvTabPanel value="7">
              <messenger-settings :form="form" :key="key" />
            </PvTabPanel>
            <PvTabPanel value="8">
              <appearance-settings :form="form" :key="key" />
            </PvTabPanel>
            <PvTabPanel value="9">
              <scrub-settings :form="form" :key="key" />
            </PvTabPanel>
          </PvTabPanels>
        </PvTabs>
      </section>
    </section>
  </form>
</template>

<script>
import { mapState } from 'pinia';
import { useMainStore } from '../store';
import AppearanceSettings from './settings/appearance.vue';
import ScrubSettings from './settings/scrub.vue';
import BounceSettings from './settings/bounces.vue';
import GeneralSettings from './settings/general.vue';
import MediaSettings from './settings/media.vue';
import MessengerSettings from './settings/messengers.vue';
import PerformanceSettings from './settings/performance.vue';
import PrivacySettings from './settings/privacy.vue';
import SecuritySettings from './settings/security.vue';
import SmtpSettings from './settings/smtp.vue';

export default {
  components: {
    GeneralSettings,
    PerformanceSettings,
    PrivacySettings,
    SecuritySettings,
    MediaSettings,
    SmtpSettings,
    BounceSettings,
    MessengerSettings,
    AppearanceSettings,
    ScrubSettings,
  },

  data() {
    return {
      // :key="key" is a ack to re-render child components every time settings
      // is pulled. Otherwise, props don't react.
      key: 0,

      isLoading: false,

      // formCopy is a stringified copy of the original settings against which
      // form is compared to detect changes.
      formCopy: '',
      form: null,
      tab: '0',
    };
  },

  methods: {
    async onSubmit() {
      const form = JSON.parse(JSON.stringify(this.form));

      // SMTP boxes.
      let hasDummy = '';
      for (let i = 0; i < form.smtp.length; i += 1) {
        // trim the host before saving
        form.smtp[i].host = form.smtp[i].host?.trim();

        // If it's the dummy UI password placeholder, ignore it.
        if (this.isDummy(form.smtp[i].password)) {
          form.smtp[i].password = '';
        } else if (this.hasDummy(form.smtp[i].password)) {
          hasDummy = `smtp #${i + 1}`;
        }

        if (form.smtp[i].strEmailHeaders && form.smtp[i].strEmailHeaders !== '[]') {
          form.smtp[i].email_headers = JSON.parse(form.smtp[i].strEmailHeaders);
        } else {
          form.smtp[i].email_headers = [];
        }
      }

      // Bounces boxes.
      for (let i = 0; i < form['bounce.mailboxes'].length; i += 1) {
        // trim the host before saving
        form['bounce.mailboxes'][i].host = form['bounce.mailboxes'][i].host?.trim();

        // If it's the dummy UI password placeholder, ignore it.
        if (this.isDummy(form['bounce.mailboxes'][i].password)) {
          form['bounce.mailboxes'][i].password = '';
        } else if (this.hasDummy(form['bounce.mailboxes'][i].password)) {
          hasDummy = `bounce #${i + 1}`;
        }
      }

      if (this.isDummy(form['upload.s3.aws_secret_access_key'])) {
        form['upload.s3.aws_secret_access_key'] = '';
      } else if (this.hasDummy(form['upload.s3.aws_secret_access_key'])) {
        hasDummy = 's3';
      }

      if (this.isDummy(form['bounce.sendgrid_key'])) {
        form['bounce.sendgrid_key'] = '';
      } else if (this.hasDummy(form['bounce.sendgrid_key'])) {
        hasDummy = 'sendgrid';
      }

      if (this.isDummy(form['bounce.azure'].shared_secret)) {
        form['bounce.azure'].shared_secret = '';
      } else if (this.hasDummy(form['bounce.azure'].shared_secret)) {
        hasDummy = 'azure shared secret';
      }

      if (this.isDummy(form['security.captcha'].hcaptcha.secret)) {
        form['security.captcha'].hcaptcha.secret = '';
      } else if (this.hasDummy(form['security.captcha'].hcaptcha.secret)) {
        hasDummy = 'captcha';
      }

      if (this.isDummy(form['security.oidc'].client_secret)) {
        form['security.oidc'].client_secret = '';
      } else if (this.hasDummy(form['security.oidc'].client_secret)) {
        hasDummy = 'oidc';
      }

      if (this.isDummy(form['bounce.postmark'].password)) {
        form['bounce.postmark'].password = '';
      } else if (this.hasDummy(form['bounce.postmark'].password)) {
        hasDummy = 'postmark';
      }

      if (this.isDummy(form['bounce.forwardemail'].key)) {
        form['bounce.forwardemail'].key = '';
      } else if (this.hasDummy(form['bounce.forwardemail'].key)) {
        hasDummy = 'forwardemail';
      }

      if (this.isDummy(form['bounce.lettermint'].key)) {
        form['bounce.lettermint'].key = '';
      } else if (this.hasDummy(form['bounce.lettermint'].key)) {
        hasDummy = 'lettermint';
      }

      if (this.isDummy(form.scrub.api_key)) {
        form.scrub.api_key = '';
      } else if (this.hasDummy(form.scrub.api_key)) {
        hasDummy = 'scrub';
      }

      for (let i = 0; i < form.messengers.length; i += 1) {
        // If it's the dummy UI password placeholder, ignore it.
        if (this.isDummy(form.messengers[i].password)) {
          form.messengers[i].password = '';
        } else if (this.hasDummy(form.messengers[i].password)) {
          hasDummy = `messenger #${i + 1}`;
        }
      }

      if (hasDummy) {
        this.$utils.toast(this.$t('globals.messages.passwordChangeFull', { name: hasDummy }), 'is-danger');
        return false;
      }

      // Domain blocklist array from multi-line strings.
      form['privacy.domain_blocklist'] = form['privacy.domain_blocklist'].split('\n').map((v) => v.trim().toLowerCase()).filter((v) => v !== '');
      form['privacy.domain_allowlist'] = form['privacy.domain_allowlist'].split('\n').map((v) => v.trim().toLowerCase()).filter((v) => v !== '');

      this.isLoading = true;
      try {
        const data = await this.$api.updateSettings(form);
        await this.$root.awaitRestart(data);
        this.getSettings();
      } finally {
        this.isLoading = false;
      }

      return false;
    },

    getSettings() {
      this.isLoading = true;
      this.$api.getSettings().then((data) => {
        let d = {};
        try {
          // Create a deep-copy of the settings hierarchy.
          d = JSON.parse(JSON.stringify(data));
        } catch (err) {
          return;
        }

        // Serialize the `email_headers` array map to display on the form.
        for (let i = 0; i < d.smtp.length; i += 1) {
          d.smtp[i].strEmailHeaders = JSON.stringify(d.smtp[i].email_headers, null, 4);
        }

        // Domain blocklist array to multi-line string.
        d['privacy.domain_blocklist'] = d['privacy.domain_blocklist'].join('\n');
        d['privacy.domain_allowlist'] = d['privacy.domain_allowlist'].join('\n');

        this.key += 1;
        this.form = d;
        this.formCopy = JSON.stringify(d);

        this.$nextTick(() => {
          this.isLoading = false;
        });
      });
    },

    isDummy(pwd) {
      return !pwd || (pwd.match(/•/g) || []).length === pwd.length;
    },

    hasDummy(pwd) {
      return pwd.includes('•');
    },
  },

  computed: {
    ...mapState(useMainStore, ['serverConfig', 'loading']),

    hasFormChanged() {
      if (!this.formCopy) {
        return false;
      }
      return JSON.stringify(this.form) !== this.formCopy;
    },
  },

  beforeRouteLeave(to, from, next) {
    if (this.hasFormChanged) {
      this.$utils.confirm(this.$t('globals.messages.confirmDiscard'), () => next(true));
      return;
    }
    next(true);
  },

  mounted() {
    this.tab = String(this.$utils.getPref('settings.tab') || '0');
    this.getSettings();
  },

  watch: {
    tab(t) {
      this.$utils.setPref('settings.tab', t);
    },
  },
};
</script>
