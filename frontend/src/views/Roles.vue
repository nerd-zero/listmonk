<template>
  <section class="roles">
    <header class="grid page-header">
      <div class="col-10">
        <h1 class="title is-4">
          {{ $t(isUser ? 'users.userRoles' : 'users.listRoles') }}
          <span v-if="!isNaN(roles.length)">({{ roles.length }})</span>
        </h1>
      </div>
      <div class="col has-text-right">
        <div v-if="$can('users:manage')" class="field">
          <PvButton severity="primary" icon="pi pi-plus" class="btn-new" @click="showNewForm('user')"
            data-cy="btn-new" :label="$t('globals.buttons.new')" />
        </div>
      </div>
    </header>
    <PvDataTable :value="roles" :loading="isLoading()" hoverable>
      <PvColumn field="role" :header="$tc('users.role')" sortable>
        <template #body="{ data }">
          <a href="#" @click.prevent="showEditForm(data, 'user')">
            <PvTag v-if="data.id === 1" class="enabled">
              {{ data.name }}
            </PvTag>
            <template v-else>{{ data.name }}</template>
          </a>
        </template>
      </PvColumn>

      <PvColumn field="created_at" :header="$t('globals.fields.createdAt')"
        header-class="cy-created_at" sortable>
        <template #body="{ data }">
          {{ $utils.niceDate(data.createdAt) }}
        </template>
      </PvColumn>

      <PvColumn field="updated_at" :header="$t('globals.fields.updatedAt')"
        header-class="cy-updated_at" sortable>
        <template #body="{ data }">
          {{ $utils.niceDate(data.updatedAt) }}
        </template>
      </PvColumn>

      <PvColumn body-class="actions has-text-right">
        <template #body="{ data }">
          <template v-if="$can('roles:manage')">
            <a href="#" @click.prevent="$utils.prompt($t('globals.buttons.clone'),
              {
                placeholder: $t('globals.fields.name'),
                value: $t('campaigns.copyOf', { name: data.name }),
              },
              (name) => onCloneRole(name, data))" data-cy="btn-clone" :aria-label="$t('globals.buttons.clone')">
              <i class="pi pi-copy" v-tooltip.bottom="$t('globals.buttons.clone')" />
            </a>

            <template v-if="data.id !== 1">
              <a href="#" @click.prevent="showEditForm(data, 'user')" data-cy="btn-edit"
                :aria-label="$t('globals.buttons.edit')">
                <i class="pi pi-pencil" v-tooltip.bottom="$t('globals.buttons.edit')" />
              </a>

              <a href="#" @click.prevent="onDeleteRole(data)" data-cy="btn-delete"
                :aria-label="$t('globals.buttons.delete')">
                <i class="pi pi-trash" v-tooltip.bottom="$t('globals.buttons.delete')" />
              </a>
            </template>
          </template>
        </template>
      </PvColumn>

      <template #empty v-if="!isLoading()">
        <empty-placeholder />
      </template>
    </PvDataTable>

    <!-- Add / edit form modal -->
    <PvDialog v-model:visible="isFormVisible" :style="{ width: '700px' }" modal @hide="onFormClose">
      <role-form :data="curItem" :type="curType" :is-editing="isEditing" @finished="formFinished" />
    </PvDialog>
  </section>
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
    isLoading() {
      return this.curType === 'user' ? this.loading.userRoles : this.loading.listRoles;
    },

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
