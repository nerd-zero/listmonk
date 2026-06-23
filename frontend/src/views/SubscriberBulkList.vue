<template>
  <form @submit.prevent="onSubmit">
    <div class="modal-card" style="width: auto">
      <header class="modal-card-head">
        <h4 class="title is-size-5">
          {{ $t('subscribers.manageLists') }}
        </h4>
      </header>

      <section expanded class="modal-card-body">
        <div class="field">
          <label class="block mb-1 text-sm font-medium">Action</label>
          <div>
            <div class="flex items-center gap-2 mb-1">
              <PvRadioButton v-model="form.action" name="action" value="add" input-id="action-add" data-cy="check-list-add" />
              <label for="action-add">{{ $t('globals.buttons.add') }}</label>
            </div>
            <div class="flex items-center gap-2 mb-1">
              <PvRadioButton v-model="form.action" name="action" value="remove" input-id="action-remove" data-cy="check-list-remove" />
              <label for="action-remove">{{ $t('globals.buttons.remove') }}</label>
            </div>
            <div class="flex items-center gap-2 mb-1">
              <PvRadioButton v-model="form.action" name="action" value="unsubscribe" input-id="action-unsubscribe" data-cy="check-list-unsubscribe" />
              <label for="action-unsubscribe">{{ $t('subscribers.markUnsubscribed') }}</label>
            </div>
          </div>
        </div>

        <list-selector label="Target lists" placeholder="Lists to apply to" v-model="form.lists" :selected="form.lists"
          :all="lists.results" />

        <div class="field">
          <div class="flex items-center gap-2">
            <PvCheckbox v-model="form.preconfirm" data-cy="preconfirm" :binary="true" :true-value="true" :false-value="false" :disabled="!hasOptinList" input-id="preconfirm" />
            <label for="preconfirm">{{ $t('subscribers.preconfirm') }}</label>
          </div>
          <small class="block mt-1 text-color-secondary">{{ $t('subscribers.preconfirmHelp') }}</small>
        </div>
      </section>

      <footer class="modal-card-foot has-text-right">
        <PvButton @click="$parent.close()" :label="$t('globals.buttons.close')" />
        <PvButton type="submit" severity="primary" :disabled="form.lists.length === 0" :label="$t('globals.buttons.save')" />
      </footer>
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

  data() {
    return {
      // Binds form input values.
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
      this.$parent.close();
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
