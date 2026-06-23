<template>
  <section class="dashboard content">
    <header class="columns">
      <div class="column is-two-thirds">
        <h1 class="title is-5">
          {{ $utils.niceDate(new Date()) }}
        </h1>
      </div>
    </header>

    <section class="counts wrap">
      <div class="tile is-ancestor">
        <div class="tile is-vertical is-12">
          <div class="tile">
            <div class="tile is-parent is-vertical relative">
              <b-loading v-if="isCountsLoading" active :is-full-page="false" />
              <article class="tile is-child notification" data-cy="lists">
                <div class="columns is-mobile">
                  <div class="column is-6">
                    <p class="title">
                      <b-icon icon="format-list-bulleted-square" />
                      {{ $utils.niceNumber(counts.lists.total) }}
                    </p>
                    <p class="is-size-6 has-text-grey">
                      {{ $tc('globals.terms.list', counts.lists.total) }}
                    </p>
                  </div>
                  <div class="column is-6">
                    <ul class="no has-text-grey">
                      <li>
                        <label for="#">{{ $utils.niceNumber(counts.lists.public) }}</label>
                        {{ $t('lists.types.public') }}
                      </li>
                      <li>
                        <label for="#">{{ $utils.niceNumber(counts.lists.private) }}</label>
                        {{ $t('lists.types.private') }}
                      </li>
                      <li>
                        <label for="#">{{ $utils.niceNumber(counts.lists.optinSingle) }}</label>
                        {{ $t('lists.optins.single') }}
                      </li>
                      <li>
                        <label for="#">{{ $utils.niceNumber(counts.lists.optinDouble) }}</label>
                        {{ $t('lists.optins.double') }}
                      </li>
                    </ul>
                  </div>
                </div>
              </article><!-- lists -->

              <article class="tile is-child notification" data-cy="campaigns">
                <div class="columns is-mobile">
                  <div class="column is-6">
                    <p class="title">
                      <b-icon icon="rocket-launch-outline" />
                      {{ $utils.niceNumber(counts.campaigns.total) }}
                    </p>
                    <p class="is-size-6 has-text-grey">
                      {{ $tc('globals.terms.campaign', counts.campaigns.total) }}
                    </p>
                  </div>
                  <div class="column is-6">
                    <ul class="no has-text-grey">
                      <li v-for="(num, status) in counts.campaigns.byStatus" :key="status">
                        <label for="#" :data-cy="`campaigns-${status}`">{{ num }}</label>
                        {{ $t(`campaigns.status.${status}`) }}
                        <span v-if="status === 'running'" class="spinner is-tiny">
                          <b-loading :is-full-page="false" active />
                        </span>
                      </li>
                    </ul>
                  </div>
                </div>
              </article><!-- campaigns -->

              <!-- Mail Validation (Scrub) widget -->
              <article class="tile is-child notification scrub-widget" v-if="settings && settings.scrub">
                <div class="scrub-header columns is-mobile is-vcentered mb-2">
                  <div class="column">
                    <span class="is-size-6 has-text-weight-semibold">
                      <b-icon :icon="scrubStatus === 'active' ? 'email-check-outline' : 'email-off-outline'"
                        size="is-small" class="mr-1" />
                      {{ $t('settings.scrub.name') }}
                    </span>
                  </div>
                  <div class="column is-narrow">
                    <b-tag :type="scrubStatusType" size="is-small">
                      {{ $t(`dashboard.scrub.${scrubStatus}`) }}
                    </b-tag>
                  </div>
                  <div class="column is-narrow">
                    <router-link to="/settings" class="has-text-grey is-size-7">
                      {{ $t('dashboard.scrub.configure') }} →
                    </router-link>
                  </div>
                </div>

                <!-- Usage breakdown (last 30 days) -->
                <div v-if="scrubData && scrubData.usage" class="scrub-breakdown">
                  <p class="is-size-7 has-text-grey mb-1">{{ $t('dashboard.scrub.last30Days') }}</p>
                  <div class="columns is-mobile is-gapless scrub-breakdown-cols">
                    <div class="column has-text-centered">
                      <p class="scrub-count has-text-success">{{ $utils.niceNumber(scrubData.usage.deliverable) }}</p>
                      <p class="is-size-7 has-text-grey">{{ $t('dashboard.scrub.deliverable') }}</p>
                    </div>
                    <div class="column has-text-centered">
                      <p class="scrub-count has-text-warning">{{ $utils.niceNumber(scrubData.usage.risky) }}</p>
                      <p class="is-size-7 has-text-grey">{{ $t('dashboard.scrub.risky') }}</p>
                    </div>
                    <div class="column has-text-centered">
                      <p class="scrub-count has-text-danger">
                        {{ $utils.niceNumber(scrubData.usage.undeliverable_syntax + scrubData.usage.undeliverable_domain) }}
                      </p>
                      <p class="is-size-7 has-text-grey">{{ $t('dashboard.scrub.undeliverable') }}</p>
                    </div>
                    <div class="column has-text-centered">
                      <p class="scrub-count has-text-grey">{{ $utils.niceNumber(scrubData.usage.total) }}</p>
                      <p class="is-size-7 has-text-grey">{{ $t('dashboard.scrub.total') }}</p>
                    </div>
                  </div>
                </div>

                <!-- Active jobs -->
                <div v-if="scrubData && scrubData.active_jobs && scrubData.active_jobs.length > 0"
                  class="scrub-active-jobs mt-3">
                  <p class="is-size-7 has-text-grey mb-1">
                    <b-icon icon="progress-clock" size="is-small" />
                    {{ $t('dashboard.scrub.activeJobs') }}
                    <b-tag type="is-warning is-light" size="is-small" rounded>
                      {{ scrubData.active_jobs.length }}
                    </b-tag>
                  </p>
                  <div v-for="job in scrubData.active_jobs" :key="job.request_id" class="scrub-job mb-2">
                    <div class="columns is-mobile is-vcentered is-gapless">
                      <div class="column">
                        <span class="is-size-7">
                          {{ job.list_name || job.job_type }}
                        </span>
                      </div>
                      <div class="column is-narrow">
                        <span class="is-size-7 has-text-grey">{{ job.progress_percentage }}%</span>
                      </div>
                    </div>
                    <b-progress :value="job.progress_percentage" size="is-small"
                      :type="job.status === 'running' ? 'is-info' : 'is-warning'" />
                  </div>
                </div>

                <!-- No data placeholder when configured but Scrub unreachable -->
                <p v-if="scrubData && scrubData.configured && !scrubData.usage"
                  class="is-size-7 has-text-grey mt-2">
                  {{ $t('dashboard.scrub.unavailable') }}
                </p>
              </article><!-- mail validation -->
            </div><!-- block -->

            <div class="tile is-parent relative">
              <b-loading v-if="isCountsLoading" active :is-full-page="false" />
              <article class="tile is-child notification" data-cy="subscribers">
                <div class="columns is-mobile">
                  <div class="column is-6">
                    <p class="title">
                      <b-icon icon="account-multiple" />
                      {{ $utils.niceNumber(counts.subscribers.total) }}
                    </p>
                    <p class="is-size-6 has-text-grey">
                      {{ $tc('globals.terms.subscriber', counts.subscribers.total) }}
                    </p>
                  </div>

                  <div class="column is-6">
                    <ul class="no has-text-grey">
                      <li>
                        <label for="#">{{ $utils.niceNumber(counts.subscribers.blocklisted) }}</label>
                        {{ $t('subscribers.status.blocklisted') }}
                      </li>
                      <li>
                        <label for="#">{{ $utils.niceNumber(counts.subscribers.orphans) }}</label>
                        {{ $t('dashboard.orphanSubs') }}
                      </li>
                    </ul>
                  </div><!-- subscriber breakdown -->
                </div><!-- subscriber columns -->
                <hr />
                <div class="columns" data-cy="messages">
                  <div class="column is-12">
                    <p class="title">
                      <b-icon icon="email-outline" />
                      {{ $utils.niceNumber(counts.messages) }}
                    </p>
                    <p class="is-size-6 has-text-grey">
                      {{ $t('dashboard.messagesSent') }}
                    </p>
                  </div>
                </div>
              </article><!-- subscribers -->
            </div>
          </div>
          <div class="tile is-parent relative">
            <b-loading v-if="isChartsLoading" active :is-full-page="false" />
            <article class="tile is-child notification charts">
              <div class="columns">
                <div class="column is-6">
                  <h3 class="title is-size-6">
                    {{ $t('dashboard.campaignViews') }}
                  </h3><br />
                  <chart type="line" v-if="campaignViews" :data="campaignViews" />
                </div>
                <div class="column is-6">
                  <h3 class="title is-size-6 has-text-right">
                    {{ $t('dashboard.linkClicks') }}
                  </h3><br />
                  <chart type="line" v-if="campaignClicks" :data="campaignClicks" />
                </div>
              </div>
            </article>
          </div>
        </div>
      </div><!-- tile block -->
      <p v-if="settings['app.cache_slow_queries']" class="has-text-grey">
        *{{ $t('globals.messages.slowQueriesCached') }}
        <a href="https://listmonk.app/docs/maintenance/performance/" target="_blank" rel="noopener noreferer"
          class="has-text-grey">
          <b-icon icon="link-variant" /> {{ $t('globals.buttons.learnMore') }}
        </a>
      </p>
    </section>
  </section>
</template>

<script>
import dayjs from 'dayjs';
import Vue from 'vue';
import { mapState } from 'vuex';
import { colors } from '../constants';
import Chart from '../components/Chart.vue';

export default Vue.extend({
  components: {
    Chart,
  },

  data() {
    return {
      isChartsLoading: true,
      isCountsLoading: true,
      campaignViews: null,
      campaignClicks: null,
      scrubData: null,
      counts: {
        lists: {},
        subscribers: {},
        campaigns: {},
        messages: 0,
      },
    };
  },

  methods: {
    fetchData() {
      this.isCountsLoading = true;
      this.isChartsLoading = true;

      this.$api.getDashboardCounts().then((data) => {
        this.counts = data;
        this.isCountsLoading = false;
      });

      this.$api.getDashboardCharts().then((data) => {
        this.isChartsLoading = false;
        this.campaignViews = this.makeChart(data.campaignViews);
        this.campaignClicks = this.makeChart(data.linkClicks);
      });

      // Load settings for the Scrub widget visibility check.
      if (!this.settings || !this.settings.scrub) {
        this.$api.getSettings();
      }

      // Fetch Scrub status (usage breakdown + active jobs).
      this.$api.getScrubStatus().then((data) => {
        this.scrubData = data;
      }).catch(() => {});
    },

    makeChart(data) {
      if (data.length === 0) {
        return {};
      }
      return {
        labels: data.map((d) => dayjs(d.date).format('DD MMM')),
        datasets: [
          {
            data: [...data.map((d) => d.count)],
            borderColor: colors.primary,
            borderWidth: 2,
            pointHoverBorderWidth: 5,
            pointBorderWidth: 0.5,
          },
        ],
      };
    },
  },

  computed: {
    ...mapState(['settings']),
    dayjs() {
      return dayjs;
    },

    scrubStatus() {
      const s = this.settings && this.settings.scrub;
      if (!s || !s.enabled) return 'disabled';
      if (!s.url || !s.api_key) return 'notConfigured';
      return 'active';
    },

    scrubStatusType() {
      if (this.scrubStatus === 'active') return 'is-success';
      if (this.scrubStatus === 'notConfigured') return 'is-warning';
      return '';
    },
  },

  created() {
    this.$root.$on('page.refresh', this.fetchData);
  },

  destroyed() {
    this.$root.$off('page.refresh', this.fetchData);
  },

  mounted() {
    this.fetchData();
  },
});
</script>

<style scoped>
.scrub-count {
  font-size: 1.1rem;
  font-weight: 600;
  line-height: 1.2;
}
.scrub-breakdown-cols .column {
  border-right: 1px solid #f0f0f0;
}
.scrub-breakdown-cols .column:last-child {
  border-right: none;
}
.scrub-job .b-progress {
  margin-top: 2px;
}
</style>
