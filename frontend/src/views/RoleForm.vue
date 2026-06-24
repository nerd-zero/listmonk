<template>
  <form class="lm-form" @submit.prevent="onSubmit">
    <div class="lm-form-header">
      <div class="lm-form-title-row">
        <h3 class="lm-form-title">
          {{ isEditing ? data.name : (type === 'user' ? $t('users.newUserRole') : $t('users.newListRole')) }}
        </h3>
      </div>
      <p v-if="isEditing" class="lm-form-meta">{{ $t('globals.fields.id') }}: <copy-text :text="`${data.id}`" /></p>
    </div>

    <div class="lm-form-body">
        <div class="field">
          <label class="block mb-1 text-sm font-medium">{{ $t('globals.fields.name') }}</label>
          <PvInputText autofocus :disabled="disabled" :maxlength="200" v-model="form.name" name="name" ref="focus"
            required />
        </div>

        <div v-if="type === 'list'" class="box">
          <h5>{{ $t('users.listPerms') }}</h5>
          <div class="mb-5">
            <div class="grid">
              <div class="col-9">
                <PvSelect v-model="form.curList" name="list"
                  :placeholder="$tc('globals.terms.list')"
                  :disabled="disabled || filteredLists.length < 1"
                  :options="filteredLists"
                  option-label="name"
                  option-value="id"
                  class="mb-3"
                  style="width:100%" />
              </div>
              <div class="col">
                <PvButton @click="onAddListPerm" :disabled="!form.curList" severity="primary" style="width:100%"
                  :label="$t('globals.buttons.add')" />
              </div>
            </div>
            <span
              v-if="form.lists.length > 0 && (form.permissions['lists:get_all'] || form.permissions['lists:manage_all'])"
              class="is-size-6 has-text-danger">
              <i class="pi pi-exclamation-triangle" />
              {{ $t('users.listPermsWarning') }}
            </span>
          </div>

          <PvDataTable :value="form.lists">
            <PvColumn field="name" :header="$tc('globals.terms.list')">
              <template #body="{ data }">
                <router-link :to="`/lists/${data.id}`" target="_blank">
                  {{ data.name }}
                </router-link>
              </template>
            </PvColumn>

            <PvColumn field="permissions" :header="$t('users.perms')" style="width:40%">
              <template #body="{ data }">
                <div class="flex items-center gap-2">
                  <PvCheckbox v-model="data.permissions" value="list:get" :input-id="`list-get-${data.id}`" />
                  <label :for="`list-get-${data.id}`">{{ $t('globals.buttons.view') }}</label>
                </div>
                <div class="flex items-center gap-2">
                  <PvCheckbox v-model="data.permissions" value="list:manage" :input-id="`list-manage-${data.id}`" />
                  <label :for="`list-manage-${data.id}`">{{ $t('globals.buttons.manage') }}</label>
                </div>
              </template>
            </PvColumn>

            <PvColumn style="width:10%">
              <template #body="{ data }">
                <a href="#" @click.prevent="onDeleteListPerm(data.id)" data-cy="btn-delete"
                  :aria-label="$t('globals.buttons.delete')">
                  <i class="pi pi-trash" v-tooltip.bottom="$t('globals.buttons.delete')" />
                </a>
              </template>
            </PvColumn>
          </PvDataTable>
        </div>

        <template v-if="type === 'user'">
          <div class="grid">
            <div class="col-7">
              <h5 class="mb-0">
                {{ $t('users.perms') }}
              </h5>
            </div>
            <div v-if="!disabled" style="text-align:right">
              <a href="#" @click.prevent="onToggleSelect">{{ $t('globals.buttons.toggleSelect') }}</a>
            </div>
          </div>

          <PvDataTable :value="serverConfig.permissions">
            <PvColumn field="group" :header="$t('users.roleGroup')">
              <template #body="{ data }">
                {{ $tc(`globals.terms.${data.group}`) }}
              </template>
            </PvColumn>

            <PvColumn field="permissions" header="Permissions">
              <template #body="{ data }">
                <div v-for="p in data.permissions" :key="p" class="flex items-center gap-2 mb-1">
                  <PvCheckbox v-model="form.permissions" :value="p" :input-id="`perm-${p}`" :disabled="disabled" />
                  <label :for="`perm-${p}`">
                    {{ p }}
                    <a v-if="p === 'subscribers:sql_query'"
                      href="https://listmonk.app/docs/roles-and-permissions/#subscriberssql_query" target="_blank"
                      rel="noopener noreferrer" aria-label="Warning: high risk permission">
                      <i class="pi pi-exclamation-triangle text-red-500" />
                    </a>
                  </label>
                </div>
              </template>
            </PvColumn>
          </PvDataTable>
        </template>
        <a href="https://listmonk.app/docs/roles-and-permissions" target="_blank" rel="noopener noreferrer">
          <i class="pi pi-external-link" /> {{ $t('globals.buttons.learnMore') }}
        </a>
    </div>

    <div class="lm-form-footer">
      <PvButton @click="$emit('close')" :label="$t('globals.buttons.close')" severity="secondary" />
      <PvButton v-if="!disabled" type="submit" severity="primary" :loading="loading.roles" data-cy="btn-save"
        :label="$t('globals.buttons.save')" />
    </div>
  </form>
</template>

<script>
import { mapState } from 'pinia';
import { useMainStore } from '../store';
import CopyText from '../components/CopyText.vue';

export default {
  name: 'RoleForm',

  components: {
    CopyText,
  },

  props: {
    data: { type: Object, default: () => ({}) },
    isEditing: { type: Boolean, default: false },
    type: { type: String, default: 'user' },
  },

  emits: ['finished', 'close'],

  data() {
    return {
      // Binds form input values.
      form: {
        curList: null,
        lists: [],
        name: null,
        permissions: {},
      },
      hasToggle: false,
      disabled: false,
    };
  },

  methods: {
    onAddListPerm() {
      const list = this.lists.results.find((l) => l.id === this.form.curList);
      this.form.lists.push({ id: list.id, name: list.name, permissions: ['list:get', 'list:manage'] });

      this.form.curList = (this.filteredLists.length > 0) ? this.filteredLists[0].id : null;
    },

    onDeleteListPerm(id) {
      this.form.lists = this.form.lists.filter((p) => p.id !== id);
      this.form.curList = (this.filteredLists.length > 0) ? this.filteredLists[0].id : null;
    },

    onSubmit() {
      if (this.isEditing) {
        this.updateRole();
        return;
      }

      this.createRole();
    },

    onToggleSelect() {
      if (this.hasToggle) {
        this.form.permissions = [];
      } else {
        this.form.permissions = this.serverConfig.permissions.reduce((acc, item) => {
          item.permissions.forEach((p) => {
            acc.push(p);
          });
          return acc;
        }, []);
      }

      this.hasToggle = !this.hasToggle;
    },

    createRole() {
      let fn;
      const form = { name: this.form.name };

      if (this.$props.type === 'user') {
        fn = this.$api.createUserRole;
        form.permissions = this.form.permissions;
      } else {
        fn = this.$api.createListRole;
        form.lists = this.form.lists.reduce((acc, item) => {
          acc.push({ id: item.id, permissions: item.permissions });
          return acc;
        }, []);
      }

      fn(form).then((data) => {
        this.$emit('finished');
        this.$utils.toast(this.$t('globals.messages.created', { name: data.name }));
        this.$emit('close');
      });
    },

    updateRole() {
      let fn;
      const form = { id: this.$props.data.id, name: this.form.name };

      if (this.$props.type === 'user') {
        fn = this.$api.updateUserRole;
        form.permissions = this.form.permissions;
      } else {
        fn = this.$api.updateListRole;
        form.lists = this.form.lists.reduce((acc, item) => {
          acc.push({ id: item.id, permissions: item.permissions });
          return acc;
        }, []);
      }

      fn(form).then((data) => {
        this.$emit('finished');
        this.$utils.toast(this.$t('globals.messages.updated', { name: data.name }));
        this.$emit('close');
      });
    },
  },

  computed: {
    ...mapState(useMainStore, ['loading', 'serverConfig', 'lists']),

    // Return the list of unselected lists.
    filteredLists() {
      if (!this.lists.results || this.type !== 'list') {
        return [];
      }

      const subIDs = this.form.lists.reduce((obj, item) => ({ ...obj, [item.id]: true }), {});
      return this.lists.results.filter((l) => (!(l.id in subIDs)));
    },

  },

  mounted() {
    if (this.isEditing) {
      this.form = { ...this.form, ...this.$props.data };

      // It's the superadmin role. Disable the form.
      if (this.$props.data.id === 1 || !this.$can('roles:manage')) {
        this.disabled = true;
      }
    } else {
      const skip = ['admin', 'users'];
      this.form.permissions = this.serverConfig.permissions.reduce((acc, item) => {
        if (skip.includes(item.group)) {
          return acc;
        }
        item.permissions.forEach((p) => {
          if (p !== 'subscribers:sql_query' && !p.startsWith('lists:') && !p.startsWith('settings:')) {
            acc.push(p);
          }
        });
        return acc;
      }, []);
    }

    this.$nextTick(() => {
      if (this.filteredLists.length > 0) {
        this.form.curList = this.filteredLists[0].id;
      }
      this.$refs.focus.$el.focus();
    });
  },
};
</script>

<style scoped lang="scss">

</style>
