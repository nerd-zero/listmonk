<template>
  <section class="subscribers">
    <header class="columns page-header">
      <div class="column is-10">
        <h1 class="title is-4">
          {{ $t('globals.terms.subscribers') }}
          <span v-if="!isNaN(subscribers.total)">
            (<span data-cy="count">{{ subscribers.total }}</span>)
          </span>
          <span v-if="currentList">
            &raquo; {{ currentList.name }}
            <span v-if="queryParams.subStatus" class="has-text-grey has-text-weight-normal is-capitalized">({{
              queryParams.subStatus }})</span>
          </span>
        </h1>
      </div>
      <div class="column has-text-right">
        <div v-if="$can('subscribers:manage')" class="field">
          <PvButton severity="primary" icon="pi pi-plus" @click="showNewForm" data-cy="btn-new" class="btn-new w-full"
            :label="$t('globals.buttons.new')" />
        </div>
      </div>
    </header>

    <section class="subscribers-controls">
      <div class="columns">
        <div class="column is-8">
          <form @submit.prevent="onSubmit">
            <div>
              <div class="field has-addons">
                <div class="control is-expanded">
                  <PvInputText @input="onSimpleQueryInput" v-model="queryInput" class="w-full"
                    :placeholder="$t('subscribers.queryPlaceholder')" ref="query"
                    :disabled="isSearchAdvanced" data-cy="search" />
                </div>
                <p class="controls">
                  <PvButton type="submit" severity="primary" icon="pi pi-search" :disabled="isSearchAdvanced"
                    data-cy="btn-search" />
                </p>
              </div>

              <div v-if="isSearchAdvanced">
                <PvTextarea v-model="queryParams.queryExp" @keydown="onAdvancedQueryEnter" rows="3"
                  ref="queryExp" placeholder="subscribers.name LIKE '%user%' or subscribers.status='blocklisted'"
                  class="w-full" data-cy="query" />
                <span class="is-size-6 has-text-grey">
                  {{ $t('subscribers.advancedQueryHelp') }}.{{ ' ' }}
                  <a href="https://listmonk.app/docs/querying-and-segmentation" target="_blank"
                    rel="noopener noreferrer">
                    {{ $t('globals.buttons.learnMore') }}.
                  </a>
                </span>
                <div class="buttons">
                  <PvButton type="submit" severity="primary" icon="pi pi-search" data-cy="btn-query"
                    :label="$t('subscribers.query')" />
                  <PvButton @click.prevent="toggleAdvancedSearch" icon="pi pi-times" data-cy="btn-query-reset"
                    :label="$t('subscribers.reset')" />
                </div>
              </div><!-- advanced query -->
            </div>
          </form>
          <div v-if="!isSearchAdvanced" class="toggle-advanced">
            <a href="#" @click.prevent="toggleAdvancedSearch" data-cy="btn-advanced-search">
              <i class="pi pi-cog" />
              {{ $t('subscribers.advancedQuery') }}
            </a>
          </div>
        </div><!-- search -->
      </div>
    </section><!-- control -->

    <br />
    <PvDataTable :value="subscribers.results ?? []" :loading="loading.subscribers"
      data-key="id"
      :rows="subscribers.perPage" :paginator="true" paginator-position="both"
      :total-records="subscribers.total" :lazy="true"
      @page="(e) => onPageChange(e.page + 1)"
      @sort="(e) => onSort(e.sortField, e.sortOrder === 1 ? 'asc' : 'desc')"
      selection-mode="checkbox" v-model:selection="bulk.checked"
      @row-select="onTableCheck" @row-unselect="onTableCheck"
      @row-select-all="onTableCheck" @row-unselect-all="onTableCheck"
      :current-page-report-template="'{first} - {last} of {totalRecords}'"
      hoverable>
      <template #header>
        <div class="actions">
          <a class="a" href="#" @click.prevent="exportSubscribers" data-cy="btn-export-subscribers">
            <i class="pi pi-download" />
            {{ $t('subscribers.export') }}
          </a>
          <template v-if="bulk.checked.length > 0">
            <a class="a" href="#" @click.prevent="showBulkListForm" data-cy="btn-manage-lists">
              <i class="pi pi-list" /> Manage lists
            </a>
            <a class="a" href="#" @click.prevent="deleteSubscribers" data-cy="btn-delete-subscribers">
              <i class="pi pi-trash" /> Delete
            </a>
            <a class="a" href="#" @click.prevent="blocklistSubscribers" data-cy="btn-manage-blocklist">
              <i class="pi pi-user-minus" /> Blocklist
            </a>
            <span class="a">
              {{ $t('globals.messages.numSelected', { num: numSelectedSubscribers }) }}
              <span v-if="!bulk.all && subscribers.total > subscribers.perPage">
                &mdash;
                <a href="#" @click.prevent="selectAllSubscribers">
                  {{ $t('globals.messages.selectAll', { num: subscribers.total }) }}
                </a>
              </span>
            </span>
          </template>
        </div>
      </template>

      <PvColumn selection-mode="multiple" header-style="width:3rem" />

      <PvColumn field="email" :header="$t('subscribers.email')" header-class="cy-email" sortable>
        <template #body="{ data }">
          <a :href="`/subscribers/${data.id}`" @click.prevent="showEditForm(data)"
            :class="{ 'blocklisted': data.status === 'blocklisted' }">
            {{ data.email }}
            <copy-text :text="`${data.email}`" hide-text />
          </a>
          <PvTag v-if="data.status !== 'enabled'" :class="data.status" data-cy="blocklisted"
            :value="$t(`subscribers.status.${data.status}`)" />
          <span class="tags">
            <template v-for="l in data.lists" :key="l.id">
              <router-link :to="`/subscribers/lists/${l.id}`" style="padding-right:0.5em;">
                <PvTag :class="l.subscriptionStatus" size="small">
                  {{ l.name }}
                  <sup v-if="l.optin === 'double' || l.subscriptionStatus == 'unsubscribed'">
                    {{ $t(`subscribers.status.${l.subscriptionStatus}`) }}
                  </sup>
                </PvTag>
              </router-link>
            </template>
          </span>
        </template>
      </PvColumn>

      <PvColumn field="name" :header="$t('globals.fields.name')" header-class="cy-name" sortable>
        <template #body="{ data }">
          <a :href="`/subscribers/${data.id}`" @click.prevent="showEditForm(data)"
            :class="{ 'blocklisted': data.status === 'blocklisted' }">
            {{ data.name }}
            <copy-text :text="`${data.name}`" hide-text />
          </a>
        </template>
      </PvColumn>

      <PvColumn field="lists" :header="$t('globals.terms.lists')" header-class="cy-lists" style="text-align:center">
        <template #body="{ data }">
          {{ listCount(data.lists) }}
        </template>
      </PvColumn>

      <PvColumn field="created_at" :header="$t('globals.fields.createdAt')" header-class="cy-created_at" sortable>
        <template #body="{ data }">
          {{ $utils.niceDate(data.createdAt) }}
        </template>
      </PvColumn>

      <PvColumn field="updated_at" :header="$t('globals.fields.updatedAt')" header-class="cy-updated_at" sortable>
        <template #body="{ data }">
          {{ $utils.niceDate(data.updatedAt) }}
        </template>
      </PvColumn>

      <PvColumn body-class="actions" style="text-align:right">
        <template #body="{ data }">
          <div>
            <a :href="`/api/subscribers/${data.id}/export`" data-cy="btn-download"
              :aria-label="$t('subscribers.downloadData')">
              <i class="pi pi-download" v-tooltip.bottom="$t('subscribers.downloadData')" />
            </a>
            <a v-if="$can('subscribers:manage')" :href="`/subscribers/${data.id}`"
              @click.prevent="showEditForm(data)" data-cy="btn-edit" :aria-label="$t('globals.buttons.edit')">
              <i class="pi pi-pencil" v-tooltip.bottom="$t('globals.buttons.edit')" />
            </a>
            <a v-if="$can('subscribers:manage')" href="#" @click.prevent="deleteSubscriber(data)"
              data-cy="btn-delete" :aria-label="$t('globals.buttons.delete')">
              <i class="pi pi-trash" v-tooltip.bottom="$t('globals.buttons.delete')" />
            </a>
          </div>
        </template>
      </PvColumn>

      <template #empty v-if="!loading.subscribers">
        <empty-placeholder />
      </template>
    </PvDataTable>

    <!-- Manage list modal -->
    <PvDialog v-model:visible="isBulkListFormVisible" :style="{ width: '500px' }" :closable="true" modal
      class="has-overflow">
      <subscriber-bulk-list :num-subscribers="this.numSelectedSubscribers" @finished="bulkChangeLists" />
    </PvDialog>

    <!-- Add / edit form modal -->
    <PvDialog v-model:visible="isFormVisible" :style="{ width: '850px' }" :closable="true" modal
      @hide="onFormClose">
      <subscriber-form :data="curItem" :is-editing="isEditing" @finished="querySubscribers" />
    </PvDialog>
  </section>
</template>

<script>
import { mapState } from 'vuex';
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
    ...mapState(['subscribers', 'lists', 'loading']),

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

  created() {
    this.$root.$on('page.refresh', this.querySubscribers);
  },

  unmounted() {
    this.$root.$off('page.refresh', this.querySubscribers);
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
      // Get subscribers on load.
      this.querySubscribers();
    }
  },
};
</script>
