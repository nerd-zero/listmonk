<template>
  <section class="campaign">
    <div class="page-header">
      <div class="page-header-left">
        <h1 class="page-title">
          <template v-if="isEditing">{{ data.name }}</template>
          <template v-else>{{ $t('campaigns.newCampaign') }}</template>
        </h1>
        <div v-if="isEditing && data.status" class="header-meta">
          <PvTag :class="data.status" :value="$t(`campaigns.status.${data.status}`)" />
          <PvTag v-if="data.type === 'optin'" :class="data.type" :value="$t('lists.optin')" />
          <span class="id-meta" :data-campaign-id="data.id">
            {{ $t('globals.fields.id') }}: <copy-text :text="`${data.id}`" />
            &nbsp;{{ $t('globals.fields.uuid') }}: <copy-text :text="data.uuid" />
          </span>
        </div>
      </div>
      <div v-if="(canManage || canSend) && isEditing && canEdit" class="header-actions">
        <PvButton v-if="canManage" @click="() => onSubmit('update')" :loading="loading.campaigns"
          severity="primary" data-cy="btn-save" aria-keyshortcuts="ctrl+s">
          <i class="pi pi-save" /><span class="has-kbd">{{ $t('globals.buttons.saveChanges') }} <span class="kbd">Ctrl+S</span></span>
        </PvButton>
        <PvButton v-if="canSend && canStart" @click="startCampaign" :loading="loading.campaigns"
          severity="primary" icon="pi pi-send" data-cy="btn-start" :label="$t('campaigns.start')" />
        <PvButton v-if="canSend && canSchedule" @click="startCampaign" :loading="loading.campaigns"
          severity="primary" icon="pi pi-clock" data-cy="btn-schedule" :label="$t('campaigns.schedule')" />
        <PvButton v-if="canSend && canUnSchedule" @click="$utils.confirm(null, unscheduleCampaign)"
          :loading="loading.campaigns" severity="primary" icon="pi pi-clock"
          data-cy="btn-unschedule" :label="$t('campaigns.unSchedule')" />
      </div>
    </div>

    <div v-if="loading.campaigns" class="flex justify-center p-8">
      <PvProgressSpinner />
    </div>

    <PvTabs v-model:value="activeTab" @update:value="onTab">
      <PvTabList>
        <PvTab value="campaign">
          <i class="pi pi-send mr-1" />{{ $tc('globals.terms.campaign') }}
        </PvTab>
        <PvTab value="content" :disabled="isNew">
          <i class="pi pi-file mr-1" />{{ $t('campaigns.content') }}
        </PvTab>
        <PvTab value="attribs" :disabled="isNew">
          <i class="pi pi-code mr-1" />{{ $t('globals.terms.attribs') }}
        </PvTab>
        <PvTab value="archive" :disabled="isNew">
          <i class="pi pi-file mr-1" />{{ $t('campaigns.archive') }}
        </PvTab>
      </PvTabList>

      <PvTabPanels>
        <!-- campaign tab -->
        <PvTabPanel value="campaign">
          <section class="wrap">
            <div class="grid">
              <div class="col-7">
                <form class="campaign-form" @submit.prevent="() => onSubmit(isNew ? 'create' : 'update')">
                  <div class="field">
                    <label class="block mb-1 text-sm font-medium">{{ $t('globals.fields.name') }}</label>
                    <PvInputText :maxlength="200" ref="focus" v-model="form.name" name="name" :disabled="!canEdit"
                      :placeholder="$t('globals.fields.name')" required autofocus class="w-full" />
                  </div>

                  <div class="field">
                    <label class="block mb-1 text-sm font-medium">{{ $t('campaigns.subject') }}</label>
                    <PvInputText :maxlength="5000" v-model="form.subject" name="subject" :disabled="!canEdit"
                      :placeholder="$t('campaigns.subject')" required class="w-full" />
                  </div>

                  <div class="field">
                    <label class="block mb-1 text-sm font-medium">{{ $t('campaigns.fromAddress') }}</label>
                    <PvInputText :maxlength="200" v-model="form.fromEmail" name="from_email" :disabled="!canEdit"
                      :placeholder="$t('campaigns.fromAddressPlaceholder')" required class="w-full" />
                  </div>

                  <list-selector v-model="form.lists" :selected="form.lists" :all="lists.results" :disabled="!canEdit"
                    :label="$t('globals.terms.lists')" :placeholder="$t('campaigns.sendToLists')" />

                  <div class="grid">
                    <div class="col-6">
                      <div class="field">
                        <label class="block mb-1 text-sm font-medium">{{ $tc('globals.terms.messenger') }}</label>
                        <PvSelect v-model="form.messenger" :options="allMessengers" :disabled="!canEdit"
                          required class="w-full" />
                      </div>
                    </div>
                    <div class="col-6">
                      <div class="field">
                        <label class="block mb-1 text-sm font-medium">{{ $t('campaigns.format') }}</label>
                        <PvSelect v-model="form.content.contentType" :options="contentTypeOptions"
                          option-label="label" option-value="value"
                          :disabled="!canEdit || isEditing" class="w-full" />
                      </div>
                    </div>
                  </div>

                  <div class="field">
                    <label class="block mb-1 text-sm font-medium">{{ $t('globals.terms.tags') }}</label>
                    <PvAutoComplete v-model="form.tags" name="tags" :disabled="!canEdit"
                      :placeholder="$t('globals.terms.tags')" multiple class="w-full" />
                  </div>

                  <div class="form-divider" />

                  <div class="field" data-cy="btn-send-later">
                    <div class="flex items-center gap-2 mb-1">
                      <PvToggleSwitch v-model="form.sendLater" :disabled="!canEdit" />
                      <span class="text-sm font-medium">{{ $t('campaigns.sendLater') }}</span>
                    </div>
                    <div v-if="form.sendLater" data-cy="send_at" class="mt-2">
                      <PvDatePicker v-model="form.sendAtDate" :disabled="!canEdit" show-time hour-format="24"
                        :placeholder="$t('campaigns.dateAndTime')" required />
                      <small v-if="form.sendAtDate" class="block mt-1 text-color-secondary">
                        {{ $utils.duration(Date(), form.sendAtDate) }}
                      </small>
                    </div>
                  </div>

                  <div class="field">
                    <a href="#" class="form-link" @click.prevent="onShowHeaders" data-cy="btn-headers">
                      <i class="pi pi-plus" />{{ $t('settings.smtp.setCustomHeaders') }}
                    </a>
                    <div v-if="form.headersStr !== '[]' || isHeadersVisible" class="mt-2">
                      <PvTextarea v-model="form.headersStr" name="headers"
                        placeholder="[{&quot;X-Custom&quot;: &quot;value&quot;}, {&quot;X-Custom2&quot;: &quot;value&quot;}]"
                        :disabled="!canEdit" class="w-full" />
                      <small class="block mt-1 text-color-secondary">{{ $t('campaigns.customHeadersHelp') }}</small>
                    </div>
                  </div>

                  <div class="form-divider" />

                  <div class="field" v-if="isNew">
                    <PvButton type="submit" severity="primary" :loading="loading.campaigns" data-cy="btn-continue"
                      :label="$t('campaigns.continue')" />
                  </div>
                </form>
              </div>
              <div v-if="canManage" class="col-4 col-offset-1">
                <div class="test-message-card">
                  <div class="test-message-card__header">
                    <i class="pi pi-envelope" />
                    <span>{{ $t('campaigns.sendTest') }}</span>
                  </div>
                  <div class="test-message-card__body">
                    <small class="block mb-2 text-color-secondary">{{ $t('campaigns.sendTestHelp') }}</small>
                    <PvAutoComplete v-model="form.testEmails" :disabled="isNew"
                      :placeholder="$t('campaigns.testEmails')" multiple class="w-full mb-3" />
                    <PvButton @click="() => onSubmit('test')" :loading="loading.campaigns" :disabled="isNew"
                      severity="primary" icon="pi pi-send" :label="$t('campaigns.send')" class="w-full"
                      justify="center" />
                  </div>
                </div>
              </div>
            </div>
          </section>
        </PvTabPanel><!-- campaign -->

        <!-- content tab -->
        <PvTabPanel value="content">
          <editor v-if="data.id" v-model="form.content" :id="data.id" :title="data.name" :disabled="!canEdit"
            :templates="templates" :content-types="contentTypes" />

          <div class="grid">
            <div class="col-6">
              <p v-if="!isAttachFieldVisible" class="is-size-6 has-text-grey">
                <a href="#" @click.prevent="onShowAttachField()" data-cy="btn-attach">
                  <i class="pi pi-upload" />
                  {{ $t('campaigns.addAttachments') }}
                </a>
              </p>

              <div class="field" v-if="isAttachFieldVisible" data-cy="media">
                <label class="block mb-1 text-sm font-medium">{{ $t('campaigns.attachments') }}</label>
                <PvAutoComplete v-model="form.media" name="media" ref="media" option-label="filename"
                  @focus="onOpenAttach" :disabled="!canEdit" multiple />
              </div>
            </div>
            <div class="col" style="text-align:right">
              <a href="https://listmonk.app/docs/templating/#template-expressions" target="_blank"
                rel="noopener noreferer">
                <i class="pi pi-code" /> {{ $t('campaigns.templatingRef') }}</a>
              <span v-if="canEdit && form.content.contentType !== 'plain'" class="is-size-6 has-text-grey ml-6">
                <a v-if="form.altbody === null" href="#" @click.prevent="onAddAltBody">
                  <i class="pi pi-file" /> {{ $t('campaigns.addAltText') }}
                </a>
                <a v-else href="#" @click.prevent="$utils.confirm(null, onRemoveAltBody)">
                  <i class="pi pi-trash" />
                  {{ $t('campaigns.removeAltText') }}
                </a>
              </span>
            </div>
          </div>

          <div v-if="canEdit && form.content.contentType !== 'plain'" class="alt-body">
            <PvTextarea v-if="form.altbody !== null" v-model="form.altbody" :disabled="!canEdit" class="w-full" />
          </div>
        </PvTabPanel><!-- content -->

        <!-- attribs tab -->
        <PvTabPanel value="attribs">
          <section class="wrap">
            <div class="field">
              <label class="block mb-1 text-sm font-medium">{{ $t('globals.terms.attribs') }}</label>
              <PvTextarea v-model="form.attribsStr" :disabled="!canEdit" rows="15" class="w-full" />
              <small class="block mt-1 text-color-secondary">{{ $t('campaigns.attribsHelp') }}</small>
            </div>
          </section>
        </PvTabPanel><!-- attribs -->

        <!-- archive tab -->
        <PvTabPanel value="archive">
          <section class="wrap">
            <div class="grid">
              <div class="col-4">
                <div class="field" data-cy="btn-archive">
                  <label class="block mb-1 text-sm font-medium">{{ $t('campaigns.archiveEnable') }}</label>
                  <small class="block mt-1 text-color-secondary">{{ $t('campaigns.archiveHelp') }}</small>
                  <div class="grid">
                    <div class="col">
                      <div class="flex items-center gap-2">
                        <PvToggleSwitch data-cy="btn-archive" v-model="form.archive" :disabled="!canArchive" />
                      </div>
                    </div>
                    <div class="col-12">
                      <a :href="`${serverConfig.root_url}/archive/${data.uuid}`" target="_blank" rel="noopener noreferer"
                        :class="{ 'has-text-grey-light': !form.archive }" aria-label="$t('campaigns.archive')">
                        <i class="pi pi-external-link" />
                      </a>
                    </div>
                  </div>
                </div>
              </div>
              <div class="col-8">
                <div class="field is-grouped" style="justify-content: flex-end;">
                  <div class="field" v-if="!canEdit && canArchive">
                    <PvButton @click="onUpdateCampaignArchive" :loading="loading.campaigns" severity="primary"
                      icon="pi pi-save" data-cy="btn-save" :label="$t('globals.buttons.saveChanges')" />
                  </div>
                </div>
              </div>
            </div>

            <div class="grid">
              <div class="col-6">
                <div class="field">
                  <label class="block mb-1 text-sm font-medium">{{ $tc('globals.terms.template') }}</label>
                  <PvSelect v-model="form.archiveTemplateId" :options="campaignTemplates"
                    option-label="name" option-value="id"
                    :disabled="!canArchive || !form.archive || form.content.contentType === 'visual'"
                    required class="w-full" />
                </div>
              </div>

              <div class="col-6">
                <div class="field is-grouped" style="justify-content: flex-end;">
                  <div class="field" v-if="form.archive && (!this.form.archiveMetaStr || this.form.archiveMetaStr === '{}')">
                    <a class="button is-primary" href="#" @click.prevent="onFillArchiveMeta" aria-label="{}"><i class="pi pi-code" /></a>
                  </div>
                  <div class="field" v-if="form.archive">
                    <PvButton @click="onToggleArchivePreview" severity="primary" icon="pi pi-eye"
                      data-cy="btn-preview" :label="$t('campaigns.preview')" />
                  </div>
                </div>
              </div>
            </div>

            <div class="field">
              <div class="field">
                <label class="block mb-1 text-sm font-medium">{{ $t('campaigns.archiveSlug') }}</label>
                <small class="block mt-1 text-color-secondary">{{ $t('campaigns.archiveSlugHelp') }}</small>
                <PvInputText :maxlength="200" ref="focus" v-model="form.archiveSlug" name="archive_slug"
                  data-cy="archive-slug" :disabled="!canArchive || !form.archive" class="w-full" />
              </div>
            </div>
            <div class="field">
              <label class="block mb-1 text-sm font-medium">{{ $t('campaigns.archiveMeta') }}</label>
              <small class="block mt-1 text-color-secondary">{{ $t('campaigns.archiveMetaHelp') }}</small>
              <PvTextarea v-model="form.archiveMetaStr" name="archive_meta" data-cy="archive-meta"
                :disabled="!canArchive || !form.archive" rows="20" class="w-full" />
            </div>
          </section>
        </PvTabPanel><!-- archive -->
      </PvTabPanels>
    </PvTabs>

    <PvDialog v-model:visible="isAttachModalOpen" :style="{ width: '900px' }" :closable="true" modal>
      <media is-modal @selected="onAttachSelect" @close="isAttachModalOpen = false" />
    </PvDialog>

    <campaign-preview v-if="isPreviewingArchive" @close="onToggleArchivePreview" type="campaign" :id="data.id"
      :archive-meta="form.archiveMetaStr" :title="data.title" :content-type="data.contentType"
      :template-id="form.archiveTemplateId" is-post is-archive />
  </section>
</template>

<script>
import dayjs from 'dayjs';
import htmlToPlainText from 'textversionjs';
import { mapState } from 'pinia';
import { useMainStore } from '../store';

import CampaignPreview from '../components/CampaignPreview.vue';
import CopyText from '../components/CopyText.vue';
import Editor from '../components/Editor.vue';
import ListSelector from '../components/ListSelector.vue';
import Media from './Media.vue';

export default {
  components: {
    ListSelector,
    Editor,
    Media,
    CopyText,
    CampaignPreview,
  },

  data() {
    return {
      contentTypes: Object.freeze({
        richtext: this.$t('campaigns.richText'),
        html: this.$t('campaigns.rawHTML'),
        markdown: this.$t('campaigns.markdown'),
        plain: this.$t('campaigns.plainText'),
        visual: this.$t('campaigns.visual'),
      }),

      isNew: false,
      isEditing: false,
      isHeadersVisible: false,
      isAttachFieldVisible: false,
      isAttachModalOpen: false,
      isPreviewingArchive: false,
      activeTab: 'campaign',

      data: {},

      // IDs from ?list_id query param.
      selListIDs: [],

      // Binds form input values.
      form: {
        archiveSlug: null,
        name: '',
        subject: '',
        fromEmail: '',
        headersStr: '[]',
        headers: [],
        attribsStr: '{}',
        messenger: 'email',
        lists: [],
        tags: [],
        sendAt: null,
        content: {
          contentType: 'richtext',
          body: '',
          bodySource: null,
          templateId: null,
        },
        altbody: null,
        media: [],

        // Parsed Date() version of send_at from the API.
        sendAtDate: null,
        sendLater: false,
        archive: false,
        archiveMetaStr: '{}',
        archiveMeta: {},
        testEmails: [],
      },
    };
  },

  methods: {
    formatDateTime(s) {
      return dayjs(s).format('YYYY-MM-DD HH:mm');
    },

    onToggleArchivePreview() {
      this.isPreviewingArchive = !this.isPreviewingArchive;
    },

    onAddAltBody() {
      this.form.altbody = htmlToPlainText(this.form.content.body);
    },

    onRemoveAltBody() {
      this.form.altbody = null;
    },

    onShowHeaders() {
      this.isHeadersVisible = !this.isHeadersVisible;
    },

    onShowAttachField() {
      this.isAttachFieldVisible = true;
      this.$nextTick(() => {
        this.$refs.media.focus();
      });
    },

    onOpenAttach() {
      this.isAttachModalOpen = true;
    },

    onAttachSelect(o) {
      if (this.form.media.some((m) => m.id === o.id)) {
        return;
      }

      this.form.media.push(o);
    },

    isUnsaved() {
      return this.data.body !== this.form.content.body
        || this.data.contentType !== this.form.content.contentType;
    },

    onTab(tab) {
      if (tab === 'content' && window.tinymce && window.tinymce.editors.length > 0) {
        this.$nextTick(() => {
          window.tinymce.editors[0].focus();
        });
      }

      // this.$router.replace({ hash: `#${tab}` });
      window.history.replaceState({}, '', `#${tab}`);
    },

    onFillArchiveMeta() {
      const archiveStr = `{"email": "email@domain.com", "name": "${this.$t('globals.fields.name')}", "attribs": {}}`;
      this.form.archiveMetaStr = this.$utils.getPref('campaign.archiveMetaStr') || JSON.stringify(JSON.parse(archiveStr), null, 4);
    },

    onSubmit(typ) {
      // Validate custom JSON headers.
      if (this.form.headersStr && this.form.headersStr !== '[]') {
        try {
          this.form.headers = JSON.parse(this.form.headersStr);
        } catch (e) {
          this.$utils.toast(e.toString(), 'is-danger');
          return;
        }
      } else {
        this.form.headers = [];
      }

      // Validate archive JSON body.
      if (this.form.archive && this.form.archiveMetaStr) {
        try {
          this.form.archiveMeta = JSON.parse(this.form.archiveMetaStr);
        } catch (e) {
          this.$utils.toast(e.toString(), 'is-danger');
          return;
        }
      }

      // Validate custom JSON attribs.
      let attribs = null;
      if (this.form.attribsStr && this.form.attribsStr.trim()) {
        try {
          attribs = JSON.parse(this.form.attribsStr);
        } catch (e) {
          this.$utils.toast(
            `${this.$t('subscribers.invalidJSON')}: ${e.toString()}`,
            'is-danger',

            3000,
          );
          return;
        }
      }
      this.form.attribs = attribs;

      switch (typ) {
        case 'create':
          this.createCampaign();
          break;
        case 'test':
          this.sendTest();
          break;
        default:
          this.updateCampaign();
          break;
      }
    },

    getCampaign(id) {
      return this.$api.getCampaign(id).then((data) => {
        this.data = data;
        this.form = {
          ...this.form,
          ...data,
          headersStr: JSON.stringify(data.headers, null, 4),
          archiveMetaStr: data.archiveMeta ? JSON.stringify(data.archiveMeta, null, 4) : '{}',
          attribsStr: data.attribs ? JSON.stringify(data.attribs, null, 4) : '{}',

          // The structure that is populated by editor input event.
          content: {
            contentType: data.contentType,
            body: data.body,
            bodySource: data.bodySource,
            templateId: data.templateId,
          },
        };
        this.isAttachFieldVisible = this.form.media.length > 0;

        this.form.media = this.form.media.map((f) => {
          if (!f.id) {
            return { ...f, filename: `❌ ${f.filename}` };
          }
          return f;
        });
      });
    },

    sendTest() {
      const data = {
        id: this.data.id,
        name: this.form.name,
        subject: this.form.subject,
        lists: this.form.lists.map((l) => l.id),
        from_email: this.form.fromEmail,
        messenger: this.form.messenger,
        type: 'regular',
        headers: this.form.headers,
        tags: this.form.tags,
        template_id: this.form.content.templateId,
        content_type: this.form.content.contentType,
        body: this.form.content.body,
        altbody: this.form.content.contentType !== 'plain' ? this.form.altbody : null,
        subscribers: this.form.testEmails,
        media: this.form.media.map((m) => m.id),
      };

      this.$api.testCampaign(data).then(() => {
        this.$utils.toast(this.$t('campaigns.testSent'));
      });
      return false;
    },

    createCampaign() {
      const data = {
        archiveSlug: this.form.subject,
        name: this.form.name,
        subject: this.form.subject,
        lists: this.form.lists.map((l) => l.id),
        from_email: this.form.fromEmail,
        content_type: this.form.content.contentType,
        messenger: this.form.messenger,
        type: 'regular',
        tags: this.form.tags,
        send_at: this.form.sendLater ? this.form.sendAtDate : null,
        headers: this.form.headers,
        attribs: this.form.attribs,
        media: this.form.media.map((m) => m.id),
      };

      this.$api.createCampaign(data).then((d) => {
        this.$router.push({ name: 'campaign', hash: '#content', params: { id: d.id } });
      });
      return false;
    },

    async updateCampaign(typ) {
      const data = {
        archive_slug: this.form.archiveSlug,
        name: this.form.name,
        subject: this.form.subject,
        lists: this.form.lists.map((l) => l.id),
        from_email: this.form.fromEmail,
        messenger: this.form.messenger,
        type: 'regular',
        tags: this.form.tags,
        send_at: this.form.sendLater ? this.form.sendAtDate : null,
        headers: this.form.headers,
        attribs: this.form.attribs,
        template_id: this.form.content.templateId,
        content_type: this.form.content.contentType,
        body: this.form.content.body,
        body_source: this.form.content.bodySource,
        altbody: this.form.content.contentType !== 'plain' ? this.form.altbody : null,
        archive: this.form.archive,
        archive_template_id: this.form.archiveTemplateId,
        archive_meta: this.form.archiveMeta,
        media: this.form.media.map((m) => m.id),
      };

      let typMsg = 'globals.messages.updated';
      if (typ === 'start') {
        typMsg = 'campaigns.started';
      }

      if (!this.form.sendAtDate) {
        this.form.sendLater = false;
      }

      // This promise is used by startCampaign to first save before starting.
      return new Promise((resolve) => {
        this.$api.updateCampaign(this.data.id, data).then((d) => {
          this.data = d;
          this.form.archiveSlug = d.archiveSlug;
          this.form.attribsStr = d.attribs ? JSON.stringify(d.attribs, null, 4) : '{}';

          this.$utils.toast(this.$t(typMsg, { name: d.name }));
          resolve();
        });
      });
    },

    onUpdateCampaignArchive() {
      if (this.isEditing && this.canEdit) {
        return;
      }

      const data = {
        archive: this.form.archive,
        archive_template_id: this.form.archiveTemplateId,
        archive_meta: JSON.parse(this.form.archiveMetaStr),
        archive_slug: this.form.archiveSlug,
      };

      this.$api.updateCampaignArchive(this.data.id, data).then((d) => {
        this.form.archiveSlug = d.archiveSlug;
      });
    },

    // Starts or schedule a campaign.
    startCampaign() {
      if (!this.canStart && !this.canSchedule) {
        return;
      }

      this.$utils.confirm(
        null,
        () => {
          // First save the campaign.
          this.updateCampaign().then(() => {
            // Then start/schedule it.
            let status = '';
            if (this.canStart) {
              status = 'running';
            } else if (this.canSchedule) {
              status = 'scheduled';
            } else {
              return;
            }

            this.$api.changeCampaignStatus(this.data.id, status).then(() => {
              this.$router.push({ name: 'campaigns' });
            });
          });
        },
      );
    },

    unscheduleCampaign() {
      this.$api.changeCampaignStatus(this.data.id, 'draft').then((d) => {
        this.data = d;
      });
    },
  },

  computed: {
    ...mapState(useMainStore, ['serverConfig', 'loading', 'lists', 'templates']),

    canManage() {
      return this.$can('campaigns:manage_all', 'campaigns:manage');
    },

    canSend() {
      return this.$can('campaigns:send');
    },

    canEdit() {
      return this.isNew
        || this.data.status === 'draft' || this.data.status === 'scheduled' || this.data.status === 'paused';
    },

    canSchedule() {
      return (this.data.status === 'draft' || this.data.status === 'paused') && (this.form.sendLater && this.form.sendAtDate);
    },

    canUnSchedule() {
      return this.data.status === 'scheduled';
    },

    canStart() {
      return (this.data.status === 'draft' || this.data.status === 'paused') && !this.form.sendLater;
    },

    canArchive() {
      return this.data.status !== 'cancelled' && this.data.type !== 'optin';
    },

    selectedLists() {
      if (this.selListIDs.length === 0 || !this.lists.results) {
        return [];
      }

      return this.lists.results.filter((l) => this.selListIDs.indexOf(l.id) > -1);
    },

    emailMessengers() {
      return ['email', ...this.serverConfig.messengers.filter((m) => m.startsWith('email-'))];
    },

    otherMessengers() {
      return this.serverConfig.messengers.filter((m) => m !== 'email' && !m.startsWith('email-'));
    },

    allMessengers() {
      return [...this.emailMessengers, ...this.otherMessengers];
    },

    contentTypeOptions() {
      return Object.entries(this.contentTypes).map(([value, label]) => ({ value, label }));
    },

    campaignTemplates() {
      return (this.templates || []).filter((t) => t.type === 'campaign');
    },
  },

  beforeRouteLeave(to, from, next) {
    if (this.isUnsaved()) {
      this.$utils.confirm(this.$t('globals.messages.confirmDiscard'), () => next(true));
      return;
    }
    next(true);
  },

  watch: {
    selectedLists() {
      this.form.lists = this.selectedLists;
    },

    // eslint-disable-next-line func-names
    'data.sendAt': function () {
      if (this.data.sendAt !== null) {
        this.form.sendLater = true;
        this.form.sendAtDate = dayjs(this.data.sendAt).toDate();
      } else {
        this.form.sendLater = false;
        this.form.sendAtDate = null;
      }
    },
  },

  mounted() {
    window.onbeforeunload = () => this.isUnsaved() || null;

    // Fill default form fields.
    this.form.fromEmail = this.serverConfig.from_email;

    // New campaign.
    const { id } = this.$route.params;
    if (id === 'new') {
      this.isNew = true;

      if (this.$route.query.list_id) {
        // Multiple list_id query params.
        let strIds = [];
        if (typeof this.$route.query.list_id === 'object') {
          strIds = this.$route.query.list_id;
        } else {
          strIds = [this.$route.query.list_id];
        }

        this.selListIDs = strIds.map((v) => parseInt(v, 10));
      }
    } else {
      const intID = parseInt(id, 10);
      if (intID <= 0 || Number.isNaN(intID)) {
        this.$utils.toast(this.$t('campaigns.invalid'));
        return;
      }

      this.isEditing = true;
    }

    // Get templates list.
    this.$api.getTemplates().then((data) => {
      if (data.length > 0) {
        if (!this.form.templateId) {
          const tpl = data.find((i) => i.isDefault === true);
          this.form.templateId = tpl.id;
        }
      }
    });

    // Fetch campaign.
    if (this.isEditing) {
      this.getCampaign(id).then(() => {
        if (this.$route.hash !== '') {
          this.activeTab = this.$route.hash.replace('#', '');
        }
      });
    } else {
      this.form.messenger = 'email';
    }

    this.$nextTick(() => {
      this.$refs.focus.focus();
    });

    this.$events.$on('campaign.update', () => {
      this.onSubmit('update');
    });
  },

  beforeUnmount() {
    this.$events.$off('campaign.update');
  },
};
</script>

<style scoped lang="scss">
.campaign {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

// Header
.page-header-left {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}
.header-meta {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
}
.id-meta {
  font-size: 0.75rem;
  color: var(--lm-text-subtle);
}
.header-actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
}

// Pill / segmented-control tabs
:deep(.p-tabs) {
  .p-tablist {
    background: var(--lm-bg-subtle);
    border: 1px solid var(--lm-border);
    border-radius: 10px;
    padding: 4px;
    gap: 2px;
    width: fit-content;
    // hide the sliding ink bar and scroll nav buttons
    .p-tablist-active-bar { display: none !important; }
    .p-tablist-nav-button  { display: none !important; }
    // let the pill container scroll internally if viewport is very narrow
    overflow: visible;
  }

  .p-tab {
    border: none;
    border-radius: 7px;
    background: transparent;
    padding: 0.5rem 1.1rem;
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--lm-text-muted);
    margin-bottom: 0;
    gap: 0.4rem;
    white-space: nowrap;
    transition: background 0.15s ease, color 0.15s ease, box-shadow 0.15s ease;

    &:hover:not([aria-selected='true']):not([data-p-disabled='true']) {
      background: rgba(255, 255, 255, 0.65);
      color: var(--lm-text-secondary);
    }

    &[aria-selected='true'] {
      background: var(--lm-surface);
      color: var(--lm-primary);
      font-weight: 600;
      box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08), 0 1px 2px rgba(0, 0, 0, 0.04);
      border-bottom: none;
    }

    &[data-p-disabled='true'] {
      opacity: 0.4;
      pointer-events: none;
      cursor: not-allowed;
    }
  }

  .p-tabpanels {
    border: none;
    box-shadow: none;
    padding: 1.5rem 0 0 0;
    background: transparent;
  }
}

// Campaign form
.campaign-form {
  display: flex;
  flex-direction: column;
  gap: 1rem;

  .field { margin-bottom: 0; }
}

.form-divider {
  border-top: 1px solid var(--lm-border);
  margin: 0.25rem 0;
}

.form-link {
  font-size: 0.85rem;
  color: var(--lm-primary);
  text-decoration: none;
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;

  &:hover { text-decoration: underline; }
}

// Send test message card
.test-message-card {
  border: 1px solid var(--lm-border);
  border-radius: 10px;
  overflow: hidden;
  background: var(--lm-surface);
  margin-top: 1.75rem;

  &__header {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.75rem 1rem;
    background: var(--lm-bg-subtle);
    border-bottom: 1px solid var(--lm-border);
    font-weight: 600;
    font-size: 0.875rem;
    color: var(--lm-text-secondary);

    .pi { color: var(--lm-primary); font-size: 0.9rem; }
  }

  &__body {
    padding: 1rem;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }
}
</style>
