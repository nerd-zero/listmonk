<template>
  <div class="import-page">
    <div class="page-header">
      <h1 class="page-title">{{ $t('import.title') }}</h1>
    </div>

    <div v-if="isLoading" class="flex justify-center p-8">
      <PvProgressSpinner />
    </div>

    <section v-if="isFree()" class="wrap">
      <form @submit.prevent="onUpload" class="box">
        <div>
          <div class="grid">
            <div class="col">
              <div class="field">
                <label class="block mb-1 text-sm font-medium">{{ $t('import.mode') }}</label>
                <div>
                  <label class="flex items-center gap-2 mb-1">
                    <PvRadioButton v-model="form.mode" name="mode" value="subscribe" data-cy="check-subscribe" />
                    <span>{{ $t('import.subscribe') }}</span>
                  </label>
                  <br />
                  <label class="flex items-center gap-2">
                    <PvRadioButton v-model="form.mode" name="mode" value="blocklist" data-cy="check-blocklist" />
                    <span>{{ $t('import.blocklist') }}</span>
                  </label>
                </div>
              </div>
            </div>
            <div class="col">
              <div class="field">
                <label class="block mb-1 text-sm font-medium">{{ $t('globals.fields.status') }}</label>
                <template v-if="form.mode === 'subscribe'">
                  <label class="flex items-center gap-2 mb-1">
                    <PvRadioButton v-model="form.subStatus" name="subStatus" value="unconfirmed"
                      data-cy="check-unconfirmed" />
                    <span>{{ $t('subscribers.status.unconfirmed') }}</span>
                  </label>
                  <label class="flex items-center gap-2">
                    <PvRadioButton v-model="form.subStatus" name="subStatus" value="confirmed"
                      data-cy="check-confirmed" />
                    <span>{{ $t('subscribers.status.confirmed') }}</span>
                  </label>
                </template>

                <label v-else class="flex items-center gap-2">
                  <PvRadioButton v-model="form.subStatus" name="subStatus" value="unsubscribed"
                    data-cy="check-unsubscribed" />
                  <span>{{ $t('subscribers.status.unsubscribed') }}</span>
                </label>
              </div>
            </div>

            <div class="col">
              <div class="field delimiter">
                <label class="block mb-1 text-sm font-medium">{{ $t('import.csvDelim') }}</label>
                <PvInputText v-model="form.delim" name="delim" placeholder="," :maxlength="1" required />
                <small class="block mt-1 text-color-secondary">{{ $t('import.csvDelimHelp') }}</small>
              </div>
            </div>
          </div>

          <div class="grid">
            <div class="col-4">
              <div v-if="form.mode === 'subscribe'" class="field">
                <label class="block mb-1 text-sm font-medium">{{ $t('import.overwriteUserInfo') }}</label>
                <div>
                  <div class="flex items-center gap-2">
                    <PvToggleSwitch v-model="form.overwriteUserInfo" name="overwriteUserInfo"
                      data-cy="overwrite-user-info" />
                  </div>
                </div>
                <small class="block mt-1 text-color-secondary">{{ $t('import.overwriteUserInfoHelp') }}</small>
              </div>
            </div>

            <div class="col">
              <div v-if="form.mode === 'subscribe'" class="field">
                <label class="block mb-1 text-sm font-medium">{{ $t('import.overwriteSubStatus') }}</label>
                <div>
                  <div class="flex items-center gap-2">
                    <PvToggleSwitch v-model="form.overwriteSubStatus" name="overwriteSubStatus"
                      data-cy="overwrite-sub-status" />
                  </div>
                </div>
                <small class="block mt-1 text-color-secondary">{{ $t('import.overwriteSubStatusHelp') }}</small>
              </div>
            </div>
          </div>

          <list-selector v-if="form.mode === 'subscribe'" :label="$t('globals.terms.lists')"
            :placeholder="$t('import.listSubHelp')" :message="$t('import.listSubHelp')" v-model="form.lists"
            :selected="form.lists" :all="lists.results" />
          <hr />

          <div class="field">
            <label class="block mb-1 text-sm font-medium">{{ $t('import.csvFile') }}</label>
            <div class="upload-drop-area" @dragover.prevent @drop.prevent="onFileDrop"
              @click="fileInputEl?.click()">
              <i class="pi pi-upload upload-icon" />
              <p class="upload-label">{{ $t('import.csvFileHelp') }}</p>
              <input ref="fileInputEl" type="file" style="display:none" @change="onFileSelect" />
            </div>
          </div>
          <div class="tags" v-if="form.file">
            <PvTag :value="form.file.name" severity="secondary" style="margin-right:4px" />
            <PvButton icon="pi pi-times" severity="secondary" size="small" text @click="clearFile" />
          </div>
          <div class="buttons">
            <PvButton type="submit" severity="primary"
              :disabled="!form.file || (form.mode === 'subscribe' && form.lists.length === 0)"
              :loading="isProcessing" :label="$t('import.upload')" />
          </div>
        </div>
      </form>

      <div class="import-help">
        <h5 class="import-help-title">{{ $t('import.instructions') }}</h5>
        <p>{{ $t('import.instructionsHelp') }}</p>
        <blockquote class="csv-example">
          <code class="csv-headers"> <span>email,</span> <span>name,</span> <span>attributes</span></code>
        </blockquote>
        <hr />
        <h5 class="import-help-title">{{ $t('import.csvExample') }}</h5>
        <pre class="csv-example" v-text="example" />
      </div>
    </section>

    <section v-if="isRunning() || isDone()" class="import-status">
      <PvProgressBar :value="progress" style="height:6px" />
      <p :class="['import-status-text', { 'import-status-text--success': status.status === 'finished', 'import-status-text--danger': (status.status === 'failed' || status.status === 'stopped') }]">
        {{ status.status }}
      </p>
      <p class="import-count">{{ $t('import.recordsCount', { num: status.imported, total: status.total }) }}</p>
      <PvButton @click="onStopImport" :loading="isProcessing" icon="pi pi-upload" severity="primary"
        :label="isDone() ? $t('import.importDone') : $t('import.stopImport')" />
      <div class="import-logs">
        <log-view :lines="logs" :loading="false" />
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import {
  ref, reactive, computed, watch, onMounted, nextTick,
} from 'vue';
import { storeToRefs } from 'pinia';
import { useI18n } from 'vue-i18n';
import { useRoute } from 'vue-router';
import { useMainStore } from '../store';
import { useGlobal } from '../composables/useGlobal';
import ListSelector from '../components/ListSelector.vue';
import LogView from '../components/LogView.vue';
import { getImport } from '../api/generated/endpoints/import/import';

const { $utils } = useGlobal();
const {
  getImportLogs, getImportStatus, stopImport, importSubscribers,
} = getImport();
const { t } = useI18n();
const route = useRoute();
const { lists } = storeToRefs(useMainStore());

const form = reactive({
  mode: 'subscribe',
  subStatus: 'unconfirmed',
  delim: ',',
  lists: [] as any[],
  overwriteUserInfo: false,
  overwriteSubStatus: false,
  file: null as File | null,
});

const isLoading = ref(true);
const isProcessing = ref(false);
const status = ref<any>({ status: '' });
const logs = ref<string[]>([]);
const pollID = ref<any>(null);
const example = ref('');
const fileInputEl = ref<HTMLInputElement | null>(null);

const progress = computed(() => {
  if (!status.value || !status.value.total > 0) return 0;
  return Math.ceil((status.value.imported / status.value.total) * 100);
});

watch(() => form.mode, () => {
  nextTick(() => {
    form.subStatus = form.mode === 'subscribe' ? 'unconfirmed' : 'unsubscribed';
  });
});

function clearFile() { form.file = null; }

function onFileSelect(e: Event) {
  const target = e.target as HTMLInputElement;
  if (target.files && target.files.length > 0) {
    [form.file] = target.files as any;
  }
}

function onFileDrop(e: DragEvent) {
  if (e.dataTransfer?.files && e.dataTransfer.files.length > 0) {
    [form.file] = e.dataTransfer.files as any;
  }
}

function isFree() { return status.value.status === 'none'; }
function isRunning() { return status.value.status === 'importing' || status.value.status === 'stopping'; }
// eslint-disable-next-line @typescript-eslint/no-unused-vars
function isSuccessful() { return status.value.status === 'finished'; }
// eslint-disable-next-line @typescript-eslint/no-unused-vars
function isFailed() { return status.value.status === 'stopped' || status.value.status === 'failed'; }
function isDone() {
  return status.value.status === 'finished' || status.value.status === 'stopped' || status.value.status === 'failed';
}

function getLogs() {
  getImportLogs().then((data: any) => {
    logs.value = data.split('\n').map((line: string) => line.replace(/\s+importer\.go:\d+:\s*/, ' *: '));
    nextTick(() => {
      const el = document.getElementById('import-log');
      if (el) el.scrollTop = el.scrollHeight;
    });
  });
}

function pollStatus() {
  clearInterval(pollID.value);
  pollID.value = setInterval(() => {
    getImportStatus().then((data: any) => {
      isProcessing.value = false;
      isLoading.value = false;
      status.value = data;
      getLogs();
      if (!isRunning()) clearInterval(pollID.value);
    }, () => {
      isProcessing.value = false;
      isLoading.value = false;
      status.value = { status: 'none' };
      clearInterval(pollID.value);
    });
    return true;
  }, 250);
}

function onStopImport() {
  isProcessing.value = true;
  stopImport().then(() => { pollStatus(); form.file = null; });
}

function renderExample() {
  example.value = 'email,name,attributes\n'
    + 'user1@mail.com,"User One","{""age"": 42, ""planet"": ""Mars""}"\n'
    + 'user2@mail.com,"User Two","{""age"": 24, ""job"": ""Time Traveller""}"';
}

function resetForm() {
  form.mode = 'subscribe';
  form.overwriteUserInfo = false;
  form.overwriteSubStatus = false;
  form.file = null;
  form.lists = [];
  form.subStatus = 'unconfirmed';
  form.delim = ',';
}

function onSubmit() {
  isProcessing.value = true;
  importSubscribers({
    file: form.file as Blob,
    params: JSON.stringify({
      mode: form.mode,
      subscription_status: form.subStatus,
      delim: form.delim,
      lists: form.lists.map((l: any) => l.id),
      overwrite_userinfo: form.overwriteUserInfo,
      overwrite_subscription_status: form.overwriteSubStatus,
    }),
  }).then(() => {
    $utils.toast(t('import.importStarted'));
    pollStatus();
  }, () => {
    isProcessing.value = false;
    form.file = null;
  });
}

function onUpload() {
  if (form.mode === 'subscribe' && form.overwriteSubStatus) {
    $utils.confirm(t('import.subscribeWarning'), onSubmit, resetForm);
    return;
  }
  onSubmit();
}

onMounted(() => {
  renderExample();
  pollStatus();
  const ids = $utils.parseQueryIDs(route.query.list_id);
  if (ids.length > 0 && (lists.value as any).results) {
    nextTick(() => {
      form.lists = (lists.value as any).results.filter((l: any) => ids.indexOf(l.id) > -1);
    });
  }
});
</script>

<style scoped lang="scss">
.import-page { display: flex; flex-direction: column; gap: 1.5rem; }

:deep(.p-tag-secondary) {
  background: var(--lm-bg-subtle);
  color: var(--lm-text-secondary);
  border: 1px solid var(--lm-border);
}

.upload-drop-area {
  border: 2px dashed var(--lm-border);
  border-radius: 8px;
  cursor: pointer;
  padding: 2rem;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
  transition: border-color 0.15s, background 0.15s;

  &:hover {
    border-color: var(--lm-primary);
    background: var(--lm-primary-light);
  }
}

.upload-icon { font-size: 2rem; color: var(--lm-text-muted); }
.upload-label { font-size: 0.875rem; color: var(--lm-text-muted); margin: 0; }

.import-help-title { font-size: 0.95rem; font-weight: 600; color: #374151; margin: 0 0 0.5rem; }

.import-status {
  background: var(--lm-surface); border: 1px solid var(--lm-border); border-radius: 12px;
  padding: 2rem; display: flex; flex-direction: column; align-items: center; gap: 1rem;
  text-align: center;
}
.import-status-text { font-size: 1.25rem; font-weight: 600; text-transform: capitalize; color: #374151; margin: 0; }
.import-status-text--success { color: var(--lm-success); }
.import-status-text--danger { color: var(--lm-danger); }
.import-count { font-size: 0.875rem; color: var(--lm-text-muted); margin: 0; }
.import-logs { width: 100%; }
</style>
