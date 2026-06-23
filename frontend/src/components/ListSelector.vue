<template>
  <div class="field list-selector">
    <div :class="['list-tags', ...classes]">
      <div class="tags">
        <PvTag v-for="l in selectedItems" :key="l.id" :class="[l.subscriptionStatus, { 'is-restricted': l.restricted }, 'list']"
          :data-id="l.id" class="list">
          {{ l.name }}
          <sup v-if="l.optin === 'double' && l.subscriptionStatus">
            {{ $t(`subscribers.status.${l.subscriptionStatus}`) }}
          </sup>
          <i v-if="!$props.disabled && !l.restricted" class="pi pi-times ml-1 cursor-pointer" @click="removeList(l.id)" />
        </PvTag>
      </div>
    </div>

    <div class="field">
      <label class="block mb-1 text-sm font-medium">{{ label + (selectedItems ? ` (${selectedItems.length})` : '') }}</label>
      <PvAutoComplete v-model="query" :placeholder="placeholder"
        :disabled="all.length === 0 || $props.disabled"
        :suggestions="filteredLists" @item-select="onSelect" option-label="name"
        :dropdown="true" force-selection />
      <small v-if="message" class="block mt-1 text-color-secondary">{{ message }}</small>
    </div>
  </div>
</template>

<script>

export default {
  name: 'ListSelector',

  props: {
    label: { type: String, default: '' },
    placeholder: { type: String, default: '' },
    message: { type: String, default: '' },
    required: Boolean,
    disabled: Boolean,
    classes: {
      type: Array,
      default: () => [],
    },
    selected: {
      type: Array,
      default: () => [],
    },
    all: {
      type: Array,
      default: () => [],
    },
  },

  data() {
    return {
      query: '',
      selectedItems: [],
    };
  },

  methods: {
    onSelect(event) {
      this.selectList(event.value);
    },

    selectList(l) {
      if (!l) {
        return;
      }
      this.selectedItems.push(l);
      this.query = '';

      // Propagate the items to the parent's v-model binding.
      this.$nextTick(() => {
        this.$emit('input', this.selectedItems);
      });
    },

    removeList(id) {
      this.selectedItems = this.selectedItems.filter((l) => l.id !== id);

      // Propagate the items to the parent's v-model binding.
      this.$nextTick(() => {
        this.$emit('input', this.selectedItems);
      });
    },
  },

  computed: {
    // Return the list of unselected lists.
    filteredLists() {
      // Get a map of IDs of the user subscriptions. eg: {1: true, 2: true};
      const subIDs = this.selectedItems.reduce((obj, item) => ({ ...obj, [item.id]: true }), {});

      // Filter lists from the global lists whose IDs are not in the user's
      // subscribed ist.
      const q = typeof this.query === 'string' ? this.query.toLowerCase() : '';
      return this.$props.all.filter(
        (l) => (!(l.id in subIDs) && l.name.toLowerCase().indexOf(q) >= 0),
      );
    },
  },

  watch: {
    // This is required to update the array of lists to propagate from parent
    // components and "react" on the selector.
    selected() {
      // Deep-copy.
      this.selectedItems = JSON.parse(JSON.stringify(this.selected));
    },
  },

  mounted() {
    if (this.selected) {
      this.selectedItems = JSON.parse(JSON.stringify(this.selected));
    }
  },
};
</script>
