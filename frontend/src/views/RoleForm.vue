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
        <PvInputText :disabled="disabled" :maxlength="200" v-model="form.name" name="name" ref="focusEl"
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

<script setup lang="ts">
import {
  ref, reactive, computed, nextTick, onMounted,
} from 'vue';
import { storeToRefs } from 'pinia';
import { useI18n } from 'vue-i18n';
import { useMainStore } from '../store';
import { useGlobal } from '../composables/useGlobal';
import CopyText from '../components/CopyText.vue';
import { getRoles as rolesApi } from '../api/generated/endpoints/roles/roles';

const props = withDefaults(defineProps<{
  data?: any;
  isEditing?: boolean;
  type?: string;
}>(), { data: () => ({}), isEditing: false, type: 'user' });

const emit = defineEmits(['finished', 'close']);

const { $utils, $can } = useGlobal();
const { createUserRole, updateUserRole, createListRole, updateListRole } = rolesApi();
const { t } = useI18n();
const { loading, serverConfig, lists } = storeToRefs(useMainStore());

const focusEl = ref<any>(null);
const hasToggle = ref(false);
const disabled = ref(false);
const form = reactive<any>({
  curList: null,
  lists: [],
  name: null,
  permissions: [],
});

const filteredLists = computed(() => {
  if (!(lists.value as any).results || props.type !== 'list') return [];
  const subIDs = form.lists.reduce((obj: any, item: any) => ({ ...obj, [item.id]: true }), {});
  return (lists.value as any).results.filter((l: any) => !(l.id in subIDs));
});

function onAddListPerm() {
  const list = (lists.value as any).results.find((l: any) => l.id === form.curList);
  form.lists.push({ id: list.id, name: list.name, permissions: ['list:get', 'list:manage'] });
  form.curList = filteredLists.value.length > 0 ? filteredLists.value[0].id : null;
}

function onDeleteListPerm(id: number) {
  form.lists = form.lists.filter((p: any) => p.id !== id);
  form.curList = filteredLists.value.length > 0 ? filteredLists.value[0].id : null;
}

function onToggleSelect() {
  if (hasToggle.value) {
    form.permissions = [];
  } else {
    form.permissions = (serverConfig.value as any).permissions.reduce((acc: string[], item: any) => {
      item.permissions.forEach((p: string) => { acc.push(p); });
      return acc;
    }, []);
  }
  hasToggle.value = !hasToggle.value;
}

function createRole() {
  let fn: (payload: any) => Promise<any>;
  const payload: any = { name: form.name };
  if (props.type === 'user') {
    fn = createUserRole;
    payload.permissions = form.permissions;
  } else {
    fn = createListRole;
    payload.lists = form.lists.map((item: any) => ({ id: item.id, permissions: item.permissions }));
  }
  fn(payload).then((data: any) => {
    emit('finished');
    $utils.toast(t('globals.messages.created', { name: data.name }));
    emit('close');
  });
}

function updateRole() {
  const payload: any = { name: form.name };
  let fn: (id: number, data: any) => Promise<any>;
  if (props.type === 'user') {
    fn = updateUserRole;
    payload.permissions = form.permissions;
  } else {
    fn = updateListRole;
    payload.lists = form.lists.map((item: any) => ({ id: item.id, permissions: item.permissions }));
  }
  fn(props.data.id, payload).then((data: any) => {
    emit('finished');
    $utils.toast(t('globals.messages.updated', { name: data.name }));
    emit('close');
  });
}

function onSubmit() {
  if (props.isEditing) { updateRole(); return; }
  createRole();
}

onMounted(() => {
  if (props.isEditing) {
    Object.assign(form, props.data);
    if (props.data.id === 1 || !$can('roles:manage')) {
      disabled.value = true;
    }
  } else {
    const skip = ['admin', 'users'];
    form.permissions = (serverConfig.value as any).permissions.reduce((acc: string[], item: any) => {
      if (skip.includes(item.group)) return acc;
      item.permissions.forEach((p: string) => {
        if (p !== 'subscribers:sql_query' && !p.startsWith('lists:') && !p.startsWith('settings:')) {
          acc.push(p);
        }
      });
      return acc;
    }, []);
  }
  nextTick(() => {
    if (filteredLists.value.length > 0) form.curList = filteredLists.value[0].id;
    focusEl.value?.$el?.focus();
  });
});
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
