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
      <div class="lm-field">
        <label class="lm-label">{{ $t('globals.fields.name') }}</label>
        <PvInputText :disabled="disabled" :maxlength="200" v-model="form.name" name="name" ref="focus"
          required class="w-full" />
      </div>

      <div v-if="type === 'list'" class="form-section">
        <p class="section-label">{{ $t('users.listPerms') }}</p>

        <div class="list-add-row">
          <PvSelect v-model="form.curList" name="list"
            :placeholder="$tc('globals.terms.list')"
            :disabled="disabled || filteredLists.length < 1"
            :options="filteredLists"
            option-label="name"
            option-value="id"
            class="w-full" />
          <PvButton @click="onAddListPerm" :disabled="!form.curList" severity="primary"
            :label="$t('globals.buttons.add')" />
        </div>

        <div v-if="form.lists.length > 0 && (form.permissions['lists:get_all'] || form.permissions['lists:manage_all'])"
          class="perms-warning">
          <i class="pi pi-exclamation-triangle" />
          {{ $t('users.listPermsWarning') }}
        </div>

        <PvDataTable v-if="form.lists.length > 0" :value="form.lists">
          <PvColumn field="name" :header="$tc('globals.terms.list')">
            <template #body="{ data }">
              <router-link :to="`/lists/${data.id}`" target="_blank">{{ data.name }}</router-link>
            </template>
          </PvColumn>
          <PvColumn field="permissions" :header="$t('users.perms')" style="width:40%">
            <template #body="{ data }">
              <div class="check-row">
                <PvCheckbox v-model="data.permissions" value="list:get" :input-id="`list-get-${data.id}`" />
                <label :for="`list-get-${data.id}`" class="check-label">{{ $t('globals.buttons.view') }}</label>
              </div>
              <div class="check-row check-row--mt">
                <PvCheckbox v-model="data.permissions" value="list:manage" :input-id="`list-manage-${data.id}`" />
                <label :for="`list-manage-${data.id}`" class="check-label">{{ $t('globals.buttons.manage') }}</label>
              </div>
            </template>
          </PvColumn>
          <PvColumn style="width:3rem">
            <template #body="{ data }">
              <button type="button" class="row-action-btn row-action-btn--danger"
                @click="onDeleteListPerm(data.id)" v-tooltip.bottom="$t('globals.buttons.delete')">
                <i class="pi pi-trash" />
              </button>
            </template>
          </PvColumn>
        </PvDataTable>
      </div>

      <template v-if="type === 'user'">
        <div class="perms-header">
          <span class="section-label">{{ $t('users.perms') }}</span>
          <a v-if="!disabled" href="#" class="toggle-link" @click.prevent="onToggleSelect">
            {{ $t('globals.buttons.toggleSelect') }}
          </a>
        </div>

        <PvDataTable :value="serverConfig.permissions">
          <PvColumn field="group" :header="$t('users.roleGroup')" style="width:160px">
            <template #body="{ data }">
              <span class="group-label">{{ $tc(`globals.terms.${data.group}`) }}</span>
            </template>
          </PvColumn>
          <PvColumn field="permissions" :header="$t('users.perms')">
            <template #body="{ data }">
              <div v-for="p in data.permissions" :key="p" class="perm-row">
                <PvCheckbox v-model="form.permissions" :value="p" :input-id="`perm-${p}`" :disabled="disabled" />
                <label :for="`perm-${p}`" class="perm-label">
                  {{ p }}
                  <a v-if="p === 'subscribers:sql_query'"
                    href="https://listmonk.app/docs/roles-and-permissions/#subscriberssql_query" target="_blank"
                    rel="noopener noreferrer" aria-label="Warning: high risk permission">
                    <i class="pi pi-exclamation-triangle perm-warn-icon" />
                  </a>
                </label>
              </div>
            </template>
          </PvColumn>
        </PvDataTable>
      </template>

      <a href="https://listmonk.app/docs/roles-and-permissions" target="_blank" rel="noopener noreferrer"
        class="learn-more">
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
          item.permissions.forEach((p) => { acc.push(p); });
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
      if (this.$props.data.id === 1 || !this.$can('roles:manage')) {
        this.disabled = true;
      }
    } else {
      const skip = ['admin', 'users'];
      this.form.permissions = this.serverConfig.permissions.reduce((acc, item) => {
        if (skip.includes(item.group)) return acc;
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
.lm-field { display: flex; flex-direction: column; gap: 0.35rem; margin-bottom: 0; }
.lm-label { display: block; font-size: 0.8rem; font-weight: 600; color: var(--lm-text); }
.check-row { display: flex; align-items: center; gap: 0.5rem; &--mt { margin-top: 0.35rem; } }
.check-label { font-size: 0.875rem; color: var(--lm-text); cursor: pointer; }
.perm-warn-icon { color: var(--p-red-500); }

.section-label {
  font-size: 0.78rem;
  font-weight: 600;
  color: var(--lm-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.perms-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.toggle-link {
  font-size: 0.8rem;
  color: var(--lm-primary);
  text-decoration: none;
  &:hover { text-decoration: underline; }
}

.form-section {
  border: 1px solid var(--lm-border);
  border-radius: 8px;
  padding: 1rem 1.25rem;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.list-add-row {
  display: flex;
  gap: 0.75rem;
  align-items: center;
}

.perms-warning {
  font-size: 0.85rem;
  color: var(--p-red-500);
  display: flex;
  align-items: center;
  gap: 0.4rem;
}

.perm-row {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.15rem 0;
}

.perm-label {
  font-size: 0.85rem;
  font-family: monospace;
  display: flex;
  align-items: center;
  gap: 0.3rem;
}

.group-label {
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--lm-text);
}

.learn-more {
  font-size: 0.8rem;
  color: var(--lm-primary);
  text-decoration: none;
  &:hover { text-decoration: underline; }
}

.row-action-btn {
  background: none;
  border: none;
  cursor: pointer;
  padding: 0.3rem;
  border-radius: 4px;
  color: var(--lm-text-subtle);
  &--danger:hover { color: var(--p-red-500); background: var(--lm-danger-bg, #fef2f2); }
}
</style>
