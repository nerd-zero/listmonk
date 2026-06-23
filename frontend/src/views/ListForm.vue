<template>
  <form @submit.prevent="onSubmit">
    <div class="modal-card content" style="width: auto">
      <header class="modal-card-head">
        <p v-if="isEditing" class="has-text-grey-light is-size-7">
          {{ $t('globals.fields.id') }}: <copy-text :text="`${data.id}`" />
          {{ $t('globals.fields.uuid') }}: <copy-text :text="data.uuid" />
        </p>
        <PvTag v-if="isEditing" :class="[data.type, 'is-pulled-right']"
          :value="$t(`lists.types.${data.type}`)" />
        <h4 v-if="isEditing">
          {{ data.name }}
        </h4>
        <h4 v-else>
          {{ $t('lists.newList') }}
        </h4>
      </header>
      <section expanded class="modal-card-body">
        <div class="field">
          <label class="block mb-1 text-sm font-medium">{{ $t('globals.fields.name') }}</label>
          <PvInputText :maxlength="200" ref="focus" v-model="form.name" name="name"
            :placeholder="$t('globals.fields.name')" required />
        </div>

        <div class="field">
          <label class="block mb-1 text-sm font-medium">{{ $t('lists.type') }}</label>
          <PvSelect v-model="form.type" name="type" :placeholder="$t('lists.typeHelp')" required
            :options="[{ label: $t('lists.types.private'), value: 'private' }, { label: $t('lists.types.public'), value: 'public' }]"
            option-label="label" option-value="value" />
          <small class="block mt-1 text-color-secondary">{{ $t('lists.typeHelp') }}</small>
        </div>

        <div class="field">
          <label class="block mb-1 text-sm font-medium">{{ $t('lists.optin') }}</label>
          <PvSelect v-model="form.optin" name="optin" placeholder="Opt-in type" required
            :options="[{ label: $t('lists.optins.single'), value: 'single' }, { label: $t('lists.optins.double'), value: 'double' }]"
            option-label="label" option-value="value" />
          <small class="block mt-1 text-color-secondary">{{ $t('lists.optinHelp') }}</small>
        </div>

        <div class="field">
          <label class="block mb-1 text-sm font-medium">{{ $t('globals.terms.tags') }}</label>
          <PvAutoComplete v-model="form.tags" name="tags"
            :placeholder="$t('globals.terms.tags')" multiple />
        </div>

        <div class="field">
          <label class="block mb-1 text-sm font-medium">{{ $t('globals.fields.description') }}</label>
          <PvTextarea :maxlength="2000" v-model="form.description" name="description"
            :placeholder="$t('globals.fields.description')" />
        </div>

        <div class="field">
          <label class="block mb-1 text-sm font-medium">{{ $t('lists.archived') }}</label>
          <div class="flex items-center gap-2">
            <PvToggleSwitch v-model="isArchived" name="status" />
          </div>
          <small class="block mt-1 text-color-secondary">{{ $t('lists.archivedHelp') }}</small>
        </div>
      </section>
      <footer class="modal-card-foot has-text-right">
        <PvButton @click="$parent.close()" :label="$t('globals.buttons.close')" />
        <PvButton v-if="$can('lists:manage_all') || $canList(data.id, 'list:manage')" type="submit"
          severity="primary" :loading="loading.lists" data-cy="btn-save"
          :label="$t('globals.buttons.save')" />
      </footer>
    </div>
  </form>
</template>

<script>
import { mapState } from 'pinia';
import { useMainStore } from '../store';
import CopyText from '../components/CopyText.vue';

export default {
  name: 'ListForm',

  components: {
    CopyText,
  },

  props: {
    data: { type: Object, default: () => ({}) },
    isEditing: { type: Boolean, default: false },
  },

  data() {
    return {
      // Binds form input values.
      form: {
        name: '',
        type: 'private',
        optin: 'single',
        status: 'active',
        tags: [],
      },
    };
  },

  methods: {
    onSubmit() {
      if (this.isEditing) {
        this.updateList();
        return;
      }

      this.createList();
    },

    createList() {
      this.$api.createList(this.form).then((data) => {
        this.$emit('finished');
        this.$parent.close();
        this.$utils.toast(this.$t('globals.messages.created', { name: data.name }));
      });
    },

    updateList() {
      this.$api.updateList({ id: this.data.id, ...this.form }).then((data) => {
        this.$emit('finished');
        this.$parent.close();
        this.$utils.toast(this.$t('globals.messages.updated', { name: data.name }));
      });
    },
  },

  computed: {
    ...mapState(useMainStore, ['loading', 'profile']),

    isArchived: {
      get() {
        return this.form.status === 'archived';
      },
      set(v) {
        this.form.status = v ? 'archived' : 'active';
      },
    },
  },

  mounted() {
    this.form = { ...this.form, ...this.$props.data };

    this.$nextTick(() => {
      this.$refs.focus.$el.focus();
    });
  },
};
</script>
