<template>
  <form class="lm-form" @submit.prevent="onSubmit">
    <div class="lm-form-header">
      <div class="lm-form-title-row">
        <h3 class="lm-form-title">{{ isEditing ? data.name : $t('users.newUser') }}</h3>
      </div>
      <p v-if="isEditing" class="lm-form-meta">{{ $t('globals.fields.id') }}: <copy-text :text="`${data.id}`" /></p>
    </div>

    <div class="lm-form-body">
      <div class="type-status-row">
        <div class="lm-field type-field">
          <label class="lm-label">{{ $t('globals.fields.type') }}</label>
          <div class="radio-group">
            <div class="check-row">
              <PvRadioButton v-model="form.type" name="type" value="user"
                input-id="type-user" :disabled="isEditing" />
              <label for="type-user" class="radio-label">
                <i class="pi pi-user" /> {{ $t('users.type.user') }}
              </label>
            </div>
            <div class="check-row">
              <PvRadioButton v-model="form.type" name="type" value="api"
                input-id="type-api" :disabled="isEditing" />
              <label for="type-api" class="radio-label">
                <i class="pi pi-code" /> {{ $t('users.type.api') }}
              </label>
            </div>
          </div>
        </div>
        <div class="lm-field status-field">
          <label class="lm-label">{{ $t('globals.fields.status') }}</label>
          <PvSelect v-model="form.status" name="status" required
            :options="[{ label: $t('users.status.enabled'), value: 'enabled' }, { label: $t('users.status.disabled'), value: 'disabled' }]"
            option-label="label" option-value="value" class="w-full" />
        </div>
      </div>

      <div class="lm-field">
        <label for="username" class="lm-label">{{ $t('users.username') }}</label>
        <PvInputText id="username" :maxlength="200" v-model="form.username" name="username" ref="focusEl"
          :placeholder="$t('users.username')" required autocomplete="off"
          pattern="[a-zA-Z0-9_\-\.@]+$" class="w-full" />
        <small class="lm-help">{{ $t('users.usernameHelp') }}</small>
      </div>

      <div class="lm-field">
        <label class="lm-label">{{ $t('globals.fields.name') }}</label>
        <PvInputText :maxlength="200" v-model="form.name" name="name"
          :placeholder="$t('globals.fields.name')" class="w-full" />
      </div>

      <div v-if="form.type !== 'api'" class="lm-field">
        <label class="lm-label">{{ $t('subscribers.email') }}</label>
        <PvInputText :maxlength="200" v-model="form.email" name="email" type="email"
          :placeholder="$t('subscribers.email')" required class="w-full" />
      </div>

      <div v-if="form.type !== 'api'" class="form-section">
        <div class="check-row">
          <PvCheckbox v-model="form.passwordLogin" :binary="true" input-id="passwordLogin" />
          <label for="passwordLogin" class="check-label">{{ $t('users.passwordEnable') }}</label>
        </div>
        <div v-if="form.passwordLogin" class="lm-field-row">
          <div class="lm-field">
            <label class="lm-label">{{ $t('users.password') }}</label>
            <PvPassword v-model="form.password" name="password"
              :placeholder="$t('users.password')" :minlength="8" :maxlength="200"
              :required="!isEditing" :feedback="false" class="w-full" />
          </div>
          <div class="lm-field">
            <label class="lm-label">{{ $t('users.passwordRepeat') }}</label>
            <PvPassword v-model="form.password2" name="password2"
              :required="!isEditing && !!form.password" :feedback="false" class="w-full" />
          </div>
        </div>
      </div>

      <p class="form-section-label">{{ $tc('users.roles') }}</p>
      <div class="lm-field-row">
        <div class="lm-field">
          <label class="lm-label">{{ $tc('users.userRole') }}</label>
          <PvSelect v-model="form.userRoleId" name="user_role" required
            :options="userRoles" option-label="name" option-value="id" class="w-full" />
        </div>
        <div class="lm-field">
          <label class="lm-label">{{ $tc('users.listRole', 0) }}</label>
          <PvSelect v-model="form.listRoleId" name="list_role"
            :options="listRoleOptions" option-label="name" option-value="id" class="w-full" />
        </div>
      </div>

      <div v-if="apiToken" class="user-api-token">
        <p class="api-token-label">{{ $t('users.apiOneTimeToken') }}</p>
        <copy-text :text="apiToken" />
      </div>
    </div>

    <div class="lm-form-footer">
      <PvButton @click="$emit('close')" :label="$t('globals.buttons.close')" severity="secondary" />
      <PvButton v-if="$can('users:manage') && !apiToken" type="submit" severity="primary"
        :loading="loading.users" data-cy="btn-save" :label="$t('globals.buttons.save')" />
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

const props = withDefaults(defineProps<{
  data?: any;
  isEditing?: boolean;
}>(), { data: () => ({}), isEditing: false });

const emit = defineEmits(['finished', 'close']);

const { $api, $utils } = useGlobal();
const { t } = useI18n();
const { loading, userRoles, listRoles } = storeToRefs(useMainStore());

const focusEl = ref<any>(null);
const apiToken = ref<string | null>(null);
const form = reactive<any>({
  username: '',
  email: '',
  name: '',
  password: '',
  password2: '',
  passwordLogin: false,
  type: 'user',
  status: 'enabled',
});

const listRoleOptions = computed(() => [
  { name: `— ${t('globals.terms.none')} —`, id: '' },
  ...(listRoles.value as any[]),
]);

function createUser() {
  const payload = {
    ...form, password_login: form.passwordLogin, user_role_id: form.userRoleId, list_role_id: form.listRoleId || null,
  };
  $api.createUser(payload).then((data: any) => {
    emit('finished');
    $utils.toast(t('globals.messages.created', { name: data.name }));
    if (payload.type === 'api') { apiToken.value = data.password; return; }
    emit('close');
  });
}

function updateUser() {
  const payload = {
    ...form, password_login: form.passwordLogin, user_role_id: form.userRoleId, list_role_id: form.listRoleId || null,
  };
  $api.updateUser({ id: props.data.id, ...payload }).then((data: any) => {
    emit('finished'); emit('close');
    $utils.toast(t('globals.messages.updated', { name: data.name }));
  });
}

function onSubmit() {
  if (!form.passwordLogin) { form.password = null; form.password2 = null; }
  if (form.type !== 'api' && form.passwordLogin && form.password && form.password !== form.password2) {
    $utils.toast(t('users.passwordMismatch'), 'is-danger');
    return;
  }
  if (props.isEditing) { updateUser(); } else { createUser(); }
}

onMounted(() => {
  Object.assign(form, props.data);
  if (props.data.userRole) form.userRoleId = props.data.userRole.id;
  form.listRoleId = props.data.listRole ? props.data.listRole.id : '';
  $api.getUserRoles();
  $api.getListRoles();
  nextTick(() => { focusEl.value?.$el?.focus(); });
});
</script>

<style scoped lang="scss">
.lm-field { display: flex; flex-direction: column; gap: 0.35rem; margin-bottom: 0; }
.lm-field-row { display: grid; grid-template-columns: 1fr 1fr; gap: 1rem; }
.lm-label { display: block; font-size: 0.8rem; font-weight: 600; color: var(--lm-text); }
.lm-help { display: block; font-size: 0.75rem; color: var(--lm-text-subtle); line-height: 1.4; }

.check-row { display: flex; align-items: center; gap: 0.5rem; }
.check-label { font-size: 0.875rem; color: var(--lm-text); cursor: pointer; font-weight: 500; }
.radio-label { cursor: pointer; display: inline-flex; align-items: center; gap: 0.35rem; font-size: 0.875rem; color: var(--lm-text); }

.type-status-row { display: flex; align-items: flex-end; gap: 1.5rem; }
.type-field { flex: 1; }
.status-field { flex: 0 0 180px; }
.radio-group { display: flex; gap: 1.25rem; align-items: center; height: 2.375rem; }

.form-section {
  border: 1px solid var(--lm-border);
  border-radius: 8px;
  padding: 1rem 1.25rem;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.form-section-label {
  font-size: 0.78rem;
  font-weight: 600;
  color: var(--lm-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.04em;
  margin: 0.25rem 0 -0.25rem;
}

:deep(.p-password) { width: 100%; }
:deep(.p-password-input) { width: 100%; }

.user-api-token {
  background: var(--lm-success-bg);
  border: 1px solid var(--lm-success-border);
  border-radius: 8px;
  padding: 1rem;
  font-size: 0.85rem;
  color: var(--lm-success-text, #166534);
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}
.api-token-label { font-weight: 600; margin: 0; }
</style>
