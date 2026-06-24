<template>
  <div class="lists-page">
    <!-- Page header -->
    <div class="page-header">
      <div class="page-header-left">
        <h1 class="page-title">
          {{ $t('globals.terms.lists') }}
          <span v-if="queryParams.status === 'archived'" class="page-title-sub">/ {{ queryParams.status }}</span>
          <span v-if="!isNaN(lists.total)" class="page-title-count">{{ lists.total }}</span>
        </h1>
        <router-link
          v-if="queryParams.status !== 'archived'"
          :to="{ name: 'lists', query: { status: 'archived' } }"
          class="page-sub-link"
        >
{{ $t('globals.buttons.view') }} {{ $t('lists.archived').toLowerCase() }} &rarr;
</router-link>
        <router-link v-else :to="{ name: 'lists' }" class="page-sub-link">
          {{ $t('globals.buttons.view') }} {{ $t('menu.allLists').toLowerCase() }} &rarr;
        </router-link>
      </div>
      <PvButton
        v-if="$can('lists:manage_all')"
        severity="primary"
        icon="pi pi-plus"
        :label="$t('globals.buttons.new')"
        data-cy="btn-new"
        @click="showNewForm"
      />
    </div>

    <!-- Table card -->
    <div class="table-card">
      <PvDataTable
        :value="lists.results"
        :loading="loading.listsFull"
        v-model:selection="bulk.checked"
        selection-mode="checkbox"
        data-key="id"
        :paginator="true"
        paginator-position="bottom"
        :rows="lists.perPage"
        :total-records="lists.total"
        :lazy="true"
        @page="(e) => onPageChange(e.page + 1)"
        sort-field="createdAt"
        :sort-order="1"
        @sort="(e) => onSort(e.sortField, e.sortOrder === 1 ? 'asc' : 'desc')"
        @row-select="onTableCheck"
        @row-unselect="onTableCheck"
        @row-select-all="onTableCheck"
        @row-unselect-all="onTableCheck"
      >
        <template #header>
          <div class="table-toolbar">
            <form class="search-form" @submit.prevent="getLists">
              <PvIconField>
                <PvInputIcon class="pi pi-search" />
                <PvInputText
                  v-model="queryParams.query"
                  name="query"
                  ref="query"
                  data-cy="query"
                  placeholder="Search lists…"
                  class="search-input"
                />
              </PvIconField>
            </form>

            <div v-if="bulk.checked.length > 0" class="bulk-bar">
              <span class="bulk-count">
                {{ $tc('globals.messages.numSelected', numSelectedLists, { num: numSelectedLists }) }}
                <template v-if="!bulk.all && lists.total > lists.perPage">
                  &mdash;
                  <a href="#" @click.prevent="onSelectAll" data-cy="select-all-lists">
                    {{ $tc('globals.messages.selectAll', lists.total, { num: lists.total }) }}
                  </a>
                </template>
              </span>
              <button type="button" class="bulk-btn bulk-btn--danger" @click.prevent="deleteLists" data-cy="btn-delete-lists">
                <i class="pi pi-trash" /> {{ $t('globals.buttons.delete') }}
              </button>
            </div>
          </div>
        </template>

        <PvColumn selection-mode="multiple" header-style="width:3rem" />

        <PvColumn field="name" :header="$t('globals.fields.name')" header-class="cy-name" sortable>
          <template #body="{ data }">
            <a class="row-name" :href="`/lists/${data.id}`" @click.prevent="showEditForm(data)">{{ data.name }}</a>
            <div v-if="data.tags?.length" class="row-tags">
              <PvTag v-for="t in data.tags" :key="t" :value="t" severity="secondary" />
            </div>
          </template>
        </PvColumn>

        <PvColumn field="type" :header="$t('globals.fields.type')" header-class="cy-type" sortable style="width:20%">
          <template #body="{ data }">
            <div class="type-cell">
              <PvTag
                :severity="data.type === 'public' ? 'info' : 'secondary'"
                :data-cy="`type-${data.type}`"
                :value="$t(`lists.types.${data.type}`)"
              />
              <PvTag :severity="data.optin === 'double' ? 'warn' : 'secondary'" :data-cy="`optin-${data.optin}`">
                <i :class="['pi', data.optin === 'double' ? 'pi-user-plus' : 'pi-user-minus']" />
                {{ $t(`lists.optins.${data.optin}`) }}
              </PvTag>
              <a
                v-if="data.optin === 'double'"
                class="optin-send"
                href="#"
                @click.prevent="$utils.confirm(null, () => createOptinCampaign(data))"
                data-cy="btn-send-optin-campaign"
                v-tooltip.bottom="$t('lists.sendOptinCampaign')"
              >
                <i class="pi pi-send" /> {{ $t('lists.sendOptinCampaign') }}
              </a>
            </div>
          </template>
        </PvColumn>

        <PvColumn field="subscriber_count" :header="$t('globals.terms.subscribers')" header-class="cy-subscribers" sortable style="width:12%">
          <template #body="{ data }">
            <router-link v-if="$can('subscribers:get_all', 'subscribers:get')" class="sub-count-link" :to="`/subscribers/lists/${data.id}`">
              <span class="sub-count">{{ $utils.formatNumber(data.subscriberCount) }}</span>
              <span class="sub-view">{{ $t('globals.buttons.view') }}</span>
            </router-link>
            <span v-else class="sub-count">{{ $utils.formatNumber(data.subscriberCount) }}</span>
          </template>
        </PvColumn>

        <PvColumn field="subscriber_counts" style="width:14%">
          <template #body="{ data }">
            <div class="status-breakdown">
              <span v-for="(count, status) in filterStatuses(data)" :key="status" class="status-item">
                <router-link :to="`/subscribers/lists/${data.id}?subscription_status=${status}`" :class="`status-link status-link--${status}`">
                  {{ $utils.formatNumber(count) }}
                </router-link>
                <span class="status-label">{{ $tc(`subscribers.status.${status}`, count) }}</span>
              </span>
            </div>
          </template>
        </PvColumn>

        <PvColumn field="created_at" :header="$t('globals.fields.createdAt')" header-class="cy-created_at" sortable style="width:11%">
          <template #body="{ data }">
            <span class="date-cell">{{ $utils.niceDate(data.createdAt) }}</span>
          </template>
        </PvColumn>

        <PvColumn field="updated_at" :header="$t('globals.fields.updatedAt')" header-class="cy-updated_at" sortable style="width:11%">
          <template #body="{ data }">
            <span class="date-cell">{{ $utils.niceDate(data.updatedAt) }}</span>
          </template>
        </PvColumn>

        <PvColumn style="width:8rem; text-align:right">
          <template #body="{ data }">
            <div class="row-actions">
              <router-link
                v-if="$can('campaigns:manage')"
                :to="`/campaigns/new?list_id=${data.id}`"
                class="row-action-btn"
                data-cy="btn-campaign"
                v-tooltip.bottom="$t('lists.sendCampaign')"
              >
<i class="pi pi-send" />
</router-link>

              <button
                v-if="$can('lists:manage') || $canList(data.id, 'list:manage')"
                type="button"
                class="row-action-btn"
                data-cy="btn-edit"
                v-tooltip.bottom="$t('globals.buttons.edit')"
                @click="showEditForm(data)"
              >
<i class="pi pi-pencil" />
</button>

              <router-link
                v-if="$can('subscribers:import')"
                :to="{ name: 'import', query: { list_id: data.id } }"
                class="row-action-btn"
                data-cy="btn-import"
                v-tooltip.bottom="$t('import.title')"
              >
<i class="pi pi-upload" />
</router-link>

              <button
                v-if="$can('lists:manage') || $canList(data.id, 'list:manage')"
                type="button"
                class="row-action-btn row-action-btn--danger"
                data-cy="btn-delete"
                v-tooltip.bottom="$t('globals.buttons.delete')"
                @click="deleteList(data)"
              >
<i class="pi pi-trash" />
</button>
            </div>
          </template>
        </PvColumn>

        <template #empty v-if="!loading.listsFull">
          <empty-placeholder />
        </template>
      </PvDataTable>
    </div>

    <!-- Add / edit form modal -->
    <PvDialog v-model:visible="isFormVisible" :style="{ width: '580px' }" :closable="true" :show-header="false" modal @hide="onFormClose">
      <list-form :data="curItem" :is-editing="isEditing" @finished="formFinished" @close="isFormVisible = false" />
    </PvDialog>

    <p v-if="settings['app.cache_slow_queries']" class="cache-note">
      *{{ $t('globals.messages.slowQueriesCached') }}
      <a href="https://listmonk.app/docs/maintenance/performance/" target="_blank" rel="noopener noreferrer">
        <i class="pi pi-external-link" /> {{ $t('globals.buttons.learnMore') }}
      </a>
    </p>
  </div>
</template>

<script>
import { mapState } from 'pinia';
import { useMainStore } from '../store';
import EmptyPlaceholder from '../components/EmptyPlaceholder.vue';
import ListForm from './ListForm.vue';

export default {
  components: {
    ListForm,
    EmptyPlaceholder,
  },

  data() {
    return {
      // Current list item being edited.
      curItem: null,
      isEditing: false,
      isFormVisible: false,
      lists: [],
      queryParams: {
        page: 1,
        query: '',
        orderBy: 'id',
        order: 'asc',
        status: this.$route.query.status || 'active',
      },

      // Table bulk row selection states.
      bulk: {
        checked: [],
        all: false,
      },
    };
  },

  methods: {
    onPageChange(p) {
      this.queryParams.page = p;
      this.getLists();
    },

    onSort(field, direction) {
      this.queryParams.orderBy = field;
      this.queryParams.order = direction;
      this.getLists();
    },

    // Show the edit list form.
    showEditForm(list) {
      this.curItem = list;
      this.isFormVisible = true;
      this.isEditing = true;
    },

    // Show the new list form.
    showNewForm() {
      this.curItem = {};
      this.isFormVisible = true;
      this.isEditing = false;
    },

    formFinished() {
      this.getLists();
    },

    onFormClose() {
      if (this.$route.params.id) {
        this.$router.push({ name: 'lists' });
      }
    },

    filterStatuses(list) {
      const out = { ...list.subscriberStatuses };
      if (list.optin === 'single') {
        delete out.unconfirmed;
        delete out.confirmed;
      }
      return out;
    },

    getLists() {
      this.$api.queryLists({
        page: this.queryParams.page,
        query: this.queryParams.query.replace(/[^\p{L}\p{N}\s]/gu, ' '),
        order_by: this.queryParams.orderBy,
        order: this.queryParams.order,
        status: this.queryParams.status,
      }).then((resp) => {
        this.lists = resp;
      });

      // Also fetch the minimal lists for the global store that appears
      // in dropdown menus on other pages like import and campaigns.
      this.$api.getLists({ minimal: true, per_page: 'all', status: 'active' });
    },

    deleteList(list) {
      this.$utils.confirm(
        this.$t('lists.confirmDelete'),
        () => {
          this.$api.deleteList(list.id).then(() => {
            this.getLists();

            this.$utils.toast(this.$t('globals.messages.deleted', { name: list.name }));
          });
        },
      );
    },

    // Mark all lists in the query as selected.
    onSelectAll() {
      this.bulk.all = true;
    },

    onTableCheck() {
      // Disable bulk.all selection if there are no rows checked in the table.
      if (this.bulk.checked.length !== this.lists.total) {
        this.bulk.all = false;
      }
    },

    deleteLists() {
      const name = this.$tc('globals.terms.list', this.numSelectedCampaigns);

      const fn = () => {
        const params = {};
        if (!this.bulk.all && this.bulk.checked.length > 0) {
          // If 'all' is not selected, delete lists by IDs.
          params.id = this.bulk.checked.map((l) => l.id);
        } else {
          // 'All' is selected, delete by query.
          params.query = this.queryParams.query.replace(/[^\p{L}\p{N}\s]/gu, ' ');
          params.all = this.bulk.all;
        }

        this.$api.deleteLists(params)
          .then(() => {
            this.getLists();
            this.$utils.toast(this.$tc(
              'globals.messages.deletedCount',
              this.numSelectedLists,
              { num: this.numSelectedLists, name },
            ));
          });
      };

      this.$utils.confirm(this.$tc(
        'globals.messages.confirmDelete',
        this.numSelectedLists,
        { num: this.numSelectedLists, name: name.toLowerCase() },
      ), fn);
    },

    createOptinCampaign(list) {
      const data = {
        name: this.$t('lists.optinTo', { name: list.name }),
        subject: this.$t('lists.confirmSub', { name: list.name }),
        lists: [list.id],
        from_email: this.settings['app.from_email'],
        content_type: 'richtext',
        messenger: 'email',
        type: 'optin',
      };

      this.$api.createCampaign(data).then((d) => {
        this.$router.push({ name: 'campaign', hash: '#content', params: { id: d.id } });
      });
      return false;
    },
  },

  watch: {
    refreshTick() { this.getLists(); },
  },

  computed: {
    ...mapState(useMainStore, ['refreshTick', 'loading', 'settings']),

    numSelectedLists() {
      return this.bulk.all ? this.lists.total : this.bulk.checked.length;
    },
  },

  mounted() {
    if (this.$route.params.id) {
      this.$api.getList(parseInt(this.$route.params.id, 10)).then((data) => {
        this.showEditForm(data);
      });
    } else {
      this.getLists();
    }
  },
};
</script>

<style scoped lang="scss">
.lists-page {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

// Page header

.page-header-left { display: flex; flex-direction: column; gap: 0.25rem; }

.page-title-sub { color: var(--lm-text-subtle); font-weight: 400; }

.page-sub-link { font-size: 0.8rem; color: var(--lm-text-muted); text-decoration: none; &:hover { color: var(--lm-primary); } }

// Table card

// Toolbar
.table-toolbar {
  display: flex;
  align-items: center;
  gap: 1rem;
  flex-wrap: wrap;
}
.search-form { flex: 1; min-width: 220px; max-width: 360px; }
:deep(.search-input) { width: 100%; }

.bulk-bar {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-left: auto;
}
.bulk-count { font-size: 0.85rem; color: var(--lm-text-muted); a { color: var(--lm-primary); } }
.bulk-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.4rem;
  padding: 0.35rem 0.75rem;
  border-radius: 6px;
  font-size: 0.8rem;
  font-weight: 500;
  border: 1px solid;
  cursor: pointer;
  background: var(--lm-surface);
  &--danger { color: var(--lm-danger); border-color: var(--lm-danger-border); &:hover { background: var(--lm-danger-bg); } }
}

// Row cells
.row-name {
  font-weight: 500;
  color: var(--lm-primary);
  text-decoration: none;
}
.row-tags { display: flex; flex-wrap: wrap; gap: 0.3rem; margin-top: 0.3rem; }

.type-cell { display: flex; flex-wrap: wrap; align-items: center; gap: 0.4rem; }
.optin-send {
  font-size: 0.75rem;
  color: var(--lm-text-muted);
  text-decoration: none;
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
}

.sub-count-link { text-decoration: none; display: flex; flex-direction: column; }
.sub-count { font-weight: 600; color: var(--lm-text); font-size: 0.95rem; }
.sub-view { font-size: 0.72rem; color: var(--lm-text-subtle); }

.status-breakdown { display: flex; flex-direction: column; gap: 0.15rem; }
.status-item { display: flex; align-items: center; gap: 0.35rem; font-size: 0.78rem; }
.status-link { font-weight: 600; color: var(--lm-text); text-decoration: none; &:hover { color: var(--lm-primary); } }
.status-label { color: var(--lm-text-subtle); }

.date-cell { font-size: 0.82rem; color: var(--lm-text-muted); }

.cache-note {
  font-size: 0.78rem;
  color: var(--lm-text-subtle);
  a { color: var(--lm-text-subtle); text-decoration: underline; }
}
</style>
