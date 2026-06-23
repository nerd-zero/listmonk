<template>
  <section class="campaign">
    <header class="grid page-header">
      <div class="col-6">
        <p v-if="isEditing && data.status" class="tags">
          <PvTag v-if="isEditing" :class="data.status" :value="$t(`campaigns.status.${data.status}`)" />
          <PvTag v-if="data.type === 'optin'" :class="data.type" :value="$t('lists.optin')" />
          <span v-if="isEditing" class="has-text-grey-light is-size-7" :data-campaign-id="data.id">
            {{ $t('globals.fields.id') }}: <copy-text :text="`${data.id}`" />
            {{ $t('globals.fields.uuid') }}: <copy-text :text="data.uuid" />
          </span>
        </p>
        <h4 v-if="isEditing" class="title is-4">
          {{ data.name }}
        </h4>
        <h4 v-else class="title is-4">
          {{ $t('campaigns.newCampaign') }}
        </h4>
      </div>

      <div class="col-6">
        <div v-if="canManage || canSend" class="buttons">
          <div class="field is-grouped" v-if="isEditing && canEdit">
            <div class="field" v-if="canManage">
              <PvButton @click="() => onSubmit('update')" :loading="loading.campaigns" severity="primary"
                icon="pi pi-save" data-cy="btn-save" aria-keyshortcuts="ctrl+s">
                <span class="has-kbd">{{ $t('globals.buttons.saveChanges') }} <span class="kbd">Ctrl+S</span></span>
              </PvButton>
            </div>
            <div class="field" v-if="canSend && canStart">
              <PvButton @click="startCampaign" :loading="loading.campaigns" severity="primary"
                icon="pi pi-send" data-cy="btn-start" :label="$t('campaigns.start')" />
            </div>
            <div class="field" v-if="canSend && canSchedule">
              <PvButton @click="startCampaign" :loading="loading.campaigns" severity="primary"
                icon="pi pi-clock" data-cy="btn-schedule" :label="$t('campaigns.schedule')" />
            </div>
            <div class="field" v-if="canSend && canUnSchedule">
              <PvButton @click="$utils.confirm(null, unscheduleCampaign)" :loading="loading.campaigns"
                severity="primary" icon="pi pi-clock" data-cy="btn-unschedule" :label="$t('campaigns.unSchedule')" />
            </div>
          </div>
        </div>
      </div>
    </header>

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
                <form @submit.prevent="() => onSubmit(isNew ? 'create' : 'update')">
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
                        <select v-model="form.messenger" name="messenger" :disabled="!canEdit" required
                          class="w-full">
                          <template v-if="emailMessengers.length > 1">
                            <optgroup label="email">
                              <option v-for="m in emailMessengers" :value="m" :key="m">
                                {{ m }}
                              </option>
                            </optgroup>
                          </template>
                          <template v-else>
                            <option value="email">email</option>
                          </template>
                          <option v-for="m in otherMessengers" :value="m" :key="m">{{ m }}</option>
                        </select>
                      </div>
                    </div>
                    <div class="col-6">
                      <div class="field mr-4 mb-0">
                        <label class="block mb-1 text-sm font-medium">{{ $t('campaigns.format') }}</label>
                        <select v-model="form.content.contentType" :disabled="!canEdit || isEditing"
                          class="w-full">
                          <option v-for="(name, f) in contentTypes" :key="f" name="format" :value="f"
                            :data-cy="`check-${f}`">
                            {{ name }}
                          </option>
                        </select>
                      </div>
                    </div>
                  </div>

                  <div class="field">
                    <label class="block mb-1 text-sm font-medium">{{ $t('globals.terms.tags') }}</label>
                    <PvAutoComplete v-model="form.tags" name="tags" :disabled="!canEdit"
                      :placeholder="$t('globals.terms.tags')" multiple />
                  </div>
                  <hr />

                  <div class="grid">
                    <div class="col-4">
                      <div class="field" data-cy="btn-send-later">
                        <label class="block mb-1 text-sm font-medium">{{ $t('campaigns.sendLater') }}</label>
                        <div class="flex items-center gap-2">
                          <PvToggleSwitch v-model="form.sendLater" :disabled="!canEdit" />
                        </div>
                      </div>
                    </div>
                    <div class="col">
                      <br />
                      <div class="field" v-if="form.sendLater" data-cy="send_at">
                        <small class="block mt-1 text-color-secondary">{{ form.sendAtDate ? $utils.duration(Date(), form.sendAtDate) : '' }}</small>
                        <!-- TODO: replace b-datetimepicker with a PrimeVue equivalent (PvDatePicker) -->
                        <PvDatePicker v-model="form.sendAtDate" :disabled="!canEdit" show-time hour-format="24"
                          :placeholder="$t('campaigns.dateAndTime')" required />
                      </div>
                    </div>
                  </div>

                  <div>
                    <p class="has-text-right">
                      <a href="#" @click.prevent="onShowHeaders" data-cy="btn-headers">
                        <i class="pi pi-plus" />{{ $t('settings.smtp.setCustomHeaders') }}
                      </a>
                    </p>
                    <div class="field" v-if="form.headersStr !== '[]' || isHeadersVisible">
                      <small class="block mt-1 text-color-secondary">{{ $t('campaigns.customHeadersHelp') }}</small>
                      <PvTextarea v-model="form.headersStr" name="headers"
                        placeholder="[{&quot;X-Custom&quot;: &quot;value&quot;}, {&quot;X-Custom2&quot;: &quot;value&quot;}]"
                        :disabled="!canEdit" class="w-full" />
                    </div>
                  </div>
                  <hr />

                  <div class="field" v-if="isNew">
                    <PvButton type="submit" severity="primary" :loading="loading.campaigns" data-cy="btn-continue"
                      :label="$t('campaigns.continue')" />
                  </div>
                </form>
              </div>
              <div v-if="canManage" class="col-4 col-offset-1">
                <br />
                <div class="box">
                  <h3 class="title is-size-6">
                    {{ $t('campaigns.sendTest') }}
                  </h3>
                  <div class="field">
                    <small class="block mt-1 text-color-secondary">{{ $t('campaigns.sendTestHelp') }}</small>
                    <PvAutoComplete v-model="form.testEmails" :disabled="isNew"
                      :placeholder="$t('campaigns.testEmails')" multiple />
                  </div>
                  <div class="field">
                    <PvButton @click="() => onSubmit('test')" :loading="loading.campaigns" :disabled="isNew"
                      severity="primary" icon="pi pi-envelope" :label="$t('campaigns.send')" />
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
            <div class="col has-text-right">
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
                  <select v-model="form.archiveTemplateId" name="template"
                    :disabled="!canArchive || !form.archive || form.content.contentType === 'visual'" required
                    class="w-full">
                    <template v-for="t in templates">
                      <option v-if="t.type === 'campaign'" :value="t.id" :key="t.id">
                        {{ t.name }}
                      </option>
                    </template>
                  </select>
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
      <div class="modal-card content" style="width: auto">
        <section expanded class="modal-card-body">
          <media is-modal @selected="onAttachSelect" />
        </section>
      </div>
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
