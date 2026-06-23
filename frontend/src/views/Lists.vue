<template>
  <section class="lists">
    <header class="grid page-header">
      <div class="col-10">
        <h1 class="title is-4 mb-2">
          {{ $t('globals.terms.lists') }}
          <span v-if="queryParams.status === 'archived'" class="has-text-grey-light">/ {{ queryParams.status }} </span>
          <span v-if="!isNaN(lists.total)">({{ lists.total }})</span>
        </h1>

        <div class="is-size-7">
          <router-link v-if="queryParams.status !== 'archived'" :to="{ name: 'lists', query: { status: 'archived' } }">
            {{ $t('globals.buttons.view') }} {{ $t('lists.archived').toLowerCase() }} &rarr;
          </router-link>
          <router-link v-else :to="{ name: 'lists' }">
            {{ $t('globals.buttons.view') }} {{ $t('menu.allLists').toLowerCase() }} &rarr;
          </router-link>
        </div>
      </div>
      <div class="col has-text-right">
        <div v-if="$can('lists:manage_all')" class="field">
          <PvButton severity="primary" icon="pi pi-plus" class="btn-new" @click="showNewForm" data-cy="btn-new"
            :label="$t('globals.buttons.new')" />
        </div>
      </div>
    </header>

    <PvDataTable :value="lists.results" :loading="loading.listsFull"
      v-model:selection="bulk.checked"
      selection-mode="checkbox"
      data-key="id"
      :paginator="true"
      paginator-position="both"
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
      @row-unselect-all="onTableCheck">
<template #header>
        <div class="grid">
          <div class="col-6">
            <form @submit.prevent="getLists">
              <div class="field has-addons">
                <div class="control is-expanded">
                  <PvInputText v-model="queryParams.query" name="query" ref="query" data-cy="query" class="w-full" />
                </div>
                <div class="control">
                  <PvButton type="submit" severity="primary" icon="pi pi-search" data-cy="btn-query" />
                </div>
              </div>
            </form>
          </div>
        </div>
        <div class="actions" v-if="bulk.checked.length > 0">
          <a class="a" href="#" @click.prevent="deleteLists" data-cy="btn-delete-lists">
            <i class="pi pi-trash" /> {{ $t('globals.buttons.delete') }}
          </a>
          <span class="a">
            {{ $tc('globals.messages.numSelected', numSelectedLists, { num: numSelectedLists }) }}
            <span v-if="!bulk.all && lists.total > lists.perPage">
              &mdash;
              <a href="#" @click.prevent="onSelectAll" data-cy="select-all-lists">
                {{ $tc('globals.messages.selectAll', lists.total, { num: lists.total }) }}
              </a>
            </span>
          </span>
        </div>
      </template>

      <PvColumn selection-mode="multiple" header-style="width:3rem" />

      <PvColumn field="name" :header="$t('globals.fields.name')" header-class="cy-name" sortable style="width:25%">
        <template #body="{ data }">
          <div>
            <a :href="`/lists/${data.id}`" @click.prevent="showEditForm(data)">
              {{ data.name }}
            </a>
            <div class="tags">
              <PvTag v-for="t in data.tags" :key="t" :value="t" class="is-small" />
            </div>
          </div>
        </template>
      </PvColumn>

      <PvColumn field="type" :header="$t('globals.fields.type')" header-class="cy-type" sortable style="width:15%">
        <template #body="{ data }">
          <div class="tags">
            <PvTag :class="data.type" :data-cy="`type-${data.type}`" :value="$t(`lists.types.${data.type}`)" />
            {{ ' ' }}

            <PvTag :class="data.optin" :data-cy="`optin-${data.optin}`">
              <i :class="data.optin === 'double' ? 'pi pi-user-plus' : 'pi pi-user-minus'" />
              {{ ' ' }}
              {{ $t(`lists.optins.${data.optin}`) }}
            </PvTag>{{ ' ' }}

            <a v-if="data.optin === 'double'" class="is-size-7 send-optin" href="#"
              @click="$utils.confirm(null, () => createOptinCampaign(data))" data-cy="btn-send-optin-campaign">
              <i class="pi pi-send" v-tooltip.bottom="$t('lists.sendOptinCampaign')" />
              {{ $t('lists.sendOptinCampaign') }}
            </a>
          </div>
        </template>
      </PvColumn>

      <PvColumn field="subscriber_count" :header="$t('globals.terms.subscribers')" header-class="cy-subscribers"
        sortable>
        <template #body="{ data }">
          <template v-if="$can('subscribers:get_all', 'subscribers:get')">
            <router-link :to="`/subscribers/lists/${data.id}`">
              {{ $utils.formatNumber(data.subscriberCount) }}
              <span class="is-size-7 view">{{ $t('globals.buttons.view') }}</span>
            </router-link>
          </template>
          <template v-else>
            {{ $utils.formatNumber(data.subscriberCount) }}
          </template>
        </template>
      </PvColumn>

      <PvColumn field="subscriber_counts" header-class="cy-subscribers" style="width:10%">
        <template #body="{ data }">
          <div class="fields stats">
            <p v-for="(count, status) in filterStatuses(data)" :key="status">
              <label for="#">{{ $tc(`subscribers.status.${status}`, count) }}</label>
              <router-link :to="`/subscribers/lists/${data.id}?subscription_status=${status}`" :class="status">
                {{ $utils.formatNumber(count) }}
              </router-link>
            </p>
          </div>
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
            <router-link v-if="$can('campaigns:manage')" :to="`/campaigns/new?list_id=${data.id}`"
              data-cy="btn-campaign">
              <i class="pi pi-send" v-tooltip.bottom="$t('lists.sendCampaign')" />
            </router-link>

            <a v-if="$can('lists:manage') || $canList(data.id, 'list:manage')" href="#"
              @click.prevent="showEditForm(data)" data-cy="btn-edit" :aria-label="$t('globals.buttons.edit')">
              <i class="pi pi-pencil" v-tooltip.bottom="$t('globals.buttons.edit')" />
            </a>

            <router-link v-if="$can('subscribers:import')" :to="{ name: 'import', query: { list_id: data.id } }"
              data-cy="btn-import">
              <i class="pi pi-upload" v-tooltip.bottom="$t('import.title')" />
            </router-link>

            <a v-if="$can('lists:manage') || $canList(data.id, 'list:manage')" href="#"
              @click.prevent="deleteList(data)" data-cy="btn-delete" :aria-label="$t('globals.buttons.delete')">
              <i class="pi pi-trash" v-tooltip.bottom="$t('globals.buttons.delete')" />
            </a>
          </div>
        </template>
      </PvColumn>

      <template #empty v-if="!loading.listsFull">
        <empty-placeholder />
      </template>
    </PvDataTable>

    <!-- Add / edit form modal -->
    <PvDialog v-model:visible="isFormVisible" :style="{ width: '600px' }" :closable="true" modal
      @hide="onFormClose">
      <list-form :data="curItem" :is-editing="isEditing" @finished="formFinished" />
    </PvDialog>

    <p v-if="settings['app.cache_slow_queries']" class="has-text-grey">
      *{{ $t('globals.messages.slowQueriesCached') }}
      <a href="https://listmonk.app/docs/maintenance/performance/" target="_blank" rel="noopener noreferer"
        class="has-text-grey">
        <i class="pi pi-external-link" /> {{ $t('globals.buttons.learnMore') }}
      </a>
    </p>
  </section>
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
