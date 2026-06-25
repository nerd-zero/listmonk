<template>
  <div class="analytics-page">
    <div class="page-header">
      <h1 class="page-title">{{ $t('analytics.title') }}</h1>
    </div>

    <div v-if="serverConfig.privacy.disable_tracking || !serverConfig.privacy.individual_tracking"
      class="analytics-notice">
      <i class="pi pi-info-circle" />
      <span v-if="serverConfig.privacy.disable_tracking">{{ $t('analytics.trackingDisabled') }}</span>
      <span v-else-if="!serverConfig.privacy.individual_tracking">{{ $t('analytics.nonIndividualTracking') }}</span>
    </div>

    <div class="table-card">
      <div class="analytics-filters">
        <form class="filter-form" @submit.prevent="onSubmit">
          <div class="filter-field">
            <label class="filter-label">{{ $t('globals.terms.campaigns') }}</label>
            <PvAutoComplete v-model="form.campaigns" :suggestions="queriedCampaigns" name="campaigns"
              :placeholder="$t('globals.terms.campaigns')" multiple option-label="name"
              :loading="isSearchLoading" @complete="(e) => queryCampaigns(e.query)"
              @focus="queryCampaigns('')" class="w-full" />
          </div>
          <div class="filter-field">
            <label class="filter-label">{{ $t('analytics.fromDate') }}</label>
            <PvDatePicker v-model="form.from" show-time hour-format="24"
              :date-format="'yy-mm-dd'" @date-select="onFromDateChange" @update:model-value="onFromDateChange" />
          </div>
          <div class="filter-field">
            <label class="filter-label">{{ $t('analytics.toDate') }}</label>
            <PvDatePicker v-model="form.to" show-time hour-format="24"
              :date-format="'yy-mm-dd'" @date-select="onToDateChange" @update:model-value="onToDateChange" />
          </div>
          <div class="filter-action">
            <PvButton type="submit" severity="primary" icon="pi pi-search" :disabled="form.campaigns.length === 0"
              data-cy="btn-search" />
          </div>
        </form>
      </div>

      <div v-if="hasChartData" class="charts-section">
        <div v-for="(v, k) in charts" :key="k" class="chart-block">
          <div class="chart-block-header">
            <span class="chart-block-title">{{ v.name }}</span>
            <span v-if="v.type !== 'bar' && v.data" class="chart-block-count">({{ $utils.niceNumber(counts[k]) }})</span>
          </div>
          <div v-if="v.loading" class="chart-spinner">
            <PvProgressSpinner style="width:2rem;height:2rem" />
          </div>
          <div v-else-if="v.data" class="chart-block-body" :class="{ 'chart-block-body--bar': v.type === 'bar' }">
            <div class="chart-main">
              <chart :type="v.type" :data="v.data" :on-click="v.onClick" />
            </div>
            <div v-if="v.type !== 'bar'" class="chart-donut">
              <chart type="donut" :data="v.donutData" />
            </div>
          </div>
          <div v-else class="chart-empty">
            <span class="chart-empty-text">{{ $t('globals.messages.emptyState') }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import dayjs from 'dayjs';
import { mapState } from 'pinia';
import { useMainStore } from '../store';
import { colors } from '../constants';
import Chart from '../components/Chart.vue';

const chartColorRed = '#ee7d5b';
const chartColors = [
  colors.primary,
  '#FFB50D',
  '#41AC9C',
  chartColorRed,
  '#7FC7BC',
  '#3a82d6',
  '#688ED9',
  '#FFC43D',
];

export default {
  components: {
    Chart,
  },

  data() {
    return {
      isSearchLoading: false,
      queriedCampaigns: [],

      // Data for each view.
      counts: {
        views: 0,
        clicks: 0,
        bounces: 0,
        links: 0,
      },
      urls: [],
      charts: {
        views: {
          name: this.$t('campaigns.views'),
          type: 'line',
          data: null,
          fn: this.$api.getCampaignViewCounts,
          chartFn: this.makeCharts,
          loading: false,
        },

        clicks: {
          name: this.$t('campaigns.clicks'),
          type: 'line',
          data: null,
          fn: this.$api.getCampaignClickCounts,
          chartFn: this.makeCharts,
          loading: false,
        },

        bounces: {
          name: this.$t('globals.terms.bounces'),
          type: 'line',
          data: null,
          fn: this.$api.getCampaignBounceCounts,
          chartFn: this.makeCharts,
          donutColor: chartColorRed,
          loading: false,
        },

        links: {
          name: this.$t('analytics.links'),
          type: 'bar',
          data: null,
          loading: false,
          fn: this.$api.getCampaignLinkCounts,
          chartFn: this.makeLinksChart,
          onClick: this.onLinkClick,
        },
      },

      form: {
        campaigns: [],
        from: null,
        to: null,
      },
    };
  },

  methods: {
    onFromDateChange() {
      if (this.form.from > this.form.to) {
        this.form.to = dayjs(this.form.from).add(7, 'day').toDate();
      }
    },

    onToDateChange() {
      if (this.form.from > this.form.to) {
        this.form.from = dayjs(this.form.to).add(-7, 'day').toDate();
      }
    },

    formatDateTime(s) {
      return dayjs(s).format('YYYY-MM-DD HH:mm');
    },

    isCampaignSelected(camp) {
      return !this.form.campaigns.find(({ id }) => id === camp.id);
    },

    makeLinksChart(typ, camps, data) {
      const labels = data.map((l) => {
        try {
          this.urls.push(l.url);
          const u = new URL(l.url);
          if (l.url.length > 80) {
            return `${u.hostname}${u.pathname.substr(0, 50)}..`;
          }
          return u.hostname + u.pathname;
        } catch {
          return l.url;
        }
      });

      const out = {
        labels,
        datasets: [
          {
            data: data.map((l) => l.count),
            backgroundColor: chartColors,
          }],
      };

      return { points: out, donut: null };
    },

    makeCharts(typ, campaigns, data) {
      // Make a campaign id => camp lookup map to group incoming
      // data by campaigns.
      const camps = campaigns.reduce((obj, c) => {
        const out = { ...obj };
        out[c.id] = c;
        return out;
      }, {});
      const campIDs = Object.keys(camps);
      // datasets[] array for line chart.
      const lines = campIDs.map((id, n) => {
        const cId = parseInt(id, 10);
        const points = data.filter((item) => item.campaignId === cId);

        return {
          label: camps[id].name,
          data: points.map((item) => ({ x: this.formatDateTime(item.timestamp), y: item.count })),
          borderColor: chartColors[n % chartColors.length],
          borderWidth: 2,
          pointHoverBorderWidth: 5,
          pointBorderWidth: 0.5,
        };
      });

      // Donut.
      const labels = [];
      const points = campIDs.map((id) => {
        labels.push(camps[id].name);
        const cId = parseInt(id, 10);
        const sum = data.reduce((a, item) => (item.campaignId === cId ? a + item.count : a), 0);
        return sum;
      });

      const donut = {
        labels,
        datasets: [{
          data: points, backgroundColor: chartColors, borderWidth: 6,
        }],
      };
      return { points: { datasets: lines }, donut };
    },

    onSubmit() {
      this.$router.push({ query: { id: this.form.campaigns.map((c) => c.id), from: dayjs(this.form.from).unix(), to: dayjs(this.form.to).unix() } });
    },

    queryCampaigns(q) {
      this.isSearchLoading = true;
      this.$api.getCampaigns({
        query: q,
        order_by: 'created_at',
        order: 'DESC',
      }).then((data) => {
        this.isSearchLoading = false;
        this.queriedCampaigns = data.results.map((c) => {
          // Change the name to include the ID in the auto-suggest results.
          const camp = c;
          camp.name = `#${c.id}: ${c.name}`;
          return camp;
        });
      });
    },

    getData(typ, camps) {
      this.charts[typ].loading = true;
      // Call the HTTP API.
      this.charts[typ].fn({
        id: camps.map((c) => c.id),
        from: this.form.from,
        to: this.form.to,
      }).then((data) => {
        // Set the total count.
        this.counts[typ] = data.reduce((sum, d) => sum + d.count, 0);

        const { points, donut } = this.charts[typ].chartFn(typ, camps, data);
        this.charts[typ].data = points;
        this.charts[typ].donutData = donut;
        this.charts[typ].loading = false;
      });
    },

    onLinkClick(e) {
      const bars = e.chart.getElementsAtEventForMode(e, 'nearest', { intersect: true }, true);
      if (bars.length > 0) {
        window.open(this.urls[bars[0].index], '_blank', 'noopener noreferrer');
      }
    },
  },

  computed: {
    ...mapState(useMainStore, ['serverConfig']),

    hasChartData() {
      return Object.values(this.charts).some((c) => c.data !== null || c.loading);
    },
  },

  created() {
    const now = dayjs().set('hour', 23).set('minute', 59).set('seconds', 0);
    const weekAgo = now.subtract(7, 'day').set('hour', 0).set('minute', 0);
    const from = this.$route.query.from ? dayjs.unix(this.$route.query.from) : weekAgo;
    const to = this.$route.query.to ? dayjs.unix(this.$route.query.to) : now;
    this.form.from = from.toDate();
    this.form.to = to.toDate();
  },

  mounted() {
    // Fetch one or more campaigns if there are ?id params, wait for the fetches
    // to finish, add them to the campaign selector and submit the form.
    const ids = this.$utils.parseQueryIDs(this.$route.query.id);
    if (ids.length > 0) {
      this.isSearchLoading = true;
      Promise.allSettled(ids.map((id) => this.$api.getCampaign(id))).then((data) => {
        data.forEach((d) => {
          if (d.status !== 'fulfilled') {
            return;
          }

          const camp = d.value;
          camp.name = `#${camp.id}: ${camp.name}`;
          this.form.campaigns.push(camp);
        });

        this.$nextTick(() => {
          this.isSearchLoading = false;

          // Fetch count for each analytics type (views, counts, bounces);
          Object.keys(this.charts).forEach((k) => {
            this.charts[k].data = null;
            this.charts[k].donutData = null;

            // Fetch views, clicks, bounces for every campaign.
            this.getData(k, this.form.campaigns);
          });
        });
      });
    }
  },
};
</script>

<style scoped lang="scss">
.analytics-page { display: flex; flex-direction: column; gap: 1.5rem; }

.analytics-notice {
  display: flex; align-items: center; gap: 0.5rem;
  padding: 0.75rem 1rem; background: var(--lm-primary-light); border: 1px solid var(--lm-primary-border);
  border-radius: 8px; font-size: 0.875rem; color: var(--lm-primary);
  i { font-size: 1rem; }
}

.analytics-filters { padding: 1.25rem 1.5rem; border-bottom: 1px solid var(--lm-bg-subtle); }
.filter-form { display: flex; align-items: flex-end; gap: 1rem; flex-wrap: wrap; }
.filter-field { display: flex; flex-direction: column; gap: 0.3rem; flex: 1; min-width: 180px; }
.filter-label { font-size: 0.8rem; font-weight: 600; color: var(--lm-text); }
.filter-action { padding-bottom: 0; align-self: flex-end; }

.charts-section { display: flex; flex-direction: column; }
.chart-block { padding: 1.25rem 1.5rem; & + & { border-top: 1px solid var(--lm-border); } }
.chart-block-header { display: flex; align-items: center; gap: 0.5rem; margin-bottom: 0.75rem; }
.chart-block-title { font-size: 0.95rem; font-weight: 600; color: var(--lm-text); }
.chart-block-count { font-size: 0.85rem; color: var(--lm-text-subtle); }

.chart-block-body {
  display: grid;
  grid-template-columns: 1fr 200px;
  gap: 1.5rem;
  align-items: center;
  &--bar { grid-template-columns: 1fr; }
}
.chart-main { height: 220px; position: relative; }
.chart-donut { height: 200px; position: relative; }

.chart-spinner { display: flex; justify-content: center; padding: 3rem; }
.chart-empty { display: flex; justify-content: center; padding: 2rem; }
.chart-empty-text { font-size: 0.875rem; color: var(--lm-text-subtle); }
</style>
