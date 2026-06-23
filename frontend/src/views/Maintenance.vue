<template>
  <section class="maintenance wrap">
    <h1 class="title is-4">
      {{ $t('maintenance.title') }}
    </h1>
    <hr />
    <p class="has-text-grey">
      {{ $t('maintenance.help') }}
    </p>
    <br />

    <div class="box">
      <h4 class="is-size-4">
        {{ $t('globals.terms.subscribers') }}
      </h4><br />
      <div class="columns">
        <div class="column is-4">
          <div class="field">
            <label class="block mb-1 text-sm font-medium">Data</label>
            <PvSelect v-model="subscriberType" :options="subscriberTypeOptions" option-label="label" option-value="value" class="w-full" />
            <small class="block mt-1 text-color-secondary">{{ $t('maintenance.orphanHelp') }}</small>
          </div>
        </div>
        <div class="column is-5" />
        <div class="column">
          <br />
          <PvButton severity="primary" :loading="loading.maintenance" @click="deleteSubscribers" class="w-full"
            :label="$t('globals.buttons.delete')" />
        </div>
      </div>
    </div><!-- subscribers -->

    <div class="box mt-6">
      <h4 class="is-size-4">
        {{ $tc('globals.terms.subscriptions', 2) }}
      </h4><br />
      <div class="columns">
        <div class="column is-4">
          <div class="field">
            <label class="block mb-1 text-sm font-medium">Data</label>
            <PvSelect v-model="subscriptionType" :options="subscriptionTypeOptions" option-label="label" option-value="value" class="w-full" />
          </div>
        </div>
        <div class="column is-4">
          <div class="field">
            <label class="block mb-1 text-sm font-medium">{{ $t('maintenance.olderThan') }}</label>
            <PvDatePicker v-model="subscriptionDate" required class="w-full" :date-format="'yy-mm-dd'" />
          </div>
        </div>
        <div class="column is-1" />
        <div class="column">
          <br />
          <PvButton severity="primary" :loading="loading.maintenance" @click="deleteSubscriptions" class="w-full"
            :label="$t('globals.buttons.delete')" />
        </div>
      </div>
    </div><!-- subscriptions -->

    <div class="box mt-6">
      <h4 class="is-size-4">
        {{ $t('globals.terms.analytics') }}
      </h4><br />
      <div class="columns">
        <div class="column is-4">
          <div class="field">
            <label class="block mb-1 text-sm font-medium">Data</label>
            <PvSelect v-model="analyticsType" :options="analyticsTypeOptions" option-label="label" option-value="value" class="w-full" />
          </div>
        </div>
        <div class="column is-4">
          <div class="field">
            <label class="block mb-1 text-sm font-medium">{{ $t('maintenance.olderThan') }}</label>
            <PvDatePicker v-model="analyticsDate" required class="w-full" :date-format="'yy-mm-dd'" />
          </div>
        </div>
        <div class="column is-1" />
        <div class="column">
          <br />
          <PvButton severity="primary" :loading="loading.maintenance" @click="deleteAnalytics" class="w-full"
            :label="$t('globals.buttons.delete')" />
        </div>
      </div>

      <hr />
      <h5 class="is-size-5">
        {{ $t('subscribers.export') }}
      </h5>
      <br />
      <div class="columns">
        <div class="column is-4">
          <div class="field">
            <label class="block mb-1 text-sm font-medium">Data</label>
            <PvSelect v-model="exportType" :options="exportTypeOptions" option-label="label" option-value="value" class="w-full" />
          </div>
        </div>
        <div class="column is-4">
          <div class="field">
            <label class="block mb-1 text-sm font-medium">{{ $t('analytics.fromDate') }}</label>
            <PvDatePicker v-model="exportDate" required class="w-full" :date-format="'yy-mm-dd'" />
          </div>
        </div>
        <div class="column is-1" />
        <div class="column">
          <br />
          <a :href="exportURL" class="w-full">
            <PvButton severity="primary" icon="pi pi-download" class="w-full"
              :label="$t('subscribers.export')" />
          </a>
        </div>
      </div>
    </div><!-- analytics -->

    <form @submit.prevent="onUpdateDBSettings" class="box mt-6">
      <h4 class="is-size-4">
        {{ $t('maintenance.database.title') }}
      </h4><br />
      <h5 class="is-size-5">Vacuum</h5>
      <p class="has-text-grey is-size-7">
        {{ $t('maintenance.database.vacuumHelp') }}
      </p>
      <br />
      <div class="columns">
        <div class="column is-2">
          <div class="field">
            <label class="block mb-1 text-sm font-medium">{{ $t('globals.buttons.enabled') }}</label>
            <div class="flex items-center gap-2">
              <PvToggleSwitch v-model="dbSettings.vacuum" />
            </div>
          </div>
        </div>
        <div class="column is-4" :class="{ disabled: !dbSettings.vacuum }">
          <div class="field">
            <label class="block mb-1 text-sm font-medium">{{ $t('settings.maintenance.cron') }}</label>
            <PvInputText v-model="dbSettings.vacuum_cron_interval" placeholder="0 2 * * *" :disabled="!dbSettings.vacuum"
              pattern="((\*|[0-9,\-\/]+)\s+){4}(\*|[0-9,\-\/]+)" class="w-full" />
          </div>
        </div>
        <div class="column is-3" />
        <div class="column is-3">
          <br />
          <PvButton severity="primary" type="submit" :loading="loading.settings" class="w-full"
            :label="$t('globals.buttons.save')" />
        </div>
      </div>
    </form><!-- database -->

    <div v-if="isLoading" class="flex justify-center p-8">
      <PvProgressSpinner />
    </div>
  </section>
</template>

<script>
import dayjs from 'dayjs';
import { mapState } from 'vuex';

export default {
  components: {
  },

  data() {
    return {
      isLoading: false,
      subscriberType: 'orphan',
      analyticsType: 'all',
      subscriptionType: 'optin',
      analyticsDate: dayjs().subtract(7, 'day').toDate(),
      subscriptionDate: dayjs().subtract(7, 'day').toDate(),
      exportType: 'views',
      exportDate: dayjs().subtract(30, 'day').toDate(),
      dbSettings: {
        vacuum: false,
        vacuum_cron_interval: '0 2 * * *',
      },
    };
  },

  mounted() {
    this.loadDBSettings();
  },

  methods: {
    formatDateTime(s) {
      return dayjs(s).format('YYYY-MM-DD');
    },

    deleteSubscribers() {
      this.$utils.confirm(
        null,
        () => {
          this.$api.deleteGCSubscribers(this.subscriberType).then((data) => {
            this.$utils.toast(this.$t(
              'globals.messages.deletedCount',
              { name: this.$tc('globals.terms.subscribers', 2), num: data.count },
            ));
          });
        },
      );
    },

    deleteSubscriptions() {
      this.$utils.confirm(
        null,
        () => {
          this.$api.deleteGCSubscriptions(this.subscriptionDate).then((data) => {
            this.$utils.toast(this.$t(
              'globals.messages.deletedCount',
              { name: this.$tc('globals.terms.subscriptions', 2), num: data.count },
            ));
          });
        },
      );
    },

    deleteAnalytics() {
      this.$utils.confirm(
        null,
        () => {
          this.$api.deleteGCCampaignAnalytics(this.analyticsType, this.analyticsDate)
            .then(() => {
              this.$utils.toast(this.$t('globals.messages.done'));
            });
        },
      );
    },

    loadDBSettings() {
      this.$api.getSettings().then((data) => {
        if (data['maintenance.db'] !== undefined) {
          this.dbSettings = { ...data['maintenance.db'] };
        }
      });
    },

    async onUpdateDBSettings() {
      this.isLoading = true;
      const data = await this.$api.updateSettingsByKey('maintenance.db', this.dbSettings);
      await this.$root.awaitRestart(data);
      this.isLoading = false;
    },
  },

  computed: {
    ...mapState(['loading']),

    exportURL() {
      const since = encodeURIComponent(dayjs(this.exportDate).toISOString());
      return `/api/maintenance/analytics/${this.exportType}/export?since=${since}`;
    },

    subscriberTypeOptions() {
      return [
        { label: this.$t('dashboard.orphanSubs'), value: 'orphan' },
        { label: this.$t('subscribers.status.blocklisted'), value: 'blocklisted' },
      ];
    },

    subscriptionTypeOptions() {
      return [
        { label: this.$t('maintenance.maintenance.unconfirmedOptins'), value: 'optin' },
      ];
    },

    analyticsTypeOptions() {
      return [
        { label: this.$t('globals.terms.all'), value: 'all' },
        { label: this.$t('dashboard.campaignViews'), value: 'views' },
        { label: this.$t('dashboard.linkClicks'), value: 'clicks' },
      ];
    },

    exportTypeOptions() {
      return [
        { label: this.$t('dashboard.campaignViews'), value: 'views' },
        { label: this.$t('dashboard.linkClicks'), value: 'clicks' },
      ];
    },
  },
};
</script>
