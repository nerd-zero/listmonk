<template>
  <section class="bounces">
    <header class="page-header grid">
      <div class="col-8">
        <h1 class="title is-4">
          {{ $t('globals.terms.bounces') }}
          <span v-if="bounces.total > 0">({{ bounces.total }})</span>
        </h1>
      </div>
    </header>

    <PvDataTable :value="bounces.results" :hoverable="true" :loading="loading.bounces"
      sort-field="createdAt" :sort-order="-1"
      selection-mode="checkbox" v-model:selection="bulk.checked"
      @update:selection="onTableCheck"
      :rows="bounces.perPage" :paginator="true" :total-records="bounces.total"
      :lazy="true" @page="(e) => onPageChange(e.page + 1)"
      @sort="(e) => onSort(e.sortField, e.sortOrder === 1 ? 'asc' : 'desc')"
      data-key="id"
      :expanded-rows="expandedRows" @update:expanded-rows="expandedRows = $event">
      <template #header>
        <div class="actions">
          <template v-if="bulk.checked.length > 0">
            <a class="a" href="#" @click.prevent="$utils.confirm(null, () => deleteBounces())" data-cy="btn-delete">
              <i class="pi pi-trash" /> {{ $t('globals.buttons.delete') }}
            </a>
            <a class="a" href="#" @click.prevent="$utils.confirm(null, () => blocklistSubscribers())"
              data-cy="btn-manage-blocklist">
              <i class="pi pi-ban" /> {{ $t('import.blocklist') }}
            </a>
            <span>
              {{ $t('globals.messages.numSelected', { num: numSelectedBounces }) }}
              <span v-if="!bulk.all && bounces.total > bounces.perPage">
                &mdash;
                <a href="#" @click.prevent="selectAllBounces">
                  {{ $t('subscribers.selectAll', { num: bounces.total }) }}
                </a>
              </span>
            </span>
          </template>
        </div>
      </template>

      <PvColumn selection-mode="multiple" header-style="width:3rem" />

      <PvColumn expander header-style="width:3rem" />

      <PvColumn field="email" :header="$t('subscribers.email')" sortable>
        <template #body="{ data }">
          <router-link :to="{ name: 'subscriber', params: { id: data.subscriberId } }"
            :class="{ 'blocklisted': data.subscriberStatus === 'blocklisted' }">
            {{ data.email }}
            <PvTag v-if="data.subscriberStatus !== 'enabled'" :class="data.subscriberStatus"
              data-cy="blocklisted" :value="$t(`subscribers.status.${data.subscriberStatus}`)" />
          </router-link>
        </template>
      </PvColumn>

      <PvColumn field="campaign" :header="$tc('globals.terms.campaign')" sortable>
        <template #body="{ data }">
          <router-link v-if="data.campaign" :to="{ name: 'bounces', query: { campaign_id: data.campaign.id } }">
            {{ data.campaign.name }}
          </router-link>
          <span v-else>-</span>
        </template>
      </PvColumn>

      <PvColumn field="source" :header="$t('bounces.source')" sortable>
        <template #body="{ data }">
          <router-link :to="{ name: 'bounces', query: { source: data.source } }">
            {{ data.source }}
          </router-link>
        </template>
      </PvColumn>

      <PvColumn field="type" :header="$t('globals.fields.type')" sortable>
        <template #body="{ data }">
          <router-link :to="{ name: 'bounces', query: { type: data.type } }">
            {{ $t(`bounces.${data.type}`) }}
          </router-link>
        </template>
      </PvColumn>

      <PvColumn field="created_at" :header="$t('globals.fields.createdAt')" sortable>
        <template #body="{ data }">
          {{ $utils.niceDate(data.createdAt, true) }}
        </template>
      </PvColumn>

      <PvColumn body-class="actions" align-frozen="right">
        <template #body="{ data }">
          <div>
            <a v-if="!data.isDefault" href="#" @click.prevent="$utils.confirm(null, () => deleteBounce(data))"
              data-cy="btn-delete" :aria-label="$t('globals.buttons.delete')">
              <i class="pi pi-trash" v-tooltip.bottom="$t('globals.buttons.delete')" />
            </a>
            <span v-else class="a has-text-grey-light">
              <i class="pi pi-trash" />
            </span>
          </div>
        </template>
      </PvColumn>

      <template #expansion="{ data }">
        <pre class="is-size-7">{{ data.meta }}</pre>
      </template>

      <template #empty v-if="!loading.templates">
        <empty-placeholder />
      </template>
    </PvDataTable>
  </section>
</template>

<script>
import { mapState } from 'vuex';
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

  computed: {
    ...mapState(['templates', 'loading']),
    numSelectedBounces() {
      if (this.bulk.all) {
        return this.bounces.total;
      }
      return this.bulk.checked.length;
    },
  },

  created() {
    this.$root.$on('page.refresh', this.getBounces);
  },

  unmounted() {
    this.$root.$off('page.refresh', this.getBounces);
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
