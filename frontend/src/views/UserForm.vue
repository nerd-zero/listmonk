<template>
  <form class="lm-form" @submit.prevent="onSubmit">
    <div class="lm-form-header">
      <div class="lm-form-title-row">
        <h3 class="lm-form-title">{{ isEditing ? data.name : $t('users.newUser') }}</h3>
      </div>
      <p v-if="isEditing" class="lm-form-meta">{{ $t('globals.fields.id') }}: <copy-text :text="`${data.id}`" /></p>
    </div>
    <div class="lm-form-body">
        <div class="grid">
          <div class="col-6">
            <div class="field">
              <div class="flex gap-2">
                <label class="flex items-center gap-2 cursor-pointer">
                  <input type="radio" v-model="form.type" name="type" value="user" :disabled="isEditing" />
                  <i class="pi pi-user" />
                  {{ $t('users.type.user') }}
                </label>
                <label class="flex items-center gap-2 cursor-pointer">
                  <input type="radio" v-model="form.type" name="type" value="api" :disabled="isEditing" />
                  <i class="pi pi-code" />
                  {{ $t('users.type.api') }}
                </label>
              </div>
            </div>
          </div>
          <div class="col-6">
            <div class="field">
              <label class="block mb-1 text-sm font-medium">{{ $t('globals.fields.status') }}</label>
              <PvSelect v-model="form.status" name="status" required
                :options="[{ label: $t('users.status.enabled'), value: 'enabled' }, { label: $t('users.status.disabled'), value: 'disabled' }]"
                option-label="label" option-value="value" class="w-full" />
            </div>
          </div>
        </div>

        <div class="field">
          <label class="block mb-1 text-sm font-medium">{{ $t('users.username') }}</label>
          <PvInputText :maxlength="200" v-model="form.username" name="username" ref="focus" autofocus
            :placeholder="$t('users.username')" required autocomplete="off"
            pattern="[a-zA-Z0-9_\-\.@]+$" class="w-full" />
          <small class="block mt-1 text-color-secondary">{{ $t('users.usernameHelp') }}</small>
        </div>

        <div class="field">
          <label class="block mb-1 text-sm font-medium">{{ $t('globals.fields.name') }}</label>
          <PvInputText :maxlength="200" v-model="form.name" name="name" :placeholder="$t('globals.fields.name')" class="w-full" />
        </div>

        <div v-if="form.type !== 'api'" class="field">
          <label class="block mb-1 text-sm font-medium">{{ $t('subscribers.email') }}</label>
          <PvInputText :maxlength="200" v-model="form.email" name="email" type="email"
            :placeholder="$t('subscribers.email')" required class="w-full" />
        </div>

        <template v-if="form.type !== 'api'">
          <div class="form-section">
            <div class="field">
              <div class="flex items-center gap-2">
                <PvCheckbox v-model="form.passwordLogin" :binary="true" name="password_login" input-id="passwordLogin" />
                <label for="passwordLogin" class="cursor-pointer">{{ $t('users.passwordEnable') }}</label>
              </div>
            </div>

            <div class="grid">
              <div class="col-6">
                <div class="field">
                  <label class="block mb-1 text-sm font-medium">{{ $t('users.password') }}</label>
                  <PvPassword :disabled="!form.passwordLogin" :minlength="8" :maxlength="200" v-model="form.password"
                    name="password" :placeholder="$t('users.password')"
                    :required="form.passwordLogin && !isEditing" :feedback="false" class="w-full" />
                </div>
              </div>
              <div class="col-6">
                <div class="field">
                  <label class="block mb-1 text-sm font-medium">{{ $t('users.passwordRepeat') }}</label>
                  <PvPassword :disabled="!form.passwordLogin" :minlength="8" :maxlength="200" v-model="form.password2"
                    name="password2" :required="form.passwordLogin && !isEditing && form.password" :feedback="false" class="w-full" />
                </div>
              </div>
            </div>
          </div>
        </template>

        <p class="form-section-label">{{ $tc('users.roles') }}</p>
        <div class="form-section">
          <div class="grid">
            <div class="col-6">
              <div class="field">
                <label class="block mb-1 text-sm font-medium">{{ $tc('users.userRole') }}</label>
                <PvSelect v-model="form.userRoleId" name="user_role" required
                  :options="userRoles" option-label="name" option-value="id" class="w-full" />
              </div>
            </div>

            <div class="col-6">
              <div class="field">
                <label class="block mb-1 text-sm font-medium">{{ $tc('users.listRole', 0) }}</label>
                <PvSelect v-model="form.listRoleId" name="list_role"
                  :options="listRoleOptions" option-label="name" option-value="id" class="w-full" />
              </div>
            </div>
          </div>
        </div>

        <div v-if="apiToken" class="user-api-token">
          <p>{{ $t('users.apiOneTimeToken') }}</p>
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

<script>
import { mapState } from 'pinia';
import { useMainStore } from '../store';
import CopyText from '../components/CopyText.vue';

export default {
  name: 'UserForm',

  components: {
    CopyText,
  },

  props: {
    data: { type: Object, default: () => ({}) },
    isEditing: { type: Boolean, default: false },
  },

  emits: ['finished', 'close'],

  data() {
    return {
      // Binds form input values.
      form: {
        username: '',
        email: '',
        name: '',
        password: '',
        passwordLogin: false,
        type: 'user',
        status: 'enabled',
      },
      apiToken: null,
    };
  },

  methods: {
    onSubmit() {
      if (!this.form.passwordLogin) {
        this.form.password = null;
        this.form.password2 = null;
      }

      if (this.isEditing) {
        if (this.form.type !== 'api' && this.form.passwordLogin && this.form.password && this.form.password !== this.form.password2) {
          this.$utils.toast(this.$t('users.passwordMismatch'), 'is-danger');
          return;
        }

        this.updateUser();
        return;
      }

      if (this.form.type !== 'api' && this.form.passwordLogin && this.form.password !== this.form.password2) {
        this.$utils.toast(this.$t('users.passwordMismatch'), 'is-danger');
        return;
      }

      this.createUser();
    },

    createUser() {
      const form = {
        ...this.form, password_login: this.form.passwordLogin, user_role_id: this.form.userRoleId, list_role_id: this.form.listRoleId || null,
      };
      this.$api.createUser(form).then((data) => {
        this.$emit('finished');
        this.$utils.toast(this.$t('globals.messages.created', { name: data.name }));

        // If the user is an API user, show the one-time token.
        if (form.type === 'api') {
          this.apiToken = data.password;
          return;
        }

        this.$emit('finished');
        this.$emit('close');
      });
    },

    updateUser() {
      const form = {
        ...this.form, password_login: this.form.passwordLogin, user_role_id: this.form.userRoleId, list_role_id: this.form.listRoleId || null,
      };
      this.$api.updateUser({ id: this.data.id, ...form }).then((data) => {
        this.$emit('finished');
        this.$emit('close');
        this.$utils.toast(this.$t('globals.messages.updated', { name: data.name }));
      });
    },

    hasType(t) {
      // If the user being edited is API, then the only valid field is API.
      // Otherwise, all fields are valid except API.
      return !this.$props.isEditing || (this.form.type === 'api' ? t === 'api' : t !== 'api');
    },
  },

  computed: {
    ...mapState(useMainStore, ['loading', 'userRoles', 'listRoles']),

    listRoleOptions() {
      return [{ name: `— ${this.$t('globals.terms.none')} —`, id: '' }, ...this.listRoles];
    },
  },

  mounted() {
    this.form = { ...this.form, ...this.$props.data };
    if (this.$props.data.userRole) {
      this.form.userRoleId = this.$props.data.userRole.id;
    }

    this.form.listRoleId = this.$props.data.listRole ? this.$props.data.listRole.id : '';

    this.$api.getUserRoles();
    this.$api.getListRoles();

    this.$nextTick(() => {
      this.$refs.focus.$el.focus();
    });
  },
};
</script>

<style scoped lang="scss">
.form-section {
  border: 1px solid var(--lm-border);
  border-radius: 8px;
  padding: 1rem 1.25rem;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.form-section-label {
  font-size: 0.8rem;
  font-weight: 600;
  color: var(--lm-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.04em;
  margin: 0.25rem 0 -0.25rem;
}

.user-api-token {
  background: var(--lm-success-bg);
  border: 1px solid var(--lm-success-border);
  border-radius: 8px;
  padding: 1rem;
  font-size: 0.85rem;
  color: #166534;
}
</style>
