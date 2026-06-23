<template>
  <div class="subs-page">
    <!-- Page header -->
    <div class="page-header">
      <div class="page-header-left">
        <h1 class="page-title">
          {{ $t('globals.terms.subscribers') }}
          <span v-if="!isNaN(subscribers.total)" class="page-title-count">{{ subscribers.total }}</span>
          <template v-if="currentList">
            <span class="page-title-sub">/ {{ currentList.name }}</span>
            <span v-if="queryParams.subStatus" class="page-title-sub is-capitalized">({{ queryParams.subStatus }})</span>
          </template>
        </h1>
      </div>
      <PvButton
        v-if="$can('subscribers:manage')"
        severity="primary"
        icon="pi pi-plus"
        :label="$t('globals.buttons.new')"
        data-cy="btn-new"
        @click="showNewForm"
      />
    </div>

    <!-- Table card -->
    <div class="table-card">
      <!-- Search toolbar (outside table header for cleaner layout) -->
      <div class="search-toolbar">
        <form class="search-form" @submit.prevent="onSubmit">
          <PvIconField class="search-field">
            <PvInputIcon class="pi pi-search" />
            <PvInputText
              @input="onSimpleQueryInput"
              v-model="queryInput"
              :placeholder="$t('subscribers.queryPlaceholder')"
              ref="query"
              :disabled="isSearchAdvanced"
              data-cy="search"
              class="search-input"
            />
          </PvIconField>
        </form>

        <a href="#" class="advanced-toggle" @click.prevent="toggleAdvancedSearch" data-cy="btn-advanced-search">
          <i :class="['pi', isSearchAdvanced ? 'pi-times' : 'pi-sliders-h']" />
          {{ isSearchAdvanced ? $t('subscribers.reset') : $t('subscribers.advancedQuery') }}
        </a>

        <button type="button" class="toolbar-btn" @click.prevent="exportSubscribers" data-cy="btn-export-subscribers">
          <i class="pi pi-download" /> {{ $t('subscribers.export') }}
        </button>
      </div>

      <!-- Advanced query panel -->
      <div v-if="isSearchAdvanced" class="advanced-panel">
        <form @submit.prevent="onSubmit">
          <PvTextarea
            v-model="queryParams.queryExp"
            @keydown="onAdvancedQueryEnter"
            rows="3"
            placeholder="subscribers.name LIKE '%user%' or subscribers.status='blocklisted'"
            class="w-full"
            data-cy="query"
          />
          <div class="advanced-footer">
            <span class="advanced-help">
              {{ $t('subscribers.advancedQueryHelp') }}.
              <a href="https://listmonk.app/docs/querying-and-segmentation" target="_blank" rel="noopener noreferrer">
                {{ $t('globals.buttons.learnMore') }}
              </a>
            </span>
            <PvButton type="submit" severity="primary" size="small" icon="pi pi-search" :label="$t('subscribers.query')" data-cy="btn-query" />
          </div>
        </form>
      </div>

      <PvDataTable
        :value="subscribers.results ?? []"
        :loading="loading.subscribers"
        data-key="id"
        :rows="subscribers.perPage"
        :paginator="true"
        paginator-position="bottom"
        :total-records="subscribers.total"
        :lazy="true"
        @page="(e) => onPageChange(e.page + 1)"
        @sort="(e) => onSort(e.sortField, e.sortOrder === 1 ? 'asc' : 'desc')"
        selection-mode="checkbox"
        v-model:selection="bulk.checked"
        @row-select="onTableCheck"
        @row-unselect="onTableCheck"
        @row-select-all="onTableCheck"
        @row-unselect-all="onTableCheck"
      >
        <template #header>
          <div v-if="bulk.checked.length > 0" class="bulk-bar">
            <span class="bulk-count">
              {{ $t('globals.messages.numSelected', { num: numSelectedSubscribers }) }}
              <template v-if="!bulk.all && subscribers.total > subscribers.perPage">
                &mdash;
                <a href="#" @click.prevent="selectAllSubscribers">
                  {{ $t('globals.messages.selectAll', { num: subscribers.total }) }}
                </a>
              </template>
            </span>
            <button type="button" class="bulk-btn" @click.prevent="showBulkListForm" data-cy="btn-manage-lists">
              <i class="pi pi-list" /> Manage lists
            </button>
            <button type="button" class="bulk-btn bulk-btn--danger" @click.prevent="deleteSubscribers" data-cy="btn-delete-subscribers">
              <i class="pi pi-trash" /> {{ $t('globals.buttons.delete') }}
            </button>
            <button type="button" class="bulk-btn bulk-btn--warn" @click.prevent="blocklistSubscribers" data-cy="btn-manage-blocklist">
              <i class="pi pi-user-minus" /> Blocklist
            </button>
          </div>
        </template>

        <PvColumn selection-mode="multiple" header-style="width:3rem" />

        <PvColumn field="email" :header="$t('subscribers.email')" header-class="cy-email" sortable>
          <template #body="{ data }">
            <div class="email-cell">
              <a class="row-name" :class="{ 'row-name--blocked': data.status === 'blocklisted' }"
                :href="`/subscribers/${data.id}`" @click.prevent="showEditForm(data)">
                {{ data.email }}
                <copy-text :text="`${data.email}`" hide-text />
              </a>
              <PvTag v-if="data.status !== 'enabled'" severity="danger" size="small" data-cy="blocklisted"
                :value="$t(`subscribers.status.${data.status}`)" />
            </div>
            <div v-if="data.lists?.length" class="list-tags">
              <router-link v-for="l in data.lists" :key="l.id" :to="`/subscribers/lists/${l.id}`">
                <PvTag severity="secondary" size="small">
                  {{ l.name }}<sup v-if="l.optin === 'double' || l.subscriptionStatus === 'unsubscribed'"> {{ $t(`subscribers.status.${l.subscriptionStatus}`) }}</sup>
                </PvTag>
              </router-link>
            </div>
          </template>
        </PvColumn>

        <PvColumn field="name" :header="$t('globals.fields.name')" header-class="cy-name" sortable style="width:18%">
          <template #body="{ data }">
            <a class="row-name" :class="{ 'row-name--blocked': data.status === 'blocklisted' }"
              :href="`/subscribers/${data.id}`" @click.prevent="showEditForm(data)">
              {{ data.name }}
            </a>
          </template>
        </PvColumn>

        <PvColumn field="lists" :header="$t('globals.terms.lists')" header-class="cy-lists" style="width:6rem; text-align:center">
          <template #body="{ data }">
            <span class="list-count">{{ listCount(data.lists) }}</span>
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

        <PvColumn style="width:7rem; text-align:right">
          <template #body="{ data }">
            <div class="row-actions">
              <a :href="`/api/subscribers/${data.id}/export`" class="row-action-btn"
                data-cy="btn-download" v-tooltip.bottom="$t('subscribers.downloadData')">
                <i class="pi pi-download" />
              </a>
              <button v-if="$can('subscribers:manage')" type="button" class="row-action-btn"
                data-cy="btn-edit" v-tooltip.bottom="$t('globals.buttons.edit')" @click="showEditForm(data)">
                <i class="pi pi-pencil" />
              </button>
              <button v-if="$can('subscribers:manage')" type="button" class="row-action-btn row-action-btn--danger"
                data-cy="btn-delete" v-tooltip.bottom="$t('globals.buttons.delete')" @click="deleteSubscriber(data)">
                <i class="pi pi-trash" />
              </button>
            </div>
          </template>
        </PvColumn>

        <template #empty v-if="!loading.subscribers">
          <empty-placeholder />
        </template>
      </PvDataTable>
    </div>

    <!-- Manage list modal -->
    <PvDialog v-model:visible="isBulkListFormVisible" :style="{ width: '500px' }" show-header="false" :closable="false" modal>
      <subscriber-bulk-list :num-subscribers="numSelectedSubscribers" @finished="bulkChangeLists" @close="isBulkListFormVisible = false" />
    </PvDialog>

    <!-- Add / edit form modal -->
    <PvDialog v-model:visible="isFormVisible" :style="{ width: '850px' }" show-header="false" :closable="false" modal @hide="onFormClose">
      <subscriber-form :data="curItem" :is-editing="isEditing" @finished="querySubscribers" @close="isFormVisible = false" />
    </PvDialog>
  </div>
</template>

<script>
import { mapState } from 'pinia';
import { useMainStore } from '../store';
import EmptyPlaceholder from '../components/EmptyPlaceholder.vue';
import { uris } from '../constants';
import SubscriberBulkList from './SubscriberBulkList.vue';
import SubscriberForm from './SubscriberForm.vue';
import CopyText from '../components/CopyText.vue';

export default {
  components: {
    SubscriberForm,
    SubscriberBulkList,
    CopyText,
    EmptyPlaceholder,
  },

  data() {
    return {
      // Current subscriber item being edited.
      curItem: null,
      isSearchAdvanced: false,
      isEditing: false,
      isFormVisible: false,
      isBulkListFormVisible: false,

      // Table bulk row selection states.
      bulk: {
        checked: [],
        all: false,
      },

      queryInput: '',

      // Query params to filter the getSubscribers() API call.
      queryParams: {
        // Search query expression.
        queryExp: '',
        search: '',

        // ID of the list the current subscriber view is filtered by.
        listID: null,
        page: 1,
        orderBy: 'id',
        order: 'desc',
        subStatus: null,
      },
    };
  },

  methods: {
    // Count the lists from which a subscriber has not unsubscribed.
    listCount(lists) {
      return lists.reduce((defVal, item) => (defVal + (item.subscriptionStatus !== 'unsubscribed' ? 1 : 0)), 0);
    },

    toggleAdvancedSearch() {
      this.isSearchAdvanced = !this.isSearchAdvanced;
      this.queryParams.search = '';

      // Toggling to simple search.
      if (!this.isSearchAdvanced) {
        this.queryInput = '';
        this.queryParams.queryExp = '';
        this.queryParams.page = 1;
        this.querySubscribers();
        this.$refs.query.$el.focus();
        return;
      }

      // Toggling to advanced search.
      const q = this.queryInput.replace(/'/, "''").trim();
      if (q) {
        if (this.$utils.validateEmail(q)) {
          this.queryParams.queryExp = `email = '${q.toLowerCase()}'`;
        } else {
          this.queryParams.queryExp = `(name ~* '${q}' OR email ~* '${q.toLowerCase()}')`;
        }
      }

      // Toggling to advanced search.
      this.$nextTick(() => {
        this.$refs.queryExp.$el.focus();
      });
    },

    // Mark all subscribers in the query as selected.
    selectAllSubscribers() {
      this.bulk.all = true;
    },

    onTableCheck() {
      // Disable bulk.all selection if there are no rows checked in the table.
      if (this.bulk.checked.length !== this.subscribers.total) {
        this.bulk.all = false;
      }
    },

    // Show the edit list form.
    showEditForm(sub) {
      this.curItem = sub;
      this.isFormVisible = true;
      this.isEditing = true;
    },

    // Show the new list form.
    showNewForm() {
      this.curItem = {};
      this.isFormVisible = true;
      this.isEditing = false;
    },

    showBulkListForm() {
      this.isBulkListFormVisible = true;
    },

    onFormClose() {
      if (this.$route.params.id) {
        this.$router.push({ name: 'subscribers' });
      }
    },

    onPageChange(p) {
      this.querySubscribers({ page: p });
    },

    onSort(field, direction) {
      this.querySubscribers({ orderBy: field, order: direction });
    },

    // Prepares an SQL expression for simple name search inputs and saves it
    // in this.queryExp.
    onSimpleQueryInput(v) {
      const q = v.replace(/'/, "''").trim();
      this.queryParams.queryExp = '';
      this.queryParams.page = 1;
      this.queryParams.search = q.toLowerCase();
    },

    // Ctrl + Enter on the advanced query searches.
    onAdvancedQueryEnter(e) {
      if (e.ctrlKey && e.key === 'Enter') {
        this.onSubmit();
      }
    },

    onSubmit() {
      this.querySubscribers({ page: 1 });
    },

    // Search / query subscribers.
    querySubscribers(params) {
      this.queryParams = { ...this.queryParams, ...params };

      const qp = {
        list_id: this.queryParams.listID,
        search: this.queryParams.search,
        query: this.queryParams.queryExp,
        page: this.queryParams.page,
        subscription_status: this.queryParams.subStatus,
        order_by: this.queryParams.orderBy,
        order: this.queryParams.order,
      };

      if (this.queryParams.queryExp) {
        delete qp.search;
      } else {
        delete qp.queryExp;
      }

      this.$nextTick(() => {
        this.$api.getSubscribers(qp).then(() => {
          this.bulk.checked = [];
        });
      });
    },

    deleteSubscriber(sub) {
      this.$utils.confirm(
        null,
        () => {
          this.$api.deleteSubscriber(sub.id).then(() => {
            this.querySubscribers();

            this.$utils.toast(this.$t('globals.messages.deleted', { name: sub.name }));
          });
        },
      );
    },

    blocklistSubscribers() {
      let fn = null;
      if (!this.bulk.all && this.bulk.checked.length > 0) {
        // If 'all' is not selected, blocklist subscribers by IDs.
        fn = () => {
          const ids = this.bulk.checked.map((s) => s.id);
          this.$api.blocklistSubscribers({ ids })
            .then(() => this.querySubscribers());
        };
      } else {
        // 'All' is selected, blocklist by query.
        fn = () => {
          this.$api.blocklistSubscribersByQuery({
            search: this.queryParams.search,
            query: this.queryParams.queryExp,
            list_ids: this.queryParams.listID ? [this.queryParams.listID] : null,
            subscription_status: this.queryParams.subStatus,
          }).then(() => this.querySubscribers());
        };
      }

      this.$utils.confirm(this.$t('subscribers.confirmBlocklist', { num: this.numSelectedSubscribers }), fn);
    },

    exportSubscribers() {
      const num = !this.bulk.all && this.bulk.checked.length > 0
        ? this.bulk.checked.length : this.subscribers.total;

      this.$utils.confirm(this.$t('subscribers.confirmExport', { num }), () => {
        const q = new URLSearchParams();

        if (this.queryParams.search) {
          q.append('search', this.queryParams.search);
        } else if (this.queryParams.queryExp) {
          q.append('query', this.queryParams.queryExp);
        }

        if (this.queryParams.listID) {
          q.append('list_id', this.queryParams.listID);
        }

        if (this.queryParams.subStatus) {
          q.append('subscription_status', this.queryParams.subStatus);
        }

        // Export selected subscribers.
        if (!this.bulk.all && this.bulk.checked.length > 0) {
          this.bulk.checked.map((s) => q.append('id', s.id));
        }

        document.location.href = `${uris.exportSubscribers}?${q.toString()}`;
      });
    },

    deleteSubscribers() {
      let fn = null;
      if (!this.bulk.all && this.bulk.checked.length > 0) {
        // If 'all' is not selected, delete subscribers by IDs.
        fn = () => {
          const ids = this.bulk.checked.map((s) => s.id);
          this.$api.deleteSubscribers({ id: ids })
            .then(() => {
              this.querySubscribers();

              this.$utils.toast(this.$t('subscribers.subscribersDeleted', { num: this.numSelectedSubscribers }));
            });
        };
      } else {
        // 'All' is selected, delete by query.
        fn = () => {
          this.$api.deleteSubscribersByQuery({
            // If the query expression is empty, explicitly pass `all=true`
            // so that the backend deletes all records in the DB with an empty query string.
            all: this.queryParams.queryExp.trim() === '' && this.queryParams.search.trim() === '',
            search: this.queryParams.search,
            query: this.queryParams.queryExp,
            list_ids: this.queryParams.listID ? [this.queryParams.listID] : null,
            subscription_status: this.queryParams.subStatus,
          }).then(() => {
            this.querySubscribers();

            this.$utils.toast(this.$t(
              'subscribers.subscribersDeleted',
              { num: this.numSelectedSubscribers },
            ));
          });
        };
      }

      this.$utils.confirm(this.$t('subscribers.confirmDelete', { num: this.numSelectedSubscribers }), fn);
    },

    bulkChangeLists(action, preconfirm, lists) {
      const data = {
        action,
        query: this.fullQueryExp,
        search: this.queryParams.search,
        list_ids: this.queryParams.listID ? [this.queryParams.listID] : null,
        target_list_ids: lists.map((l) => l.id),
      };

      if (preconfirm) {
        data.status = 'confirmed';
      }

      let fn = null;
      if (!this.bulk.all && this.bulk.checked.length > 0) {
        // If 'all' is not selected, perform by IDs.
        fn = this.$api.addSubscribersToLists;
        data.ids = this.bulk.checked.map((s) => s.id);
      } else {
        // 'All' is selected, perform by query.
        data.query = this.queryParams.queryExp;
        data.subscription_status = this.queryParams.subStatus;
        fn = this.$api.addSubscribersToListsByQuery;
      }

      fn(data).then(() => {
        this.querySubscribers();
        this.$utils.toast(this.$t('subscribers.listChangeApplied'));
      });
    },
  },

  computed: {
    ...mapState(useMainStore, ['refreshTick', 'subscribers', 'lists', 'loading']),

    numSelectedSubscribers() {
      if (this.bulk.all) {
        return this.subscribers.total;
      }
      return this.bulk.checked.length;
    },

    // Returns the list that the subscribers are being filtered by in.
    currentList() {
      if (!this.queryParams.listID || !this.lists.results) {
        return null;
      }

      return this.lists.results.find((l) => l.id === this.queryParams.listID);
    },
  },

  watch: {
    refreshTick() { this.querySubscribers(); },
  },

  mounted() {
    if (this.$route.params.listID) {
      this.queryParams.listID = parseInt(this.$route.params.listID, 10);
    }
    if (this.$route.query.subscription_status) {
      this.queryParams.subStatus = this.$route.query.subscription_status;
    }

    if (this.$route.params.id) {
      this.$api.getSubscriber(parseInt(this.$route.params.id, 10)).then((data) => {
        this.showEditForm(data);
      });
    } else {
      this.querySubscribers();
    }
  },
};
</script>

<style scoped lang="scss">
.subs-page { display: flex; flex-direction: column; gap: 1.5rem; }

// Page header
.page-header { display: flex; align-items: flex-start; justify-content: space-between; gap: 1rem; }
.page-header-left { display: flex; flex-direction: column; gap: 0.25rem; }
.page-title {
  font-size: 1.4rem; font-weight: 700; color: #0f172a; margin: 0;
  display: flex; align-items: center; gap: 0.5rem; flex-wrap: wrap;
}
.page-title-count {
  font-size: 0.85rem; font-weight: 600; background: #f1f5f9; color: #64748b;
  border: 1px solid #e2e8f0; border-radius: 999px; padding: 0.1rem 0.55rem;
}
.page-title-sub { color: #94a3b8; font-weight: 400; font-size: 1rem; }

// Table card
.table-card { background: #fff; border: 1px solid #e2e8f0; border-radius: 12px; overflow: hidden; }

// Search toolbar
.search-toolbar {
  display: flex; align-items: center; gap: 0.75rem; padding: 1rem 1rem 0;
  flex-wrap: wrap; border-bottom: 1px solid #f1f5f9;
  padding-bottom: 1rem;
}
.search-form { flex: 1; min-width: 220px; max-width: 400px; }
:deep(.search-input) { width: 100%; }
.advanced-toggle {
  font-size: 0.8rem; color: #64748b; text-decoration: none; display: inline-flex; align-items: center; gap: 0.3rem;
  &:hover { color: #3b82f6; }
}
.toolbar-btn {
  display: inline-flex; align-items: center; gap: 0.35rem; padding: 0.4rem 0.75rem;
  border: 1px solid #e2e8f0; border-radius: 7px; background: #fff; color: #64748b;
  font-size: 0.8rem; cursor: pointer; white-space: nowrap;
  &:hover { background: #f8fafc; color: #0f172a; }
}

// Advanced panel
.advanced-panel {
  padding: 1rem; background: #f8fafc; border-bottom: 1px solid #e2e8f0;
}
.advanced-footer { display: flex; align-items: center; justify-content: space-between; margin-top: 0.5rem; }
.advanced-help { font-size: 0.78rem; color: #94a3b8; a { color: #64748b; } }

// Bulk bar
.bulk-bar {
  display: flex; align-items: center; gap: 0.5rem; flex-wrap: wrap;
}
.bulk-count { font-size: 0.85rem; color: #64748b; a { color: #3b82f6; } }
.bulk-btn {
  display: inline-flex; align-items: center; gap: 0.35rem; padding: 0.3rem 0.65rem;
  border-radius: 6px; font-size: 0.8rem; font-weight: 500; border: 1px solid; cursor: pointer; background: #fff;
  color: #64748b; border-color: #e2e8f0;
  &:hover { background: #f8fafc; }
  &--danger { color: #ef4444; border-color: #fecaca; &:hover { background: #fef2f2; } }
  &--warn { color: #f59e0b; border-color: #fde68a; &:hover { background: #fffbeb; } }
}

// Row cells
.email-cell { display: flex; align-items: center; gap: 0.4rem; flex-wrap: wrap; }
.row-name { font-weight: 500; color: #3b82f6; text-decoration: none; &:hover { text-decoration: underline; } }
.row-name--blocked { color: #ef4444; text-decoration: line-through; }
.list-tags { display: flex; flex-wrap: wrap; gap: 0.25rem; margin-top: 0.3rem; a { text-decoration: none; } }
.list-count { font-weight: 600; color: #0f172a; }
.date-cell { font-size: 0.82rem; color: #64748b; }

// Row actions
.row-actions {
  display: flex; align-items: center; justify-content: flex-end; gap: 0.25rem;
  opacity: 0; transition: opacity 0.15s;
}
:deep(tr:hover) .row-actions { opacity: 1; }
.row-action-btn {
  display: inline-flex; align-items: center; justify-content: center;
  width: 30px; height: 30px; border-radius: 6px; border: 1px solid #e2e8f0;
  background: #fff; color: #64748b; text-decoration: none; cursor: pointer;
  transition: background 0.12s, border-color 0.12s, color 0.12s;
  i { font-size: 0.8rem; }
  &:hover { background: #f8fafc; border-color: #cbd5e1; color: #0f172a; }
  &--danger:hover { background: #fef2f2; border-color: #fecaca; color: #ef4444; }
}
</style>
