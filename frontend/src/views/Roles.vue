<template>
  <div class="roles-page">
    <div class="page-header">
      <h1 class="page-title">
        {{ $t(isUser ? 'users.userRoles' : 'users.listRoles') }}
        <span v-if="!isNaN(roles.length)" class="page-title-count">{{ roles.length }}</span>
      </h1>
      <PvButton v-if="$can('users:manage')" severity="primary" icon="pi pi-plus"
        data-cy="btn-new" @click="showNewForm('user')" :label="$t('globals.buttons.new')" />
    </div>

    <div class="table-card">
      <PvDataTable :value="roles" :loading="isLoading">
        <PvColumn field="role" :header="$tc('users.role')" sortable>
          <template #body="{ data }">
            <div class="role-name-cell">
              <a href="#" class="row-name" @click.prevent="showEditForm(data, 'user')">{{ data.name }}</a>
              <PvTag v-if="data.id === 1" severity="success" size="small" value="Default" />
            </div>
          </template>
        </PvColumn>

        <PvColumn field="created_at" :header="$t('globals.fields.createdAt')"
          header-class="cy-created_at" sortable style="width:11rem">
          <template #body="{ data }">{{ $utils.niceDate(data.createdAt) }}</template>
        </PvColumn>

        <PvColumn field="updated_at" :header="$t('globals.fields.updatedAt')"
          header-class="cy-updated_at" sortable style="width:11rem">
          <template #body="{ data }">{{ $utils.niceDate(data.updatedAt) }}</template>
        </PvColumn>

        <PvColumn style="width:7rem; text-align:right">
          <template #body="{ data }">
            <div v-if="$can('roles:manage')" class="row-actions">
              <button type="button" class="row-action-btn" data-cy="btn-clone"
                v-tooltip.bottom="$t('globals.buttons.clone')"
                @click="$utils.prompt($t('globals.buttons.clone'), { placeholder: $t('globals.fields.name'), value: $t('campaigns.copyOf', { name: data.name }) }, (name) => onCloneRole(name, data))">
                <i class="pi pi-copy" />
              </button>
              <template v-if="data.id !== 1">
                <button type="button" class="row-action-btn" data-cy="btn-edit"
                  v-tooltip.bottom="$t('globals.buttons.edit')" @click="showEditForm(data, 'user')">
                  <i class="pi pi-pencil" />
                </button>
                <button type="button" class="row-action-btn row-action-btn--danger" data-cy="btn-delete"
                  v-tooltip.bottom="$t('globals.buttons.delete')" @click="onDeleteRole(data)">
                  <i class="pi pi-trash" />
                </button>
              </template>
            </div>
          </template>
        </PvColumn>

        <template #empty v-if="!isLoading">
          <empty-placeholder />
        </template>
      </PvDataTable>
    </div>

    <!-- Add / edit form modal -->
    <PvDialog v-model:visible="isFormVisible" :style="{ width: '700px' }" show-header="false" :closable="false" modal @hide="onFormClose">
      <role-form :data="curItem" :type="curType" :is-editing="isEditing" @finished="formFinished" @close="isFormVisible = false" />
    </PvDialog>
  </div>
</template>

<script>
import { mapState } from 'pinia';
import { useMainStore } from '../store';
import EmptyPlaceholder from '../components/EmptyPlaceholder.vue';
import RoleForm from './RoleForm.vue';

export default {
  components: {
    EmptyPlaceholder,
    RoleForm,
  },

  data() {
    return {
      curItem: null,
      curType: null,
      isEditing: false,
      isFormVisible: false,
    };
  },

  methods: {
    fetchRoles() {
      if (this.isUser) {
        this.$api.getUserRoles();
      } else {
        this.$api.getListRoles();
      }
    },

    // Show the edit form.
    showEditForm(item) {
      this.curItem = item;
      this.curType = this.isUser ? 'user' : 'list';
      this.isFormVisible = true;
      this.isEditing = true;
    },

    // Show the new form.
    showNewForm() {
      this.isEditing = false;
      this.isFormVisible = true;
    },

    formFinished() {
      this.fetchRoles();
    },

    onFormClose() {
      if (this.$route.params.id) {
        this.$router.push({ name: 'users' });
      }
    },

    onCloneRole(name, item) {
      const form = { name };
      let fn;
      if (this.isUser) {
        fn = this.$api.createUserRole;
        form.permissions = item.permissions;
      } else {
        fn = this.$api.createListRole;
        form.lists = item.lists;
      }

      fn(form).then(() => {
        this.fetchRoles();
        this.$utils.toast(this.$t('globals.messages.created', { name }));
      });
    },

    onDeleteRole(item) {
      this.$utils.confirm(
        this.$t('globals.messages.confirm'),
        () => {
          this.$api.deleteRole(item.id).then(() => {
            this.fetchRoles();

            this.$utils.toast(this.$t('globals.messages.deleted', { name: item.name }));
          });
        },
      );
    },

  },

  computed: {
    ...mapState(useMainStore, ['loading', 'userRoles', 'listRoles']),

    isLoading {
      return this.curType === 'user' ? this.loading.userRoles : this.loading.listRoles;
    },

    isUser() {
      return this.curType === 'user';
    },

    isList() {
      return this.curType === 'list';
    },

    roles() {
      return this.isUser ? this.userRoles : this.listRoles;
    },
  },

  mounted() {
    this.curType = this.$route.name === 'userRoles' ? 'user' : 'list';
    this.fetchRoles();
  },
};
</script>

<style scoped lang="scss">
.roles-page { display: flex; flex-direction: column; gap: 1.5rem; }

.role-name-cell { display: flex; align-items: center; gap: 0.5rem; }
.row-name { color: var(--lm-text); font-weight: 500; text-decoration: none; &:hover { color: var(--lm-primary); } }
</style>
