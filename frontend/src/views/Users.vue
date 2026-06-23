<template>
  <section class="users">
    <header class="grid page-header">
      <div class="col-10">
        <h1 class="title is-4">
          {{ $t('globals.terms.users') }}
          <span v-if="!isNaN(users.length)">({{ users.length }})</span>
        </h1>
      </div>
      <div class="col has-text-right">
        <div v-if="$can('users:manage')" class="field">
          <PvButton severity="primary" icon="pi pi-plus" class="btn-new" @click="showNewForm" data-cy="btn-new"
            :label="$t('globals.buttons.new')" />
        </div>
      </div>
    </header>

    <PvDataTable :value="users" :loading="loading.users" :rows="20"
      sort-field="createdAt" sort-order="1"
      v-model:selection="checked"
      data-key="id">
      <template #header>
        <div class="grid">
          <div class="col-6">
            <form @submit.prevent="getUsers">
              <div class="field">
                <div class="p-inputgroup">
                  <PvInputText v-model="queryParams.query" name="query" ref="query"
                    data-cy="query" />
                  <PvButton type="submit" severity="primary" icon="pi pi-search" data-cy="btn-query" />
                </div>
              </div>
            </form>
          </div>
        </div>
      </template>

      <PvColumn selection-mode="multiple" header-style="width:3rem" />

      <PvColumn field="username" :header="$t('users.username')" header-class="cy-username" sortable>
        <template #body="{ data }">
          <a :href="`/users/${data.id}`" @click.prevent="showEditForm(data)"
            :class="{ 'has-text-grey': data.status === 'disabled' }">
            {{ data.username }}
          </a>
          <PvTag v-if="data.status === 'disabled'" class="ml-1">
            {{ $t(`users.status.${data.status}`) }}
          </PvTag>
          <PvTag v-if="data.type === 'api'" class="api ml-1">
            <i class="pi pi-code mr-1" />
            {{ $t(`users.type.${data.type}`) }}
          </PvTag>
          <div class="has-text-grey is-size-7 mt-2">
            {{ data.name }}
          </div>
        </template>
      </PvColumn>

      <PvColumn field="status" :header="$tc('users.role')" header-class="cy-status" sortable>
        <template #body="{ data }">
          <router-link :to="{ name: 'userRoles' }">
            <PvTag :class="data.userRole.id === 1 ? 'enabled' : 'primary'">
              <i class="pi pi-user mr-1" />
              {{ data.userRole.name }}
            </PvTag>
          </router-link>
          <router-link :to="{ name: 'listRoles' }">
            <PvTag v-if="data.listRole" class="ml-1">
              <i class="pi pi-list mr-1" />
              {{ data.listRole.name }}
            </PvTag>
          </router-link>
        </template>
      </PvColumn>

      <PvColumn field="name" :header="$t('subscribers.email')" header-class="cy-name" sortable>
        <template #body="{ data }">
          <div>
            <a v-if="data.email" :href="`/users/${data.id}`" @click.prevent="showEditForm(data)"
              :class="{ 'has-text-grey': data.status === 'disabled' }">
              {{ data.email }}
            </a>
            <template v-else>
              —
            </template>
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

      <PvColumn field="last_login" :header="$t('users.lastLogin')" header-class="cy-updated_at" sortable>
        <template #body="{ data }">
          {{ data.loggedinAt ? $utils.niceDate(data.loggedinAt, true) : '—' }}
        </template>
      </PvColumn>

      <PvColumn header-class="actions" body-class="actions" align-frozen="right">
        <template #body="{ data }">
          <div>
            <a v-if="$can('users:manage')" href="#" @click.prevent="showEditForm(data)" data-cy="btn-edit"
              :aria-label="$t('globals.buttons.edit')">
              <i class="pi pi-pencil" v-tooltip.bottom="$t('globals.buttons.edit')" />
            </a>

            <a v-if="$can('users:manage')" href="#" @click.prevent="deleteUser(data)" data-cy="btn-delete"
              :aria-label="$t('globals.buttons.delete')">
              <i class="pi pi-trash" v-tooltip.bottom="$t('globals.buttons.delete')" />
            </a>
          </div>
        </template>
      </PvColumn>

      <template #empty v-if="!loading.users">
        <empty-placeholder />
      </template>
    </PvDataTable>

    <!-- Add / edit form modal -->
    <PvDialog v-model:visible="isFormVisible" :style="{ width: '600px' }" :closable="true" modal @hide="onFormClose">
      <user-form :data="curItem" :is-editing="isEditing" @finished="formFinished" />
    </PvDialog>
  </section>
</template>

<script>
import { mapState } from 'vuex';
import EmptyPlaceholder from '../components/EmptyPlaceholder.vue';

import UserForm from './UserForm.vue';

export default {
  components: {
    EmptyPlaceholder,
    UserForm,
  },

  data() {
    return {
      curItem: null,
      isEditing: false,
      isFormVisible: false,
      users: [],
      checked: [],
      queryParams: {
        page: 1,
        query: '',
        orderBy: 'id',
        order: 'asc',
      },
    };
  },

  methods: {
    onSort(field, direction) {
      this.queryParams.orderBy = field;
      this.queryParams.order = direction;
      this.getUsers();
    },

    onTableCheck() {
      // Disable bulk.all selection if there are no rows checked in the table.
      if (this.bulk.checked.length !== this.subscribers.total) {
        this.bulk.all = false;
      }
    },

    // Show the edit form.
    showEditForm(item) {
      this.curItem = item;
      this.isFormVisible = true;
      this.isEditing = true;
    },

    // Show the new form.
    showNewForm() {
      this.curItem = {};
      this.isFormVisible = true;
      this.isEditing = false;
    },

    formFinished() {
      this.getUsers();
    },

    onFormClose() {
      if (this.$route.params.id) {
        this.$router.push({ name: 'users' });
      }
    },

    getUsers() {
      this.$api.queryUsers({
        query: this.queryParams.query.replace(/[^\p{L}\p{N}\s]/gu, ' '),
        order_by: this.queryParams.orderBy,
        order: this.queryParams.order,
      }).then((resp) => {
        this.users = resp;
      });
    },

    deleteUser(item) {
      this.$utils.confirm(
        this.$t('globals.messages.confirm'),
        () => {
          this.$api.deleteUser(item.id).then(() => {
            this.getUsers();

            this.$utils.toast(this.$t('globals.messages.deleted', { name: item.name }));
          });
        },
      );
    },
  },

  computed: {
    ...mapState(['loading', 'settings']),
  },

  created() {
    this.$root.$on('page.refresh', this.getUsers);
  },

  unmounted() {
    this.$root.$off('page.refresh', this.getUsers);
  },

  mounted() {
    if (this.$route.params.id) {
      this.$api.getUser(parseInt(this.$route.params.id, 10)).then((data) => {
        this.showEditForm(data);
      });
    } else {
      this.getUsers();
    }
  },
};
</script>
