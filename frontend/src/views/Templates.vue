<template>
  <div class="templates-page">
    <div class="page-header">
      <h1 class="page-title">
        {{ $t('globals.terms.templates') }}
        <span v-if="templates.length > 0" class="page-title-count">{{ templates.length }}</span>
      </h1>
      <PvButton v-if="$can('templates:manage')" severity="primary" icon="pi pi-plus"
        @click="showNewForm" :label="$t('globals.buttons.new')" />
    </div>

    <div class="table-card">
      <PvDataTable :value="templates" :loading="loading.templates" sort-field="createdAt" sort-order="1">
        <PvColumn field="name" :header="$t('globals.fields.name')" sortable>
          <template #body="{ data }">
            <div class="name-cell">
              <div class="name-row">
                <a href="#" class="row-name" @click.prevent="showEditForm(data)">{{ data.name }}</a>
                <PvTag v-if="data.isDefault" severity="success" size="small" :value="$t('templates.default')" />
              </div>
              <span v-if="data.type === 'tx'" class="subject-text">{{ data.subject }}</span>
            </div>
          </template>
        </PvColumn>

        <PvColumn field="type" :header="$t('globals.fields.type')" sortable style="width:14rem">
          <template #body="{ data }">
            <PvTag v-if="data.type === 'campaign'" severity="info" size="small" :data-cy="`type-${data.type}`"
              :value="$tc('templates.typeCampaignHTML')" />
            <PvTag v-else-if="data.type === 'campaign_visual'" severity="warn" size="small"
              :data-cy="`type-${data.type}`" :value="$tc('templates.typeCampaignVisual')" />
            <PvTag v-else severity="secondary" size="small" :data-cy="`type-${data.type}`"
              :value="$tc('templates.typeTransactional')" />
          </template>
        </PvColumn>

        <PvColumn field="id" :header="$t('globals.fields.id')" sortable style="width:5rem" />

        <PvColumn field="createdAt" :header="$t('globals.fields.createdAt')" sortable style="width:10rem">
          <template #body="{ data }">{{ $utils.niceDate(data.createdAt) }}</template>
        </PvColumn>

        <PvColumn field="updatedAt" :header="$t('globals.fields.updatedAt')" sortable style="width:10rem">
          <template #body="{ data }">{{ $utils.niceDate(data.updatedAt) }}</template>
        </PvColumn>

        <PvColumn style="width:9rem; text-align:right">
          <template #body="{ data }">
            <div class="row-actions">
              <button type="button" class="row-action-btn" data-cy="btn-preview"
                v-tooltip.bottom="$t('templates.preview')" @click="previewTemplate(data)">
                <i class="pi pi-file" />
              </button>
              <button type="button" class="row-action-btn" data-cy="btn-edit"
                v-tooltip.bottom="$t('globals.buttons.edit')" @click="showEditForm(data)">
                <i class="pi pi-pencil" />
              </button>
              <button type="button" class="row-action-btn" data-cy="btn-clone"
                v-tooltip.bottom="$t('globals.buttons.clone')"
                @click="$utils.prompt('Clone template', { placeholder: 'Name', value: `Copy of ${data.name}` }, (name) => cloneTemplate(name, data))">
                <i class="pi pi-copy" />
              </button>
              <button v-if="!data.isDefault && data.type === 'campaign'" type="button" class="row-action-btn"
                data-cy="btn-set-default" v-tooltip.bottom="$t('templates.makeDefault')"
                @click="$utils.confirm(null, () => makeTemplateDefault(data))">
                <i class="pi pi-check-circle" />
              </button>
              <button v-if="!data.isDefault" type="button" class="row-action-btn row-action-btn--danger"
                data-cy="btn-delete" v-tooltip.bottom="$t('globals.buttons.delete')"
                @click="$utils.confirm(null, () => deleteTemplate(data))">
                <i class="pi pi-trash" />
              </button>
            </div>
          </template>
        </PvColumn>

        <template #empty v-if="!loading.templates">
          <empty-placeholder />
        </template>
      </PvDataTable>
    </div>

    <!-- Add / edit form modal -->
    <PvDialog v-model:visible="isFormVisible" :style="{ width: '1200px' }" show-header="false" :closable="false" modal
      class="template-modal">
      <template-form :data="curItem" :is-editing="isEditing" @finished="formFinished" @close="isFormVisible = false" />
    </PvDialog>

    <campaign-preview v-if="previewItem" type="template" :id="previewItem.id" :template-type="previewItem.type"
      :title="previewItem.name" @close="closePreview" />
  </div>
</template>

<script>
import { mapState } from 'pinia';
import { useMainStore } from '../store';
import CampaignPreview from '../components/CampaignPreview.vue';
import EmptyPlaceholder from '../components/EmptyPlaceholder.vue';

import TemplateForm from './TemplateForm.vue';

export default {
  components: {
    CampaignPreview,
    TemplateForm,
    EmptyPlaceholder,
  },

  data() {
    return {
      curItem: null,
      isEditing: false,
      isFormVisible: false,
      previewItem: null,
    };
  },

  methods: {
    fetchTemplates() {
      this.$api.getTemplates();
    },

    // Show the edit form.
    showEditForm(data) {
      this.curItem = data;
      this.isFormVisible = true;
      this.isEditing = true;
    },

    // Show the new form.
    showNewForm() {
      this.curItem = { type: 'campaign' };
      this.isFormVisible = true;
      this.isEditing = false;
    },

    formFinished() {
      this.$api.getTemplates();
    },

    previewTemplate(c) {
      this.previewItem = c;
    },

    closePreview() {
      this.previewItem = null;
    },

    cloneTemplate(name, t) {
      const data = {
        name,
        type: t.type,
        subject: t.subject,
        body: t.body,
        body_source: t.bodySource,
      };
      this.$api.createTemplate(data).then((d) => {
        this.$api.getTemplates();
        this.$emit('finished');
        this.$utils.toast(`'${d.name}' created`);
      });
    },

    makeTemplateDefault(tpl) {
      this.$api.makeTemplateDefault(tpl.id).then(() => {
        this.$api.getTemplates();
        this.$utils.toast(this.$t('globals.messages.created', { name: tpl.name }));
      });
    },

    deleteTemplate(tpl) {
      this.$api.deleteTemplate(tpl.id).then(() => {
        this.$api.getTemplates();
        this.$utils.toast(this.$t('globals.messages.deleted', { name: tpl.name }));
      });
    },
  },

  computed: {
    ...mapState(useMainStore, ['refreshTick', 'templates', 'loading']),
  },

  watch: {
    refreshTick() { this.fetchTemplates(); },
  },

  mounted() {
    this.$api.getTemplates();
  },
};
</script>

<style scoped lang="scss">
.templates-page { display: flex; flex-direction: column; gap: 1.5rem; }

:deep(.p-tag-secondary) {
  background: var(--lm-bg-subtle);
  color: var(--lm-text-secondary);
  border: 1px solid var(--lm-border);
}

.name-cell { display: flex; flex-direction: column; gap: 0.2rem; }
.name-row { display: flex; align-items: center; gap: 0.5rem; }
.row-name { color: var(--lm-text); font-weight: 500; text-decoration: none; &:hover { color: var(--lm-primary); } }
.subject-text { font-size: 0.78rem; color: var(--lm-text-subtle); }
</style>
