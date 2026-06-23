<template>
  <form class="lm-form" @submit.prevent="onSubmit">
    <div class="lm-form-header">
      <h3 class="lm-form-title">{{ $t('subscribers.manageLists') }}</h3>
    </div>

    <div class="lm-form-body">
      <div class="lm-field">
        <label class="lm-label">Action</label>
        <div class="radio-group">
          <label class="radio-item">
            <PvRadioButton v-model="form.action" name="action" value="add" input-id="action-add" data-cy="check-list-add" />
            <span>{{ $t('globals.buttons.add') }}</span>
          </label>
          <label class="radio-item">
            <PvRadioButton v-model="form.action" name="action" value="remove" input-id="action-remove" data-cy="check-list-remove" />
            <span>{{ $t('globals.buttons.remove') }}</span>
          </label>
          <label class="radio-item">
            <PvRadioButton v-model="form.action" name="action" value="unsubscribe" input-id="action-unsubscribe" data-cy="check-list-unsubscribe" />
            <span>{{ $t('subscribers.markUnsubscribed') }}</span>
          </label>
        </div>
      </div>

      <list-selector label="Target lists" placeholder="Lists to apply to" v-model="form.lists" :selected="form.lists"
        :all="lists.results" />

      <div class="lm-field">
        <div class="flex items-center gap-2">
          <PvCheckbox v-model="form.preconfirm" data-cy="preconfirm" :binary="true" :true-value="true" :false-value="false" :disabled="!hasOptinList" input-id="preconfirm" />
          <label for="preconfirm" class="lm-label" style="margin:0">{{ $t('subscribers.preconfirm') }}</label>
        </div>
        <small class="lm-help">{{ $t('subscribers.preconfirmHelp') }}</small>
      </div>
    </div>

    <div class="lm-form-footer">
      <PvButton @click="$emit('close')" :label="$t('globals.buttons.close')" severity="secondary" />
      <PvButton type="submit" severity="primary" :disabled="form.lists.length === 0" :label="$t('globals.buttons.save')" />
    </div>
  </form>
</template>

<script>
import { mapState } from 'pinia';
import { useMainStore } from '../store';
import ListSelector from '../components/ListSelector.vue';

export default {
  components: {
    ListSelector,
  },

  props: {
    numSubscribers: { type: Number, default: 0 },
  },

  emits: ['finished', 'close'],

  data() {
    return {
      form: {
        action: 'add',
        lists: [],
        preconfirm: false,
      },
    };
  },

  methods: {
    onSubmit() {
      this.$emit('finished', this.form.action, this.form.preconfirm, this.form.lists);
      this.$emit('close');
    },
  },

  computed: {
    ...mapState(useMainStore, ['lists', 'loading']),

    hasOptinList() {
      return this.form.lists.some((l) => l.optin === 'double');
    },
  },
};
</script>

<style scoped lang="scss">
.lm-form { display: flex; flex-direction: column; }

.lm-form-header {
  padding: 1.5rem 1.5rem 1rem;
  border-bottom: 1px solid #e2e8f0;
}
.lm-form-title { font-size: 1.1rem; font-weight: 700; color: #0f172a; margin: 0; }

.lm-form-body {
  padding: 1.25rem 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.lm-field { display: flex; flex-direction: column; gap: 0.35rem; }
.lm-label { font-size: 0.8rem; font-weight: 600; color: #374151; }
.lm-help { font-size: 0.75rem; color: #94a3b8; line-height: 1.4; }

.radio-group { display: flex; flex-direction: column; gap: 0.5rem; }
.radio-item { display: flex; align-items: center; gap: 0.5rem; cursor: pointer; font-size: 0.875rem; color: #374151; }

.lm-form-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
  padding: 1rem 1.5rem;
  border-top: 1px solid #e2e8f0;
}
</style>
