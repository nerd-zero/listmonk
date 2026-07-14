<template>
  <form class="lm-form" @submit.prevent="onSubmit">
    <!-- Dialog header -->
    <div class="lm-form-header">
      <div class="lm-form-title-row">
        <h3 class="lm-form-title">{{ isEditing ? data.name : $t('lists.newList') }}</h3>
        <PvTag v-if="isEditing" :severity="data.type === 'public' ? 'info' : 'secondary'"
          :value="$t(`lists.types.${data.type}`)" />
      </div>
      <p v-if="isEditing" class="lm-form-meta">
        ID: <copy-text :text="`${data.id}`" /> &nbsp;·&nbsp;
        UUID: <copy-text :text="data.uuid" />
      </p>
    </div>

    <!-- Fields -->
    <div class="lm-form-body">
      <div class="lm-field">
        <label class="lm-label">{{ $t('globals.fields.name') }}</label>
        <PvInputText :maxlength="200" ref="focus" v-model="form.name" name="name"
          :placeholder="$t('globals.fields.name')" class="w-full" required />
      </div>

      <div class="lm-field-row">
        <div class="lm-field">
          <label class="lm-label">{{ $t('lists.type') }}</label>
          <PvSelect v-model="form.type" name="type" required class="w-full"
            :options="[{ label: $t('lists.types.private'), value: 'private' }, { label: $t('lists.types.public'), value: 'public' }]"
            option-label="label" option-value="value" />
          <small class="lm-help">{{ $t('lists.typeHelp') }}</small>
        </div>

        <div class="lm-field">
          <label class="lm-label">{{ $t('lists.optin') }}</label>
          <PvSelect v-model="form.optin" name="optin" required class="w-full"
            :options="[{ label: $t('lists.optins.single'), value: 'single' }, { label: $t('lists.optins.double'), value: 'double' }]"
            option-label="label" option-value="value" />
          <small class="lm-help">{{ $t('lists.optinHelp') }}</small>
        </div>
      </div>

      <div class="lm-field">
        <label class="lm-label">{{ $t('globals.terms.tags') }}</label>
        <PvAutoComplete v-model="form.tags" name="tags" :typeahead="false"
          :placeholder="$t('globals.terms.tags')" multiple class="w-full" />
      </div>

      <div class="lm-field">
        <label class="lm-label">{{ $t('globals.fields.description') }}</label>
        <PvTextarea :maxlength="2000" v-model="form.description" name="description"
          :placeholder="$t('globals.fields.description')" class="w-full" rows="3" />
      </div>

      <div class="lm-field lm-field--inline">
        <div>
          <label class="lm-label">{{ $t('lists.archived') }}</label>
          <small class="lm-help">{{ $t('lists.archivedHelp') }}</small>
        </div>
        <PvToggleSwitch v-model="isArchived" name="status" />
      </div>
    </div>

    <!-- Footer -->
    <div class="lm-form-footer">
      <PvButton severity="secondary" :label="$t('globals.buttons.close')" @click="$emit('close')" />
      <PvButton
        v-if="$can('lists:manage_all') || $canList(data.id, 'list:manage')"
        type="submit"
        severity="primary"
        :loading="loading.lists"
        data-cy="btn-save"
        :label="$t('globals.buttons.save')"
      />
    </div>
  </form>
</template>

<script setup lang="ts">
import {
  ref, reactive, computed, onMounted, nextTick,
} from 'vue';
import { storeToRefs } from 'pinia';
import { useI18n } from 'vue-i18n';
import { useMainStore } from '../store';
import { useGlobal } from '../composables/useGlobal';
import CopyText from '../components/CopyText.vue';
import { getLists as listsApi } from '../api/generated/endpoints/lists/lists';

const props = withDefaults(defineProps<{
  data?: any;
  isEditing?: boolean;
}>(), { data: () => ({}), isEditing: false });

const emit = defineEmits(['finished', 'close']);

const { $utils } = useGlobal();
const { createList, updateList } = listsApi();
const { t } = useI18n();
const { loading } = storeToRefs(useMainStore());

const focusEl = ref<any>(null);

const form = reactive({
  name: '',
  type: 'private',
  optin: 'single',
  status: 'active',
  tags: [] as string[],
});

const isArchived = computed({
  get: () => form.status === 'archived',
  set: (v: boolean) => { form.status = v ? 'archived' : 'active'; },
});

function onCreateList() {
  createList(form as any).then((data: any) => {
    emit('finished');
    emit('close');
    $utils.toast(t('globals.messages.created', { name: data.name }));
  });
}

function onUpdateList() {
  updateList(props.data.id, form as any).then((data: any) => {
    emit('finished');
    emit('close');
    $utils.toast(t('globals.messages.updated', { name: data.name }));
  });
}

function onSubmit() {
  if (props.isEditing) { onUpdateList(); return; }
  onCreateList();
}

onMounted(() => {
  Object.assign(form, props.data);
  nextTick(() => { focusEl.value?.$el?.focus(); });
});
</script>

<style scoped lang="scss">
:deep(.p-tag-secondary) {
  background: var(--lm-bg-subtle);
  color: var(--lm-text-secondary);
  border: 1px solid var(--lm-border);
}

.lm-field { display: flex; flex-direction: column; gap: 0.35rem; }
.lm-field-row { display: grid; grid-template-columns: 1fr 1fr; gap: 1rem; }
.lm-field--inline {
  flex-direction: row;
  align-items: flex-start;
  justify-content: space-between;
  gap: 1rem;
  padding: 0.75rem 1rem;
  background: var(--lm-bg);
  border: 1px solid var(--lm-border);
  border-radius: 8px;

  :deep(.p-toggleswitch) { flex-shrink: 0; }
}

.lm-label {
  display: block;
  font-size: 0.8rem;
  font-weight: 600;
  color: var(--lm-text);
}
.lm-help {
  display: block;
  font-size: 0.75rem;
  color: var(--lm-text-subtle);
  line-height: 1.4;
}

</style>
