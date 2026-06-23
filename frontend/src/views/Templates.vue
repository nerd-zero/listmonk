<template>
  <section class="templates">
    <header class="columns page-header">
      <div class="column is-10">
        <h1 class="title is-4">
          {{ $t('globals.terms.templates') }}
          <span v-if="templates.length > 0">({{ templates.length }})</span>
        </h1>
      </div>
      <div class="column has-text-right">
        <div v-if="$can('templates:manage')" class="field">
          <PvButton severity="primary" icon="pi pi-plus" class="btn-new" @click="showNewForm"
            :label="$t('globals.buttons.new')" />
        </div>
      </div>
    </header>

    <PvDataTable :value="templates" :loading="loading.templates" sort-field="createdAt" sort-order="1">
      <PvColumn field="name" :header="$t('globals.fields.name')" sortable>
        <template #body="{ data }">
          <a href="#" @click.prevent="showEditForm(data)">
            {{ data.name }}
          </a>
          <PvTag v-if="data.isDefault" :value="$t('templates.default')" />

          <p class="is-size-7 has-text-grey" v-if="data.type === 'tx'">
            {{ data.subject }}
          </p>
        </template>
      </PvColumn>

      <PvColumn field="type" :header="$t('globals.fields.type')" sortable>
        <template #body="{ data }">
          <PvTag v-if="data.type === 'campaign'" :class="data.type" :data-cy="`type-${data.type}`"
            :value="$tc('templates.typeCampaignHTML')" />
          <PvTag v-else-if="data.type === 'campaign_visual'" :class="data.type"
            :data-cy="`type-${data.type}`" :value="$tc('templates.typeCampaignVisual')" />
          <PvTag v-else :class="data.type" :data-cy="`type-${data.type}`"
            :value="$tc('templates.typeTransactional')" />
        </template>
      </PvColumn>

      <PvColumn field="id" :header="$t('globals.fields.id')" sortable />

      <PvColumn field="createdAt" :header="$t('globals.fields.createdAt')" sortable>
        <template #body="{ data }">
          {{ $utils.niceDate(data.createdAt) }}
        </template>
      </PvColumn>

      <PvColumn field="updatedAt" :header="$t('globals.fields.updatedAt')" sortable>
        <template #body="{ data }">
          {{ $utils.niceDate(data.updatedAt) }}
        </template>
      </PvColumn>

      <PvColumn class="actions" style="text-align:right">
        <template #body="{ data }">
          <div>
            <a href="#" @click.prevent="previewTemplate(data)" data-cy="btn-preview"
              :aria-label="$t('templates.preview')">
              <i class="pi pi-file" v-tooltip.bottom="$t('templates.preview')" />
            </a>
            <a href="#" @click.prevent="showEditForm(data)" data-cy="btn-edit"
              :aria-label="$t('globals.buttons.edit')">
              <i class="pi pi-pencil" v-tooltip.bottom="$t('globals.buttons.edit')" />
            </a>
            <a href="#" @click.prevent="$utils.prompt(`Clone template`,
              { placeholder: 'Name', value: `Copy of ${data.name}` },
              (name) => cloneTemplate(name, data))" data-cy="btn-clone" :aria-label="$t('globals.buttons.clone')">
              <i class="pi pi-copy" v-tooltip.bottom="$t('globals.buttons.clone')" />
            </a>
            <a v-if="!data.isDefault && data.type === 'campaign'" href="#"
              @click.prevent="$utils.confirm(null, () => makeTemplateDefault(data))" data-cy="btn-set-default"
              :aria-label="$t('templates.makeDefault')">
              <i class="pi pi-check-circle" v-tooltip.bottom="$t('templates.makeDefault')" />
            </a>
            <span v-else class="a has-text-grey-light">
              <i class="pi pi-check-circle" />
            </span>

            <a v-if="!data.isDefault" href="#" @click.prevent="$utils.confirm(null, () => deleteTemplate(data))"
              data-cy="btn-delete" :aria-label="$t('globals.buttons.delete')">
              <i class="pi pi-trash" v-tooltip.bottom="$t('globals.buttons.delete')" />
            </a>
            <span v-else class="a has-text-grey-light">
              <i class="pi pi-trash" />
            </span>
          </div>
        </template>
      </PvColumn>

      <template #empty v-if="!loading.templates">
        <empty-placeholder />
      </template>
    </PvDataTable>

    <!-- Add / edit form modal -->
    <PvDialog v-model:visible="isFormVisible" :style="{ width: '1200px' }" :closable="false" modal
      class="template-modal">
      <template-form :data="curItem" :is-editing="isEditing" @finished="formFinished" />
    </PvDialog>

    <campaign-preview v-if="previewItem" type="template" :id="previewItem.id" :template-type="previewItem.type"
      :title="previewItem.name" @close="closePreview" />
  </section>
</template>

<script>
import { mapState } from 'vuex';
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
    ...mapState(['templates', 'loading']),
  },

  created() {
    this.$root.$on('page.refresh', this.fetchTemplates);
  },

  unmounted() {
    this.$root.$off('page.refresh', this.fetchTemplates);
  },

  mounted() {
    this.$api.getTemplates();
  },
};
</script>
