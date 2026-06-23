<template>
  <section>
    <form @submit.prevent="onSubmit">
      <div class="modal-card content template-modal-content" style="width: auto">
        <header class="modal-card-head">
          <PvButton severity="primary" @click="onTogglePreview" class="is-pulled-right" icon="pi pi-file" :label="$t('templates.preview') + ' (F9)'" />

          <template v-if="isEditing">
            <h4>{{ data.name }}</h4>
            <p class="has-text-grey is-size-7">
              {{ $t('globals.fields.id') }}: <span data-cy="id"><copy-text :text="`${data.id}`" /></span>
            </p>
          </template>
          <h4 v-else>
            {{ $t('templates.newTemplate') }}
          </h4>
        </header>
        <section expanded class="modal-card-body mb-0 pb-0">
          <div class="grid">
            <div class="col-9">
              <div class="field">
                <label class="block mb-1 text-sm font-medium">{{ $t('globals.fields.name') }}</label>
                <PvInputText :maxlength="200" ref="focus" v-model="form.name" name="name"
                  :placeholder="$t('globals.fields.name')" required />
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
                  option-label="label" option-value="value" />
              </div>
            </div>
          </div>
          <div class="grid" v-if="form.type === 'tx'">
            <div class="col-12">
              <div class="field">
                <label class="block mb-1 text-sm font-medium">{{ $t('templates.subject') }}</label>
                <PvInputText :maxlength="200" ref="focus" v-model="form.subject" name="name"
                  :placeholder="$t('templates.subject')" required />
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

          <p class="is-size-7">
            <template v-if="form.type === 'campaign'">
              {{ $t('templates.placeholderHelp', { placeholder: egPlaceholder }) }}
            </template>
            <a target="_blank" rel="noopener noreferer" href="https://listmonk.app/docs/templating">
              {{ $t('globals.buttons.learnMore') }}
            </a>
          </p>
        </section>
        <footer class="modal-card-foot has-text-right">
          <PvButton @click="$parent.close()" :label="$t('globals.buttons.close')" />
          <PvButton v-if="$can('templates:manage')" type="submit" severity="primary" :loading="loading.templates"
            :label="$t('globals.buttons.save')" />
        </footer>
      </div>
    </form>
    <campaign-preview v-if="previewItem" is-post type="template" :title="previewItem.name"
      :template-type="previewItem.type" :body="form.body" @close="onTogglePreview" />
  </section>
</template>

<script>
import { mapState } from 'vuex';
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
        this.$parent.close();
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
        this.$parent.close();
        this.$utils.toast(`'${d.name}' updated`);
      });
    },

    onChangeVisualEditor({ source, body }) {
      this.form.body = body;
      this.form.bodySource = source;
    },
  },

  computed: {
    ...mapState(['loading']),
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
