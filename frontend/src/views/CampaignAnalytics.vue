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

<script setup lang="ts">
import {
  ref, reactive, computed, nextTick, onMounted,
} from 'vue';
import dayjs from 'dayjs';
import { storeToRefs } from 'pinia';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';
import { useMainStore } from '../store';
import { useGlobal } from '../composables/useGlobal';
import { colors } from '../constants';
import Chart from '../components/Chart.vue';

const { $api, $utils } = useGlobal();
const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const { serverConfig } = storeToRefs(useMainStore());

const chartColorRed = '#ee7d5b';
const chartColors = [colors.primary, '#FFB50D', '#41AC9C', chartColorRed, '#7FC7BC', '#3a82d6', '#688ED9', '#FFC43D'];

const isSearchLoading = ref(false);
const queriedCampaigns = ref<any[]>([]);
const urls = ref<string[]>([]);
const counts = reactive({
  views: 0, clicks: 0, bounces: 0, links: 0,
});
const form = reactive({ campaigns: [] as any[], from: null as any, to: null as any });

function formatDateTime(s: any) { return dayjs(s).format('YYYY-MM-DD HH:mm'); }

function onLinkClick(e: any) {
  const bars = e.chart.getElementsAtEventForMode(e, 'nearest', { intersect: true }, true);
  if (bars.length > 0) window.open(urls.value[bars[0].index], '_blank', 'noopener noreferrer');
}

function makeLinksChart(_typ: string, _camps: any[], data: any[]) {
  const labels = data.map((l) => {
    try {
      urls.value.push(l.url);
      const u = new URL(l.url);
      return l.url.length > 80 ? `${u.hostname}${u.pathname.substr(0, 50)}..` : u.hostname + u.pathname;
    } catch { return l.url; }
  });
  return { points: { labels, datasets: [{ data: data.map((l) => l.count), backgroundColor: chartColors }] }, donut: null };
}

function makeCharts(_typ: string, campaigns: any[], data: any[]) {
  const camps = campaigns.reduce((obj: any, c: any) => ({ ...obj, [c.id]: c }), {});
  const campIDs = Object.keys(camps);
  const lines = campIDs.map((id, n) => {
    const cId = parseInt(id, 10);
    const points = data.filter((item) => item.campaignId === cId);
    return {
      label: camps[id].name,
      data: points.map((item: any) => ({ x: formatDateTime(item.timestamp), y: item.count })),
      borderColor: chartColors[n % chartColors.length],
      borderWidth: 2,
      pointHoverBorderWidth: 5,
      pointBorderWidth: 0.5,
    };
  });
  const labels: string[] = [];
  const points = campIDs.map((id) => {
    labels.push(camps[id].name);
    const cId = parseInt(id, 10);
    return data.reduce((a: number, item: any) => (item.campaignId === cId ? a + item.count : a), 0);
  });
  return { points: { datasets: lines }, donut: { labels, datasets: [{ data: points, backgroundColor: chartColors, borderWidth: 6 }] } };
}

const charts = reactive<Record<string, any>>({
  views: {
    name: '', type: 'line', data: null, loading: false, apiFn: 'getCampaignViewCounts', chartFnName: 'makeCharts',
  },
  clicks: {
    name: '', type: 'line', data: null, loading: false, apiFn: 'getCampaignClickCounts', chartFnName: 'makeCharts',
  },
  bounces: {
    name: '', type: 'line', data: null, loading: false, donutColor: chartColorRed, apiFn: 'getCampaignBounceCounts', chartFnName: 'makeCharts',
  },
  links: {
    name: '', type: 'bar', data: null, loading: false, apiFn: 'getCampaignLinkCounts', chartFnName: 'makeLinksChart', onClick: onLinkClick,
  },
});

const hasChartData = computed(() => Object.values(charts).some((c) => c.data !== null || c.loading));

function getData(typ: string, camps: any[]) {
  charts[typ].loading = true;
  $api[charts[typ].apiFn]({ id: camps.map((c: any) => c.id), from: form.from, to: form.to }).then((data: any) => {
    (counts as any)[typ] = data.reduce((sum: number, d: any) => sum + d.count, 0);
    const chartFn = charts[typ].chartFnName === 'makeCharts' ? makeCharts : makeLinksChart;
    const { points, donut } = chartFn(typ, camps, data);
    charts[typ].data = points;
    charts[typ].donutData = donut;
    charts[typ].loading = false;
  });
}

function onFromDateChange() {
  if (form.from > form.to) form.to = dayjs(form.from).add(7, 'day').toDate();
}

function onToDateChange() {
  if (form.from > form.to) form.from = dayjs(form.to).add(-7, 'day').toDate();
}

function queryCampaigns(q: string) {
  isSearchLoading.value = true;
  $api.getCampaigns({ query: q, order_by: 'created_at', order: 'DESC' }).then((data: any) => {
    isSearchLoading.value = false;
    queriedCampaigns.value = data.results.map((c: any) => ({ ...c, name: `#${c.id}: ${c.name}` }));
  });
}

function onSubmit() {
  router.push({ query: { id: form.campaigns.map((c: any) => c.id), from: dayjs(form.from).unix(), to: dayjs(form.to).unix() } });
}

onMounted(() => {
  const { t: tFn } = { t };
  charts.views.name = t('campaigns.views');
  charts.clicks.name = t('campaigns.clicks');
  charts.bounces.name = t('globals.terms.bounces');
  charts.links.name = t('analytics.links');

  const now = dayjs().set('hour', 23).set('minute', 59).set('seconds', 0);
  const weekAgo = now.subtract(7, 'day').set('hour', 0).set('minute', 0);
  form.from = (route.query.from ? dayjs.unix(Number(route.query.from)) : weekAgo).toDate();
  form.to = (route.query.to ? dayjs.unix(Number(route.query.to)) : now).toDate();

  const ids = $utils.parseQueryIDs(route.query.id);
  if (ids.length > 0) {
    isSearchLoading.value = true;
    Promise.allSettled(ids.map((id: number) => $api.getCampaign(id))).then((data: any[]) => {
      data.forEach((d) => {
        if (d.status !== 'fulfilled') return;
        const camp = d.value;
        camp.name = `#${camp.id}: ${camp.name}`;
        form.campaigns.push(camp);
      });
      nextTick(() => {
        isSearchLoading.value = false;
        Object.keys(charts).forEach((k) => {
          charts[k].data = null;
          charts[k].donutData = null;
          getData(k, form.campaigns);
        });
      });
    });
  }
});
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
