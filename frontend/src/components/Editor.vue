<template>
  <!-- Two-way Data-Binding -->
  <section class="editor">
    <div class="editor-toolbar">
      <div class="editor-toolbar-left">
        <div class="field">
          <label class="block mb-1 text-sm font-medium">{{ $t('campaigns.format') }}</label>
          <PvSelect v-model="contentTypeSel" :options="contentTypeOptions" option-label="label" option-value="value"
            :disabled="disabled" name="content_type" data-cy="check-format" />
        </div>

        <div class="field" v-if="self.contentType !== 'visual'">
          <label class="block mb-1 text-sm font-medium">{{ $tc('globals.terms.template') }}</label>
          <PvSelect v-model="templateId" :options="templateOptions" option-label="name" option-value="id"
            :disabled="disabled" name="template" />
        </div>

        <div v-else class="field">
          <PvButton v-if="!isVisualTplSelector" @click="onShowVisualTplSelector" severity="secondary"
            icon="pi pi-file-find" data-cy="btn-select-visual-tpl"
            :label="$t('campaigns.importVisualTemplate')" />
          <template v-else>
            <label class="block mb-1 text-sm font-medium">{{ $tc('globals.terms.template') }}</label>
            <div class="flex items-center gap-2">
              <PvSelect v-model="visualTemplateId" :options="templateOptions" option-label="name" option-value="id"
                @change="() => isVisualTplDisabled = false" name="template" :disabled="disabled"
                class="copy-visual-template-list" />
              <PvButton :disabled="disabled || isVisualTplDisabled || !visualTemplateId"
                @click="onImportVisualTpl" severity="primary" icon="pi pi-save"
                data-cy="btn-save-visual-tpl" :label="$t('globals.terms.import')">
                <PvProgressSpinner v-if="loading.templates" style="width:1rem;height:1rem" />
              </PvButton>
            </div>
          </template>
        </div>
      </div>

      <div class="editor-toolbar-right">
        <PvButton @click="onTogglePreview" severity="secondary" outlined data-cy="btn-preview"
          aria-keyshortcuts="F9">
          <i class="pi pi-eye" /><span class="has-kbd">{{ $t('campaigns.preview') }} <span class="kbd">F9</span></span>
        </PvButton>
      </div>
    </div>

    <!-- wsywig //-->
    <richtext-editor v-if="self.contentType === 'richtext'" :disabled="disabled" v-model="self.body" />

    <!-- visual editor //-->
    <visual-editor v-if="self.contentType === 'visual'" :source="self.bodySource" @change="onVisualEditorChange"
      height="65vh" ref="visualEditor" />

    <!-- raw html editor //-->
    <code-editor lang="html" v-if="self.contentType === 'html'" v-model="self.body" key="editor-html" />

    <!-- markdown editor //-->
    <code-editor lang="markdown" v-if="self.contentType === 'markdown'" v-model="self.body" key="editor-markdown" />

    <!-- plain text //-->
    <PvTextarea v-if="self.contentType === 'plain'" v-model="self.body" name="content" ref="plainEditor"
      class="plain-editor" />

    <!-- campaign preview //-->
    <campaign-preview v-if="isPreviewing" is-post @close="onTogglePreview" type="campaign" :id="id" :title="title"
      :content-type="self.contentType" :template-id="templateId" :body="self.body" />
  </section>
</template>

<script>
import { html as beautifyHTML } from 'js-beautify';
import TurndownService from 'turndown';
import { mapState } from 'pinia';
import { useMainStore } from '../store';

import CampaignPreview from './CampaignPreview.vue';
import VisualEditor from './VisualEditor.vue';
import RichtextEditor from './RichtextEditor.vue';
import markdownToVisualBlock from './editor';
import CodeEditor from './CodeEditor.vue';

const turndown = new TurndownService();

export default {
  components: {
    CampaignPreview,
    'code-editor': CodeEditor,
    'visual-editor': VisualEditor,
    'richtext-editor': RichtextEditor,
  },

  props: {
    contentTypes: { type: Object, default: () => ({}) },
    id: { type: Number, default: 0 },
    title: { type: String, default: '' },
    disabled: { type: Boolean, default: false },
    templates: { type: Array, default: null },

    modelValue: {
      type: Object,
      default: () => ({
        body: '',
        bodySource: null,
        contentType: '',
        templateId: null,
      }),
    },
  },

  data() {
    return {
      isPreviewing: false,
      isVisualTplSelector: false,
      isVisualTplDisabled: false,
      contentTypeSel: this.$props.modelValue.contentType,
      templateId: null,
      visualTemplateId: null,
    };
  },

  methods: {
    onContentTypeChange(to, from) {
      if (!this.self.body.trim()) {
        this.convertContentType(to, from);
        return;
      }

      // Ask for confirmation as pretty much all conversions are lossy.
      this.$utils.confirm(
        this.$t('campaigns.confirmSwitchFormat'),
        () => {
          this.convertContentType(to, from);
        },
        () => {
          // Cancelled. Reset the <select> to the last value.
          this.contentTypeSel = from;
        },
      );
    },

    convertContentType(to, from) {
      let body = this.self.body ?? '';
      let bodySource = null;

      // Skip UI update (markdown => richtext, html requires a backenbd call).
      let skip = false;

      // If `from` is HTML content, strip out `<body>..` etc. and keep the beautified HTML.
      let isHTML = false;
      if (from === 'richtext' || from === 'html' || from === 'visual') {
        const d = document.createElement('div');
        d.innerHTML = body;
        body = this.beautifyHTML(d.innerHTML.trim());
        isHTML = true;
      }

      // HTML => Non-HTML.
      if (isHTML) {
        switch (to) {
          case 'plain': {
            const d = document.createElement('div');
            d.innerHTML = body;
            body = this.trimLines(d.innerText.trim(), true);
            break;
          }

          case 'markdown': {
            body = turndown.turndown(body).replace(/\n\n+/ig, '\n\n');
            break;
          }

          case 'visual': {
            const md = turndown.turndown(body).replace(/\n\n+/ig, '\n\n');
            bodySource = JSON.stringify(markdownToVisualBlock(md));
            break;
          }

          default:
            // Switching between HTML formats, no need to do anything further
            // as body is already beautified.
            // richtext|html => visual, the contents are simply lost.
            break;
        }

        // Markdown to HTML requires a backend call.
      } else if (from === 'markdown' && (to === 'richtext' || to === 'html')) {
        skip = true;
        this.$api.convertCampaignContent({
          id: 1, body, from, to,
        }).then((data) => {
          this.$nextTick(() => {
            // Both type + body should be updated in one cycle to avoid firing
            // multiple events.
            this.self.contentType = to;
            this.self.body = this.beautifyHTML(data.trim());
          });
        });

        // Plain to an HTML type, change plain line breaks to HTML breaks.
      } else if (from === 'plain' && (to === 'richtext' || to === 'html')) {
        body = body.replace(/\n/ig, '<br>\n');
      } else if (to === 'visual') {
        bodySource = JSON.stringify(markdownToVisualBlock(body));
      }

      // =======================================================================
      // Reset the campaign template ID if its converted to or from visual template.
      if (to === 'visual' || from === 'visual') {
        this.templateId = null;
        this.self.templateId = null;
      }

      // =======================================================================
      // Apply the conversion on the editor UI.
      if (!skip) {
        this.$nextTick(() => {
          // Both type + body should be updated in one cycle to avoid firing
          // multiple events.
          this.self.contentType = to;
          this.self.body = body;
          this.self.bodySource = bodySource;
        });
      }
    },

    onTogglePreview() {
      this.isPreviewing = !this.isPreviewing;
    },

    onKeyboardShortcut(e) {
      // On F9, toggle the preview.
      if (e.key === 'F9') {
        this.onTogglePreview();
        e.preventDefault();
      }

      // On Ctrl+S, trigger save.
      if (e.ctrlKey && e.key === 's') {
        this.$events.$emit('campaign.update');
        e.preventDefault();
      }
    },

    onVisualEditorChange({ body, source }) {
      this.self.body = body;
      this.self.bodySource = source;
    },

    beautifyHTML(str) {
      // Pad all tags with linebreaks.
      let s = this.trimLines(str.replace(/(<(?!(\/)?a|span)([^>]+)>)/ig, '\n$1\n'), true);
      // Remove extra linebreaks.
      s = s.replace(/\n+/g, '\n');

      return beautifyHTML(s, {
        indent_size: 4,
        indent_char: ' ',
        max_preserve_newlines: 2,
        inline: ['h1', 'h2', 'h3', 'h4', 'h5', 'h6', 'b', 'strong', 'span', 'em', 'i', 'code', 'a'],
      }).trim();
    },

    trimLines(str, removeEmptyLines) {
      const out = str.split('\n');
      for (let i = 0; i < out.length; i += 1) {
        const line = out[i].trim();
        if (removeEmptyLines) {
          out[i] = line;
        } else if (line === '') {
          out[i] = '';
        }
      }

      return out.join('\n').replace(/\n\s*\n\s*\n/g, '\n\n');
    },

    onShowVisualTplSelector() {
      this.isVisualTplSelector = true;
      this.setDefaultTemplate();
    },

    onImportVisualTpl() {
      if (!this.visualTemplateId) {
        return;
      }

      this.$utils.confirm(
        this.$t('campaigns.confirmOverwriteContent'),
        () => {
          // Fetch the template body from the server.
          this.$api.getTemplate(this.visualTemplateId).then((data) => {
            this.self.body = data.body;
            this.self.bodySource = data.bodySource;
            this.isVisualTplDisabled = true;

            this.$refs.visualEditor.render(JSON.parse(data.bodySource));
          });
        },
      );
    },

    setDefaultTemplate() {
      if (this.self.contentType === 'visual') {
        this.visualTemplateId = this.validTemplates[0]?.id || null;
      } else {
        if (this.templateId) {
          return;
        }

        const defaultTemplate = this.validTemplates.find((t) => t.isDefault === true);
        this.templateId = defaultTemplate?.id || this.validTemplates[0]?.id || null;
      }
    },
  },

  mounted() {
    this.contentTypeSel = this.modelValue.contentType;
    this.templateId = this.modelValue.templateId;

    window.addEventListener('keydown', this.onKeyboardShortcut);

    this.$events.$on('campaign.preview', () => {
      this.isPreviewing = true;
    });
  },

  beforeUnmount() {
    window.removeEventListener('keydown', this.onKeyboardShortcut);
    this.$events.$off('campaign.preview');
  },

  computed: {
    ...mapState(useMainStore, ['serverConfig', 'loading']),

    self: {
      get() {
        return this.modelValue;
      },
      set(val) {
        this.$emit('update:modelValue', val);
      },
    },

    validTemplates() {
      const typ = this.self.contentType === 'visual' ? 'campaign_visual' : 'campaign';
      return this.templates.filter((t) => (t.type === typ));
    },

    contentTypeOptions() {
      return Object.entries(this.contentTypes).map(([value, label]) => ({ value, label }));
    },

    templateOptions() {
      return [{ id: null, name: this.$t('globals.terms.none') }, ...this.validTemplates];
    },
  },

  watch: {
    validTemplates() {
      // When the filtered list of validTemplates changes (visual vs. regular),
      // select the appropriate 'default' in the template select list.
      this.setDefaultTemplate();
    },

    contentTypeSel(to, from) {
      // Show the conversion prompt if the value in the dropdown isn't the same
      // as the current selection. This happens when eg: contentTypeSel = html -> visual happens
      // in the selector, the prompt is shown, and Cancel is clicked,
      // at which point, contentTypeSel = html again, which triggers this event.
      if (from !== to && to !== this.self.contentType) {
        this.onContentTypeChange(to, from);
      }
    },

    templateId(to) {
      if (this.self.templateId === to) {
        return;
      }

      this.self.templateId = to;
    },
  },
};

</script>

<style scoped lang="scss">
.editor-toolbar {
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
  gap: 1rem;
  margin-bottom: 1rem;

  &-left {
    display: flex;
    align-items: flex-end;
    gap: 1rem;
    flex-wrap: wrap;

    .field { margin-bottom: 0; }
  }

  &-right {
    flex-shrink: 0;
  }
}
</style>
