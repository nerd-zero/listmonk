<template>
  <section>
    <form class="lm-form" @submit.prevent="onSubmit">
      <div class="lm-form-header">
        <div class="lm-form-title-row">
          <h3 class="lm-form-title">{{ isEditing ? data.name : $t('templates.newTemplate') }}</h3>
          <PvButton severity="secondary" @click="onTogglePreview" icon="pi pi-file" :label="$t('templates.preview') + ' (F9)'" />
        </div>
        <p v-if="isEditing" class="lm-form-meta">
          {{ $t('globals.fields.id') }}: <copy-text :text="`${data.id}`" data-cy="id" />
        </p>
      </div>
      <div class="lm-form-body">
          <div class="grid">
            <div class="col-9">
              <div class="field">
                <label class="block mb-1 text-sm font-medium">{{ $t('globals.fields.name') }}</label>
                <PvInputText :maxlength="200" ref="focus" v-model="form.name" name="name"
                  :placeholder="$t('globals.fields.name')" required class="w-full" />
              </div>
            </div>
            <div class="col-3">
              <div class="field">
                <label class="block mb-1 text-sm font-medium">{{ $t('globals.fields.type') }}</label>
                <PvSelect v-model="form.type" :disabled="isEditing"
                  :options="[
                    { label: $tc('templates.typeCampaignHTML'), value: 'campaign' },
                    { label: $tc('templates.typeCampaignVisual'), value: 'campaign_visual' },
                    { label: $tc('templates.typeTransactional'), value: 'tx' },
                  ]"
                  option-label="label" option-value="value" class="w-full" />
              </div>
            </div>
          </div>
          <div class="grid" v-if="form.type === 'tx'">
            <div class="col-12">
              <div class="field">
                <label class="block mb-1 text-sm font-medium">{{ $t('templates.subject') }}</label>
                <PvInputText :maxlength="200" v-model="form.subject" name="subject"
                  :placeholder="$t('templates.subject')" required class="w-full" />
              </div>
            </div>
          </div>

          <template v-if="form.body !== null">
            <div v-if="form.type === 'campaign_visual'" class="field mb-1">
              <visual-editor v-if="form.type === 'campaign_visual'" name="body" :source="form.bodySource"
                @change="onChangeVisualEditor" height="70vh" />
            </div>

            <div v-else class="field">
              <label class="block mb-1 text-sm font-medium">{{ $t('templates.rawHTML') }}</label>
              <code-editor lang="html" v-model="form.body" name="body" />
            </div>
          </template>

          <p class="template-help">
            <template v-if="form.type === 'campaign'">
              {{ $t('templates.placeholderHelp', { placeholder: egPlaceholder }) }}
            </template>
            <a target="_blank" rel="noopener noreferer" href="https://listmonk.app/docs/templating">
              {{ $t('globals.buttons.learnMore') }}
            </a>
          </p>
      </div>

      <div class="lm-form-footer">
        <PvButton @click="$emit('close')" :label="$t('globals.buttons.close')" severity="secondary" />
        <PvButton v-if="$can('templates:manage')" type="submit" severity="primary" :loading="loading.templates"
          :label="$t('globals.buttons.save')" />
      </div>
    </form>
    <campaign-preview v-if="previewItem" is-post type="template" :title="previewItem.name"
      :template-type="previewItem.type" :body="form.body" @close="onTogglePreview" />
  </section>
</template>

<script>
import { mapState } from 'pinia';
import { useMainStore } from '../store';
import CampaignPreview from '../components/CampaignPreview.vue';
import CodeEditor from '../components/CodeEditor.vue';
import VisualEditor from '../components/VisualEditor.vue';
import CopyText from '../components/CopyText.vue';

export default {
  components: {
    CampaignPreview,
    CopyText,
    'code-editor': CodeEditor,
    'visual-editor': VisualEditor,
  },

  props: {
    data: { type: Object, default: () => { } },
    isEditing: { type: Boolean, default: false },
  },

  emits: ['finished', 'close'],

  data() {
    return {
      // Binds form input values.
      form: {
        name: '',
        subject: '',
        type: 'campaign',
        optin: '',
        body: null,
        bodySource: null,
      },
      previewItem: null,
      egPlaceholder: '{{ template "content" . }}',
    };
  },

  methods: {
    onTogglePreview() {
      this.previewItem = !this.previewItem ? this.form : null;
    },

    onPreviewShortcut(e) {
      if (e.key === 'F9') {
        this.onTogglePreview();
        e.preventDefault();
      }
    },

    onSubmit() {
      if (this.isEditing) {
        this.updateTemplate();
        return;
      }

      this.createTemplate();
    },

    createTemplate() {
      const data = {
        id: this.data.id,
        name: this.form.name,
        type: this.form.type,
        subject: this.form.subject,
        body: this.form.body,
        body_source: this.form.bodySource,
      };

      this.$api.createTemplate(data).then((d) => {
        this.$emit('finished');
        this.$emit('close');
        this.$utils.toast(this.$t('globals.messages.created', { name: d.name }));
      });
    },

    updateTemplate() {
      const data = {
        id: this.data.id,
        name: this.form.name,
        type: this.form.type,
        subject: this.form.subject,
        body: this.form.body,
        body_source: this.form.bodySource,
      };

      this.$api.updateTemplate(data).then((d) => {
        this.$emit('finished');
        this.$emit('close');
        this.$utils.toast(`'${d.name}' updated`);
      });
    },

    onChangeVisualEditor({ source, body }) {
      this.form.body = body;
      this.form.bodySource = source;
    },
  },

  computed: {
    ...mapState(useMainStore, ['loading']),
  },

  mounted() {
    this.form = { ...this.$props.data };

    this.$nextTick(() => {
      this.$refs.focus.focus();
    });

    window.addEventListener('keydown', this.onPreviewShortcut);
  },

  beforeUnmount() {
    window.removeEventListener('keydown', this.onPreviewShortcut);
  },
};
</script>

<style scoped lang="scss">

.template-help { font-size: 0.78rem; color: var(--lm-text-subtle); margin: 0; }
</style>
