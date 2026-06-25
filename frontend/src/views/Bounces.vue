<template>
  <div class="bounces-page">
    <div class="page-header">
      <h1 class="page-title">
        {{ $t('globals.terms.bounces') }}
        <span v-if="bounces.total > 0" class="page-title-count">{{ bounces.total }}</span>
      </h1>
    </div>

    <div class="table-card">
      <PvDataTable :value="bounces.results" :loading="loading.bounces"
        sort-field="createdAt" :sort-order="-1"
        selection-mode="checkbox" v-model:selection="bulk.checked"
        @update:selection="onTableCheck"
        :rows="bounces.perPage" :paginator="true" paginator-position="bottom" :total-records="bounces.total"
        :lazy="true" @page="(e) => onPageChange(e.page + 1)"
        @sort="(e) => onSort(e.sortField, e.sortOrder === 1 ? 'asc' : 'desc')"
        data-key="id"
        :expanded-rows="expandedRows" @update:expanded-rows="expandedRows = $event">
        <template #header>
          <div v-if="bulk.checked.length > 0" class="bulk-bar">
            <span class="bulk-count">
              {{ $t('globals.messages.numSelected', { num: numSelectedBounces }) }}
              <template v-if="!bulk.all && bounces.total > bounces.perPage">
                &mdash; <a href="#" @click.prevent="selectAllBounces">{{ $t('subscribers.selectAll', { num: bounces.total }) }}</a>
              </template>
            </span>
            <button type="button" class="bulk-btn bulk-btn--danger" data-cy="btn-delete"
              @click="$utils.confirm(null, () => deleteBounces())">
              <i class="pi pi-trash" /> {{ $t('globals.buttons.delete') }}
            </button>
            <button type="button" class="bulk-btn bulk-btn--warn" data-cy="btn-manage-blocklist"
              @click="$utils.confirm(null, () => blocklistSubscribers())">
              <i class="pi pi-ban" /> {{ $t('import.blocklist') }}
            </button>
          </div>
        </template>

        <PvColumn selection-mode="multiple" header-style="width:3rem" />
        <PvColumn expander header-style="width:3rem" />

        <PvColumn field="email" :header="$t('subscribers.email')" sortable>
          <template #body="{ data }">
            <router-link class="row-name" :class="{ 'row-name--blocked': data.subscriberStatus === 'blocklisted' }"
              :to="{ name: 'subscriber', params: { id: data.subscriberId } }">
              {{ data.email }}
            </router-link>
            <PvTag v-if="data.subscriberStatus !== 'enabled'" severity="danger" size="small"
              data-cy="blocklisted" :value="$t(`subscribers.status.${data.subscriberStatus}`)" />
          </template>
        </PvColumn>

        <PvColumn field="campaign" :header="$tc('globals.terms.campaign')" sortable>
          <template #body="{ data }">
            <router-link v-if="data.campaign" class="row-link"
              :to="{ name: 'bounces', query: { campaign_id: data.campaign.id } }">
              {{ data.campaign.name }}
            </router-link>
            <span v-else class="text-muted">-</span>
          </template>
        </PvColumn>

        <PvColumn field="source" :header="$t('bounces.source')" sortable>
          <template #body="{ data }">
            <router-link class="row-link" :to="{ name: 'bounces', query: { source: data.source } }">
              {{ data.source }}
            </router-link>
          </template>
        </PvColumn>

        <PvColumn field="type" :header="$t('globals.fields.type')" sortable style="width:8rem">
          <template #body="{ data }">
            <PvTag :severity="data.type === 'hard' ? 'danger' : 'warn'" size="small"
              :value="$t(`bounces.${data.type}`)" />
          </template>
        </PvColumn>

        <PvColumn field="created_at" :header="$t('globals.fields.createdAt')" sortable style="width:11rem">
          <template #body="{ data }">{{ $utils.niceDate(data.createdAt, true) }}</template>
        </PvColumn>

        <PvColumn style="width:4rem; text-align:right" align-frozen="right">
          <template #body="{ data }">
            <div class="row-actions">
              <button type="button" class="row-action-btn row-action-btn--danger"
                data-cy="btn-delete" v-tooltip.bottom="$t('globals.buttons.delete')"
                @click="$utils.confirm(null, () => deleteBounce(data))">
                <i class="pi pi-trash" />
              </button>
            </div>
          </template>
        </PvColumn>

        <template #expansion="{ data }">
          <pre class="meta-expansion">{{ data.meta }}</pre>
        </template>

        <template #empty v-if="!loading.bounces">
          <empty-placeholder />
        </template>
      </PvDataTable>
    </div>
  </div>
</template>

<script>
import { mapState } from 'pinia';
import { useMainStore } from '../store';
import EmptyPlaceholder from '../components/EmptyPlaceholder.vue';

export default {
  components: {
    EmptyPlaceholder,
  },

  data() {
    return {
      bounces: {},

      expandedRows: [],

      // Table bulk row selection states.
      bulk: {
        checked: [],
        all: false,
      },

      // Query params to filter the getSubscribers() API call.
      queryParams: {
        page: 1,
        orderBy: 'created_at',
        order: 'desc',
        campaignID: 0,
        source: '',
      },
    };
  },

  methods: {
    onSort(field, direction) {
      this.queryParams.orderBy = field;
      this.queryParams.order = direction;
      this.getBounces();
    },

    onPageChange(p) {
      this.queryParams.page = p;
      this.getBounces();
    },
    // Mark all bounces in the query as selected.
    selectAllBounces() {
      this.bulk.all = true;
    },
    onTableCheck() {
      // Disable bulk.all selection if there are no rows checked in the table.
      if (this.bulk.checked.length !== this.bounces.total) {
        this.bulk.all = false;
      }
    },

    getBounces() {
      this.bulk.checked = [];
      this.bulk.all = false;

      this.$api.getBounces({
        page: this.queryParams.page,
        order_by: this.queryParams.orderBy,
        order: this.queryParams.order,
        campaign_id: this.queryParams.campaign_id,
        source: this.queryParams.source,
      }).then((data) => {
        this.bounces = data;
      });
    },

    deleteBounce(b) {
      this.$api.deleteBounce(b.id).then(() => {
        this.getBounces();
        this.$utils.toast(this.$t('globals.messages.deleted', { name: b.email }));
      });
    },

    deleteBounces() {
      const params = {};
      if (!this.bulk.all && this.bulk.checked.length > 0) {
        params.id = this.bulk.checked.map((s) => s.id);
      } else if (this.bulk.all) {
        params.all = true;
      }

      this.$api.deleteBounces(params).then(() => {
        this.getBounces();
        this.$utils.toast(this.$t(
          'globals.messages.deletedCount',
          { name: this.$tc('globals.terms.bounces'), num: this.numSelectedBounces },
        ));
      });
    },

    blocklistSubscribers() {
      const cb = () => {
        this.getBounces();
        this.$utils.toast(this.$t('globals.messages.done'));
      };

      if (!this.bulk.all && this.bulk.checked.length > 0) {
        const subIds = this.bulk.checked.map((s) => s.subscriberId);
        this.$api.blocklistSubscribers({ ids: subIds }).then(cb);
        return;
      }

      this.$api.blocklistBouncedSubscribers({ all: true }).then(cb);
    },
  },

  watch: {
    refreshTick() { this.getBounces(); },
  },

  computed: {
    ...mapState(useMainStore, ['refreshTick', 'templates', 'loading']),
    numSelectedBounces() {
      if (this.bulk.all) {
        return this.bounces.total;
      }
      return this.bulk.checked.length;
    },
  },

  mounted() {
    if (this.$route.query.campaign_id) {
      this.queryParams.campaign_id = parseInt(this.$route.query.campaign_id, 10);
    }

    if (this.$route.query.source) {
      this.queryParams.source = this.$route.query.source;
    }

    this.getBounces();
  },
};
</script>

<style scoped lang="scss">
.bounces-page { display: flex; flex-direction: column; gap: 1.5rem; }

.bulk-bar { display: flex; align-items: center; gap: 0.75rem; flex-wrap: wrap; }
.bulk-count { font-size: 0.85rem; color: var(--lm-text-muted); a { color: var(--lm-primary); text-decoration: none; } }
.bulk-btn {
  display: inline-flex; align-items: center; gap: 0.35rem;
  padding: 0.35rem 0.75rem; border: 1px solid var(--lm-border); border-radius: 6px;
  background: var(--lm-surface); font-size: 0.8rem; cursor: pointer; color: var(--lm-text-muted);
  &--danger { color: var(--lm-danger); border-color: var(--lm-danger-border); &:hover { background: var(--lm-danger-bg); } }
  &--warn   { color: #d97706; border-color: var(--lm-warn-border); &:hover { background: var(--lm-warn-bg); } }
}

.row-name { color: var(--lm-text); font-weight: 500; text-decoration: none; &:hover { color: var(--lm-primary); } &--blocked { color: var(--lm-text-subtle); text-decoration: line-through; } }
.row-link  { color: var(--lm-primary); text-decoration: none; &:hover { text-decoration: underline; } }
.text-muted { color: var(--lm-text-subtle); }

.meta-expansion { font-size: 0.78rem; padding: 1rem; background: var(--lm-bg); border-radius: 6px; margin: 0.5rem 1rem; }
</style>
