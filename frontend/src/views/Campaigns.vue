<template>
  <section class="campaigns">
    <header class="grid page-header">
      <div class="col-10">
        <h1 class="title is-4">
          {{ $t('globals.terms.campaigns') }}
          <span v-if="!isNaN(campaigns.total)">({{ campaigns.total }})</span>
        </h1>
      </div>
      <div class="col has-text-right">
        <div v-if="$can('campaigns:manage')">
          <router-link :to="{ name: 'campaign', params: { id: 'new' } }" class="btn-new" data-cy="btn-new">
            <PvButton severity="primary" icon="pi pi-plus" :label="$t('globals.buttons.new')" />
          </router-link>
        </div>
      </div>
    </header>

    <PvDataTable :value="campaigns.results" :loading="loading.campaigns" :row-class="highlightedRow"
      :rows="campaigns.perPage" :total-records="campaigns.total" paginator paginator-position="both"
      @page="(e) => onPageChange(e.page + 1)" :first="(queryParams.page - 1) * campaigns.perPage"
      hoverable data-key="id"
      v-model:selection="bulk.checked" selection-mode="multiple"
      @sort="(e) => onSort(e.sortField, e.sortOrder === 1 ? 'asc' : 'desc')"
      lazy>
      <template #header>
        <div class="grid">
          <div class="col-6">
            <form @submit.prevent="getCampaigns">
              <div class="flex gap-2">
                <PvInputText v-model="queryParams.query" name="query" class="flex-1"
                  :placeholder="$t('campaigns.queryPlaceholder')" ref="query" />
                <PvButton type="submit" severity="primary" icon="pi pi-search" />
              </div>
            </form>
          </div>
        </div>

        <div class="actions" v-if="bulk.checked.length > 0">
          <a class="a" href="#" @click.prevent="deleteCampaigns" data-cy="btn-delete-campaigns">
            <i class="pi pi-trash" /> Delete
          </a>
          <span class="a">
            {{ $tc('globals.messages.numSelected', numSelectedCampaigns, { num: numSelectedCampaigns }) }}
            <span v-if="!bulk.all && campaigns.total > campaigns.perPage">
              &mdash;
              <a href="#" @click.prevent="onSelectAll" data-cy="select-all-campaigns">
                {{ $tc('globals.messages.selectAll', campaigns.total, { num: campaigns.total }) }}
              </a>
            </span>
          </span>
        </div>
      </template>

      <PvColumn selection-mode="multiple" header-style="width:3rem" />

      <PvColumn field="status" :header="$t('globals.fields.status')" header-class="cy-status" style="width:10%" sortable>
        <template #body="{ data }">
          <div>
            <p>
              <router-link :to="{ name: 'campaign', params: { id: data.id } }">
                <PvTag :class="data.status">
                  {{ $t(`campaigns.status.${data.status}`) }}
                </PvTag>
                <span class="spinner is-tiny" v-if="isRunning(data.id)">
                  <PvProgressSpinner style="width:1rem;height:1rem" />
                </span>
              </router-link>
            </p>
            <p v-if="isSheduled(data)">
              <span class="is-size-7 has-text-grey scheduled">
                <i class="pi pi-clock" />
                <span v-if="!isDone(data) && !isRunning(data)">
                  {{ $utils.duration(new Date(), data.sendAt, true) }}
                  <br />
                </span>
                {{ $utils.niceDate(data.sendAt, true) }}
              </span>
            </p>
          </div>
        </template>
      </PvColumn>

      <PvColumn field="name" :header="$t('globals.fields.name')" header-class="cy-name" style="width:25%" sortable>
        <template #body="{ data }">
          <div>
            <p>
              <PvTag v-if="data.type === 'optin'" class="is-small">
                {{ $t('lists.optin') }}
              </PvTag>
              <router-link :to="{ name: 'campaign', params: { id: data.id } }">
                {{ data.name }}
                <copy-text :text="data.name" hide-text />
              </router-link>
            </p>
            <p class="is-size-7 has-text-grey">
              <copy-text :text="data.subject" />
            </p>
            <div class="flex flex-wrap gap-1">
              <PvTag class="is-small" v-for="t in data.tags" :key="t" :value="t" />
            </div>
          </div>
        </template>
      </PvColumn>

      <PvColumn field="lists" :header="$t('globals.terms.lists')" style="width:15%">
        <template #body="{ data }">
          <ul>
            <li v-for="l in data.lists" :key="l.id">
              <router-link :to="{ name: 'subscribers_list', params: { listID: l.id } }">
                {{ l.name }}
              </router-link>
            </li>
          </ul>
        </template>
      </PvColumn>

      <PvColumn field="created_at" :header="$t('campaigns.timestamps')" header-class="cy-timestamp" style="width:19%" sortable>
        <template #body="{ data }">
          <div class="fields timestamps" :set="stats = getCampaignStats(data)">
            <p>
              <label for="#">{{ $t('globals.fields.createdAt') }}</label>
              <span>{{ $utils.niceDate(data.createdAt, true) }}</span>
            </p>
            <p v-if="stats.startedAt">
              <label for="#">{{ $t('campaigns.startedAt') }}</label>
              <span>{{ $utils.niceDate(stats.startedAt, true) }}</span>
            </p>
            <p v-if="isDone(data)">
              <label for="#">{{ $t('campaigns.ended') }}</label>
              <span>{{ $utils.niceDate(stats.updatedAt, true) }}</span>
            </p>
            <p v-if="stats.startedAt && stats.updatedAt" class="is-capitalized">
              <label for="#"><i class="pi pi-clock" /></label>
              <span>{{ $utils.duration(stats.startedAt, stats.updatedAt) }}</span>
            </p>
          </div>
        </template>
      </PvColumn>

      <PvColumn field="stats" :header="$t('campaigns.stats')" style="width:15%">
        <template #body="{ data }">
          <div class="fields stats" :set="stats = getCampaignStats(data)">
            <p>
              <label for="#">{{ $t('campaigns.views') }}</label>
              <span>{{ $utils.formatNumber(data.views) }}</span>
            </p>
            <p>
              <label for="#">{{ $t('campaigns.clicks') }}</label>
              <span>{{ $utils.formatNumber(data.clicks) }}</span>
            </p>
            <p>
              <label for="#">{{ $t('campaigns.sent') }}</label>
              <span>
                {{ $utils.formatNumber(stats.sent) }} /
                {{ $utils.formatNumber(stats.toSend) }}
              </span>
            </p>
            <p>
              <label for="#">{{ $t('globals.terms.bounces') }}</label>
              <span>
                <router-link :to="{ name: 'bounces', query: { campaign_id: data.id } }">
                  {{ $utils.formatNumber(data.bounces) }}
                </router-link>
              </span>
            </p>
            <p v-if="stats.rate">
              <label for="#"><i class="pi pi-gauge" /></label>
              <span class="send-rate">
                <span v-tooltip.bottom="`${stats.netRate} / ${$t('campaigns.rateMinuteShort')} @ ${$utils.duration(stats.startedAt, stats.updatedAt)}`">
                  {{ stats.rate.toFixed(0) }} / {{ $t('campaigns.rateMinuteShort') }}
                </span>
              </span>
            </p>
            <p v-if="isRunning(data.id)">
              <label for="#">
                {{ $t('campaigns.progress') }}
                <span class="spinner is-tiny">
                  <PvProgressSpinner style="width:1rem;height:1rem" />
                </span>
              </label>
              <span>
                <PvProgressBar :value="stats.sent / stats.toSend * 100" style="height:6px" />
              </span>
            </p>
          </div>
        </template>
      </PvColumn>

      <PvColumn style="width:15%" align-frozen="right">
        <template #body="{ data }">
          <div>
            <!-- start / pause / resume / scheduled -->
            <template v-if="$can('campaigns:send')">
              <a v-if="canStart(data)" href="#"
                @click.prevent="$utils.confirm(null, () => changeCampaignStatus(data, 'running'))"
                data-cy="btn-start" :aria-label="$t('campaigns.start')">
                <i class="pi pi-send" v-tooltip.bottom="$t('campaigns.start')" />
              </a>

              <a v-if="canPause(data)" href="#"
                @click.prevent="$utils.confirm(null, () => changeCampaignStatus(data, 'paused'))" data-cy="btn-pause"
                :aria-label="$t('campaigns.pause')">
                <i class="pi pi-pause" v-tooltip.bottom="$t('campaigns.pause')" />
              </a>

              <a v-if="canResume(data)" href="#"
                @click.prevent="$utils.confirm(null, () => changeCampaignStatus(data, 'running'))"
                data-cy="btn-resume" :aria-label="$t('campaigns.send')">
                <i class="pi pi-send" v-tooltip.bottom="$t('campaigns.send')" />
              </a>

              <a v-if="canSchedule(data)" href="#"
                @click.prevent="$utils.confirm($t('campaigns.confirmSchedule'), () => changeCampaignStatus(data, 'scheduled'))"
                data-cy="btn-schedule" :aria-label="$t('campaigns.schedule')">
                <i class="pi pi-clock" v-tooltip.bottom="$t('campaigns.schedule')" />
              </a>

              <!-- placeholder for finished campaigns -->
              <a v-if="!canCancel(data) && !canSchedule(data) && !canStart(data)" href="#" data-disabled
                aria-label=" ">
                <i class="pi pi-send" />
              </a>

              <a v-if="canCancel(data)" href="#"
                @click.prevent="$utils.confirm(null, () => changeCampaignStatus(data, 'cancelled'))"
                data-cy="btn-cancel" :aria-label="$t('globals.buttons.cancel')">
                <i class="pi pi-times-circle" v-tooltip.bottom="$t('globals.buttons.cancel')" />
              </a>
              <a v-else href="#" data-disabled aria-label=" ">
                <i class="pi pi-times-circle" />
              </a>
            </template>

            <a href="#" @click.prevent="previewCampaign(data)" data-cy="btn-preview"
              :aria-label="$t('campaigns.preview')">
              <i class="pi pi-eye" v-tooltip.bottom="$t('campaigns.preview')" />
            </a>
            <a v-if="$can('campaigns:manage')" href="#" @click.prevent="$utils.prompt($t('globals.buttons.clone'),
              {
                placeholder: $t('globals.fields.name'),
                value: $t('campaigns.copyOf', { name: data.name }),
              },
              (name) => cloneCampaign(name, data))" data-cy="btn-clone" :aria-label="$t('globals.buttons.clone')">
              <i class="pi pi-copy" v-tooltip.bottom="$t('globals.buttons.clone')" />
            </a>
            <router-link v-if="$can('campaigns:get_analytics')"
              :to="{ name: 'campaignAnalytics', query: { id: data.id } }">
              <i class="pi pi-chart-bar" v-tooltip.bottom="$t('globals.terms.analytics')" />
            </router-link>
            <a v-if="$can('campaigns:manage')" href="#"
              @click.prevent="$utils.confirm($t('campaigns.confirmDelete', { name: data.name }), () => deleteCampaign(data))"
              data-cy="btn-delete" :aria-label="$t('globals.buttons.delete')">
              <i class="pi pi-trash" />
            </a>
          </div>
        </template>
      </PvColumn>

      <template #empty v-if="!loading.campaigns">
        <empty-placeholder />
      </template>
    </PvDataTable>

    <campaign-preview v-if="previewItem" type="campaign" :id="previewItem.id" :title="previewItem.name"
      @close="closePreview" />
  </section>
</template>

<script>
import dayjs from 'dayjs';
import { mapState } from 'pinia';
import { useMainStore } from '../store';
import CampaignPreview from '../components/CampaignPreview.vue';
import CopyText from '../components/CopyText.vue';
import EmptyPlaceholder from '../components/EmptyPlaceholder.vue';

export default {
  components: {
    CampaignPreview,
    EmptyPlaceholder,
    CopyText,
  },

  data() {
    return {
      previewItem: null,
      queryParams: {
        page: 1,
        query: '',
        orderBy: 'created_at',
        order: 'desc',
      },
      pollID: null,
      campaignStatsData: {},

      // Table bulk row selection states.
      bulk: {
        checked: [],
        all: false,
      },
    };
  },

  methods: {
    // Campaign statuses.
    canStart(c) {
      return c.status === 'draft' && !c.sendAt;
    },
    canSchedule(c) {
      return c.status === 'draft' && c.sendAt;
    },
    canPause(c) {
      return c.status === 'running';
    },
    canCancel(c) {
      return c.status === 'running' || c.status === 'paused';
    },
    canResume(c) {
      return c.status === 'paused';
    },
    isSheduled(c) {
      return c.status === 'scheduled' || c.sendAt !== null;
    },
    isDone(c) {
      return c.status === 'finished' || c.status === 'cancelled';
    },

    isRunning(id) {
      if (id in this.campaignStatsData) {
        return true;
      }
      return false;
    },

    highlightedRow(data) {
      if (data.status === 'running') {
        return ['running'];
      }
      return '';
    },

    onPageChange(p) {
      this.queryParams.page = p;
      this.getCampaigns();
    },

    onSort(field, direction) {
      this.queryParams.orderBy = field;
      this.queryParams.order = direction;
      this.getCampaigns();
    },

    // Campaign actions.
    previewCampaign(c) {
      this.previewItem = c;
    },

    closePreview() {
      this.previewItem = null;
    },

    getCampaigns() {
      this.$api.getCampaigns({
        page: this.queryParams.page,
        query: this.queryParams.query.replace(/[^\p{L}\p{N}\s]/gu, ' '),
        order_by: this.queryParams.orderBy,
        order: this.queryParams.order,
        no_body: true,
      });
    },

    // Stats returns the campaign object with stats (sent, toSend etc.)
    // if there's live stats available for running campaigns. Otherwise,
    // it returns the incoming campaign object that has the static stats
    // values.
    getCampaignStats(c) {
      if (c.id in this.campaignStatsData) {
        return this.campaignStatsData[c.id];
      }
      return c;
    },

    pollStats() {
      // Clear any running status polls.
      clearInterval(this.pollID);

      // Poll for the status as long as the import is running.
      this.pollID = setInterval(() => {
        this.$api.getCampaignStats().then((data) => {
          // Stop polling. No running campaigns.
          if (data.length === 0) {
            clearInterval(this.pollID);

            // There were running campaigns and stats earlier. Clear them
            // and refetch the campaigns list with up-to-date fields.
            if (Object.keys(this.campaignStatsData).length > 0) {
              this.getCampaigns();
              this.campaignStatsData = {};
            }
          } else {
            // Turn the list of campaigns [{id: 1, ...}, {id: 2, ...}] into
            // a map indexed by the id: {1: {}, 2: {}}.
            this.campaignStatsData = data.reduce((obj, cur) => ({ ...obj, [cur.id]: cur }), {});
          }
        }, () => {
          clearInterval(this.pollID);
        });
      }, 1000);
    },

    changeCampaignStatus(c, status) {
      this.$api.changeCampaignStatus(c.id, status).then(() => {
        this.$utils.toast(this.$t('campaigns.statusChanged', { name: c.name, status }));
        this.getCampaigns();
        this.pollStats();
      });
    },

    async cloneCampaign(name, c) {
      // Fetch the template body from the server.
      let body = '';
      let bodySource = null;
      await this.$api.getCampaign(c.id).then((data) => {
        body = data.body;
        bodySource = data.bodySource;
      });

      const now = this.$utils.getDate();
      const sendLater = !!c.sendAt;
      let sendAt = null;
      if (sendLater) {
        sendAt = dayjs(c.sendAt).isAfter(now) ? c.sendAt : now.add(7, 'day');
      }

      const data = {
        name,
        subject: c.subject,
        lists: c.lists.map((l) => l.id),
        type: c.type,
        from_email: c.fromEmail,
        content_type: c.contentType,
        messenger: c.messenger,
        tags: c.tags,
        template_id: c.templateId,
        body,
        body_source: bodySource,
        altbody: c.altbody,
        headers: c.headers,
        send_later: sendLater,
        send_at: sendAt,
        archive: c.archive,
        archive_template_id: c.archiveTemplateId,
        archive_meta: c.archiveMeta,
        media: c.media.map((m) => m.id),
      };

      if (c.archive) {
        data.archive_slug = `${name.toLowerCase().replace(/[^a-z0-9]/g, '-')}-${Date.now().toString().slice(-4)}`;
      }

      this.$api.createCampaign(data).then((d) => {
        this.$router.push({ name: 'campaign', params: { id: d.id } });
      });
    },

    deleteCampaign(c) {
      this.$api.deleteCampaign(c.id).then(() => {
        this.getCampaigns();
        this.$utils.toast(this.$t('globals.messages.deleted', { name: c.name }));
      });
    },

    // Mark all campaigns in the query as selected.
    onSelectAll() {
      this.bulk.all = true;
    },

    onTableCheck() {
      // Disable bulk.all selection if there are no rows checked in the table.
      if (this.bulk.checked.length !== this.campaigns.total) {
        this.bulk.all = false;
      }
    },

    deleteCampaigns() {
      const name = this.$tc('globals.terms.campaign', this.numSelectedCampaigns);

      const fn = () => {
        const params = {};
        if (!this.bulk.all && this.bulk.checked.length > 0) {
          // If 'all' is not selected, delete campaigns by IDs.
          params.id = this.bulk.checked.map((c) => c.id);
        } else {
          // 'All' is selected, delete by query.
          params.query = this.queryParams.query.replace(/[^\p{L}\p{N}\s]/gu, ' ');
          params.all = this.bulk.all;
        }

        this.$api.deleteCampaigns(params)
          .then(() => {
            this.getCampaigns();
            this.$utils.toast(this.$tc(
              'globals.messages.deletedCount',
              this.numSelectedCampaigns,
              { num: this.numSelectedCampaigns, name },
            ));
          });
      };

      this.$utils.confirm(this.$tc(
        'globals.messages.confirmDelete',
        this.numSelectedCampaigns,
        { num: this.numSelectedCampaigns, name: name.toLowerCase() },
      ), fn);
    },
  },

  watch: {
    refreshTick() { this.getCampaigns(); },
  },

  computed: {
    ...mapState(useMainStore, ['refreshTick', 'campaigns', 'loading']),

    numSelectedCampaigns() {
      return this.bulk.all ? this.campaigns.total : this.bulk.checked.length;
    },
  },

  mounted() {
    this.getCampaigns();
    this.pollStats();
  },

  unmounted() {
    clearInterval(this.pollID);
  },
};
</script>
