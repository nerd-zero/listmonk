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
          <i class="pi pi-send mr-1" />{{ $t('globals.terms.campaign') }}
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
                    <PvInputText :maxlength="200" ref="focusEl" v-model="form.name" name="name" :disabled="!canEdit"
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
                        <label class="block mb-1 text-sm font-medium">{{ $t('globals.terms.messenger') }}</label>
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
                  <label class="block mb-1 text-sm font-medium">{{ $t('globals.terms.template') }}</label>
                  <PvSelect v-model="form.archiveTemplateId" :options="campaignTemplates"
                    option-label="name" option-value="id"
                    :disabled="!canArchive || !form.archive || form.content.contentType === 'visual'"
                    required class="w-full" />
                </div>
              </div>

              <div class="col-6">
                <div class="field is-grouped" style="justify-content: flex-end;">
                  <div class="field" v-if="form.archive && (!form.archiveMetaStr || form.archiveMetaStr === '{}')">
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

<script setup lang="ts">
import {
  ref, reactive, computed, watch, nextTick, onMounted, onBeforeUnmount,
} from 'vue';
import dayjs from 'dayjs';
import htmlToPlainText from 'textversionjs';
import { storeToRefs } from 'pinia';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter, onBeforeRouteLeave } from 'vue-router';
import { useMainStore } from '../store';
import { useGlobal } from '../composables/useGlobal';
import CampaignPreview from '../components/CampaignPreview.vue';
import CopyText from '../components/CopyText.vue';
import Editor from '../components/Editor.vue';
import ListSelector from '../components/ListSelector.vue';
import Media from './Media.vue';

const {
  $api, $utils, $can, $events,
} = useGlobal();
const { t } = useI18n();
const route = useRoute();
const router = useRouter();
const {
  serverConfig, loading, lists, templates,
} = storeToRefs(useMainStore());

const focusEl = ref<any>(null);
const isNew = ref(false);
const isEditing = ref(false);
const isHeadersVisible = ref(false);
const isAttachFieldVisible = ref(false);
const isAttachModalOpen = ref(false);
const isPreviewingArchive = ref(false);
const activeTab = ref('campaign');
const data = ref<any>({});
const selListIDs = ref<number[]>([]);

const form = reactive<any>({
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
    contentType: 'richtext', body: '', bodySource: null, templateId: null,
  },
  altbody: null,
  media: [],
  sendAtDate: null,
  sendLater: false,
  archive: false,
  archiveMetaStr: '{}',
  archiveMeta: {},
  testEmails: [],
});

const contentTypes = computed(() => Object.freeze({
  richtext: t('campaigns.richText'),
  html: t('campaigns.rawHTML'),
  markdown: t('campaigns.markdown'),
  plain: t('campaigns.plainText'),
  visual: t('campaigns.visual'),
}));

const canManage = computed(() => $can('campaigns:manage_all', 'campaigns:manage'));
const canSend = computed(() => $can('campaigns:send'));
const canEdit = computed(() => isNew.value || data.value.status === 'draft' || data.value.status === 'scheduled' || data.value.status === 'paused');
const canSchedule = computed(() => (data.value.status === 'draft' || data.value.status === 'paused') && form.sendLater && form.sendAtDate);
const canUnSchedule = computed(() => data.value.status === 'scheduled');
const canStart = computed(() => (data.value.status === 'draft' || data.value.status === 'paused') && !form.sendLater);
const canArchive = computed(() => data.value.status !== 'cancelled' && data.value.type !== 'optin');
const selectedLists = computed(() => {
  if (selListIDs.value.length === 0 || !(lists.value as any).results) return [];
  return (lists.value as any).results.filter((l: any) => selListIDs.value.indexOf(l.id) > -1);
});
const allMessengers = computed(() => {
  const sc = serverConfig.value as any;
  const email = ['email', ...(sc.messengers || []).filter((m: string) => m.startsWith('email-'))];
  const others = (sc.messengers || []).filter((m: string) => m !== 'email' && !m.startsWith('email-'));
  return [...email, ...others];
});
const contentTypeOptions = computed(() => Object.entries(contentTypes.value).map(([value, label]) => ({ value, label })));
const campaignTemplates = computed(() => ((templates.value as any[]) || []).filter((tpl: any) => tpl.type === 'campaign'));

function isUnsaved() {
  return data.value.body !== form.content.body || data.value.contentType !== form.content.contentType;
}

function onToggleArchivePreview() { isPreviewingArchive.value = !isPreviewingArchive.value; }
function onAddAltBody() { form.altbody = htmlToPlainText(form.content.body); }
function onRemoveAltBody() { form.altbody = null; }
function onShowHeaders() { isHeadersVisible.value = !isHeadersVisible.value; }

function onShowAttachField() {
  isAttachFieldVisible.value = true;
}

function onOpenAttach() { isAttachModalOpen.value = true; }

function onAttachSelect(o: any) {
  if (!form.media.some((m: any) => m.id === o.id)) form.media.push(o);
}

function onTab(tab: string) {
  if (tab === 'content' && (window as any).tinymce && (window as any).tinymce.editors.length > 0) {
    nextTick(() => { (window as any).tinymce.editors[0].focus(); });
  }
  window.history.replaceState({}, '', `#${tab}`);
}

function onFillArchiveMeta() {
  const archiveStr = `{"email": "email@domain.com", "name": "${t('globals.fields.name')}", "attribs": {}}`;
  form.archiveMetaStr = $utils.getPref('campaign.archiveMetaStr') || JSON.stringify(JSON.parse(archiveStr), null, 4);
}

function onSubmit(typ: string) {
  if (form.headersStr && form.headersStr !== '[]') {
    try { form.headers = JSON.parse(form.headersStr); } catch (e: any) { $utils.toast(e.toString(), 'is-danger'); return; }
  } else { form.headers = []; }
  if (form.archive && form.archiveMetaStr) {
    try { form.archiveMeta = JSON.parse(form.archiveMetaStr); } catch (e: any) { $utils.toast(e.toString(), 'is-danger'); return; }
  }
  let attribs = null;
  if (form.attribsStr && form.attribsStr.trim()) {
    try { attribs = JSON.parse(form.attribsStr); } catch (e: any) { $utils.toast(`${t('subscribers.invalidJSON')}: ${e.toString()}`, 'is-danger', 3000); return; }
  }
  form.attribs = attribs;
  if (typ === 'create') { createCampaign(); } else if (typ === 'test') { sendTest(); } else { updateCampaign(); }
}

function getCampaign(id: string) {
  return $api.getCampaign(id).then((d: any) => {
    data.value = d;
    Object.assign(form, {
      ...d,
      headersStr: JSON.stringify(d.headers, null, 4),
      archiveMetaStr: d.archiveMeta ? JSON.stringify(d.archiveMeta, null, 4) : '{}',
      attribsStr: d.attribs ? JSON.stringify(d.attribs, null, 4) : '{}',
      content: {
        contentType: d.contentType, body: d.body, bodySource: d.bodySource, templateId: d.templateId,
      },
    });
    isAttachFieldVisible.value = form.media.length > 0;
    form.media = form.media.map((f: any) => (!f.id ? { ...f, filename: `❌ ${f.filename}` } : f));
  });
}

function sendTest() {
  $api.testCampaign({
    id: data.value.id,
    name: form.name,
    subject: form.subject,
    lists: form.lists.map((l: any) => l.id),
    from_email: form.fromEmail,
    messenger: form.messenger,
    type: 'regular',
    headers: form.headers,
    tags: form.tags,
    template_id: form.content.templateId,
    content_type: form.content.contentType,
    body: form.content.body,
    altbody: form.content.contentType !== 'plain' ? form.altbody : null,
    subscribers: form.testEmails,
    media: form.media.map((m: any) => m.id),
  }).then(() => { $utils.toast(t('campaigns.testSent')); });
}

function createCampaign() {
  $api.createCampaign({
    archiveSlug: form.subject,
    name: form.name,
    subject: form.subject,
    lists: form.lists.map((l: any) => l.id),
    from_email: form.fromEmail,
    content_type: form.content.contentType,
    messenger: form.messenger,
    type: 'regular',
    tags: form.tags,
    send_at: form.sendLater ? form.sendAtDate : null,
    headers: form.headers,
    attribs: form.attribs,
    media: form.media.map((m: any) => m.id),
  }).then((d: any) => { router.push({ name: 'campaign', hash: '#content', params: { id: d.id } }); });
}

async function updateCampaign(typ?: string) {
  const typMsg = typ === 'start' ? 'campaigns.started' : 'globals.messages.updated';
  if (!form.sendAtDate) form.sendLater = false;
  return new Promise<void>((resolve) => {
    $api.updateCampaign(data.value.id, {
      archive_slug: form.archiveSlug,
      name: form.name,
      subject: form.subject,
      lists: form.lists.map((l: any) => l.id),
      from_email: form.fromEmail,
      messenger: form.messenger,
      type: 'regular',
      tags: form.tags,
      send_at: form.sendLater ? form.sendAtDate : null,
      headers: form.headers,
      attribs: form.attribs,
      template_id: form.content.templateId,
      content_type: form.content.contentType,
      body: form.content.body,
      body_source: form.content.bodySource,
      altbody: form.content.contentType !== 'plain' ? form.altbody : null,
      archive: form.archive,
      archive_template_id: form.archiveTemplateId,
      archive_meta: form.archiveMeta,
      media: form.media.map((m: any) => m.id),
    }).then((d: any) => {
      data.value = d;
      form.archiveSlug = d.archiveSlug;
      form.attribsStr = d.attribs ? JSON.stringify(d.attribs, null, 4) : '{}';
      $utils.toast(t(typMsg, { name: d.name }));
      resolve();
    });
  });
}

function onUpdateCampaignArchive() {
  if (isEditing.value && canEdit.value) return;
  $api.updateCampaignArchive(data.value.id, {
    archive: form.archive,
    archive_template_id: form.archiveTemplateId,
    archive_meta: JSON.parse(form.archiveMetaStr),
    archive_slug: form.archiveSlug,
  }).then((d: any) => { form.archiveSlug = d.archiveSlug; });
}

function startCampaign() {
  if (!canStart.value && !canSchedule.value) return;
  $utils.confirm(null, () => {
    updateCampaign().then(() => {
      let status = '';
      if (canStart.value) { status = 'running'; } else if (canSchedule.value) { status = 'scheduled'; }
      if (!status) return;
      $api.changeCampaignStatus(data.value.id, status).then(() => { router.push({ name: 'campaigns' }); });
    });
  });
}

function unscheduleCampaign() {
  $api.changeCampaignStatus(data.value.id, 'draft').then((d: any) => { data.value = d; });
}

watch(selectedLists, (v) => { form.lists = v; });
watch(() => data.value.sendAt, (v) => {
  if (v !== null) { form.sendLater = true; form.sendAtDate = dayjs(v).toDate(); } else { form.sendLater = false; form.sendAtDate = null; }
});

onBeforeRouteLeave((_to, _from, next) => {
  if (isUnsaved()) {
    $utils.confirm(t('globals.messages.confirmDiscard'), () => next(true));
    return;
  }
  next(true);
});

onMounted(() => {
  window.onbeforeunload = () => isUnsaved() || null;
  form.fromEmail = (serverConfig.value as any).from_email;

  const { id } = route.params as { id: string };
  if (id === 'new') {
    isNew.value = true;
    if (route.query.list_id) {
      const strIds: string[] = typeof route.query.list_id === 'object'
        ? (route.query.list_id as string[]) : [route.query.list_id as string];
      selListIDs.value = strIds.map((v) => parseInt(v, 10));
    }
  } else {
    const intID = parseInt(id, 10);
    if (intID <= 0 || Number.isNaN(intID)) { $utils.toast(t('campaigns.invalid')); return; }
    isEditing.value = true;
  }

  $api.getTemplates().then((tpls: any) => {
    if (tpls.length > 0 && !form.content.templateId) {
      const tpl = tpls.find((i: any) => i.isDefault === true);
      if (tpl) form.content.templateId = tpl.id;
    }
  });

  if (isEditing.value) {
    getCampaign(id).then(() => {
      if (route.hash !== '') activeTab.value = route.hash.replace('#', '');
    });
  } else {
    form.messenger = 'email';
  }

  nextTick(() => { focusEl.value?.focus(); });
  $events.$on('campaign.update', () => { onSubmit('update'); });
});

onBeforeUnmount(() => { $events.$off('campaign.update'); });
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
