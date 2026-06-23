<template>
  <div>
    <div class="block box">
      <div class="columns">
        <div class="column is-2">
          <b-field>
            <b-switch v-model="data.scrub.enabled" name="scrub.enabled">
              {{ $t('globals.buttons.enabled') }}
            </b-switch>
          </b-field>
        </div>

        <div class="column" :class="{ disabled: !data.scrub.enabled }">
          <b-field :label="$t('settings.scrub.url')" label-position="on-border"
            :message="$t('settings.scrub.urlHelp')">
            <b-input v-model="data.scrub.url" name="scrub.url"
              placeholder="https://api.thescrub.app" :maxlength="300"
              :disabled="!data.scrub.enabled" />
          </b-field>

          <b-field :label="$t('settings.scrub.apiKey')" label-position="on-border"
            :message="$t('settings.scrub.apiKeyHelp')">
            <b-input v-model="data.scrub.api_key" name="scrub.api_key"
              type="password" :maxlength="300"
              :placeholder="$t('settings.scrub.apiKeyPlaceholder')"
              :disabled="!data.scrub.enabled" />
          </b-field>

          <b-field>
            <b-button type="is-primary" :loading="isTesting"
              :disabled="!data.scrub.enabled || !data.scrub.url || !data.scrub.api_key"
              icon-left="connection" @click="testConnection">
              {{ $t('settings.scrub.testConnection') }}
            </b-button>
          </b-field>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import Vue from 'vue';

export default Vue.extend({
  props: {
    form: {
      type: Object,
      default: () => {},
    },
  },

  data() {
    return {
      data: this.form,
      isTesting: false,
    };
  },

  methods: {
    async testConnection() {
      this.isTesting = true;
      try {
        await this.$api.testScrub({
          url: this.data.scrub.url,
          api_key: this.data.scrub.api_key,
        });
        this.$utils.toast(this.$t('settings.scrub.testSuccess'), 'is-success');
      } catch (e) {
        this.$utils.toast(
          e.response?.data?.message || this.$t('settings.scrub.testError'),
          'is-danger',
        );
      } finally {
        this.isTesting = false;
      }
    },
  },
});
</script>
