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
              @click="$refs.fileInput.click()" style="border:2px dashed #ccc;border-radius:4px;cursor:pointer;">
              <div class="has-text-centered section">
                <p>
                  <i class="pi pi-upload" style="font-size:2rem;" />
                </p>
                <p>{{ $t('import.csvFileHelp') }}</p>
              </div>
              <input ref="fileInput" type="file" style="display:none" @change="onFileSelect" />
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
      <br /><br />

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
      <PvButton @click="stopImport" :loading="isProcessing" icon="pi pi-upload" severity="primary"
        :label="isDone() ? $t('import.importDone') : $t('import.stopImport')" />
      <div class="import-logs">
        <log-view :lines="logs" :loading="false" />
      </div>
    </section>
  </div>
</template>

<script>
import { mapState } from 'pinia';
import { useMainStore } from '../store';
import ListSelector from '../components/ListSelector.vue';
import LogView from '../components/LogView.vue';

export default {
  components: {
    ListSelector,
    LogView,
  },

  props: {
    data: { type: Object, default: () => { } },
    isEditing: { type: Boolean, default: false },
  },

  data() {
    return {
      form: {
        mode: 'subscribe',
        subStatus: 'unconfirmed',
        delim: ',',
        lists: [],
        overwriteUserInfo: false,
        overwriteSubStatus: false,
        file: null,
        example: '',
      },

      // Initial page load still has to wait for the status API to return
      // to either show the form or the status box.
      isLoading: true,

      isProcessing: false,
      status: { status: '' },
      logs: [],
      pollID: null,
    };
  },

  watch: {
    'form.mode': function formMode() {
      // Select the appropriate status radio whenever mode changes.
      this.$nextTick(() => {
        if (this.form.mode === 'subscribe') {
          this.form.subStatus = 'unconfirmed';
        } else {
          this.form.subStatus = 'unsubscribed';
        }
      });
    },
  },

  methods: {
    clearFile() {
      this.form.file = null;
    },

    onFileSelect(e) {
      if (e.target.files && e.target.files.length > 0) {
        [this.form.file] = e.target.files;
      }
    },

    onFileDrop(e) {
      if (e.dataTransfer.files && e.dataTransfer.files.length > 0) {
        [this.form.file] = e.dataTransfer.files;
      }
    },

    // Returns true if we're free to do an upload.
    isFree() {
      if (this.status.status === 'none') {
        return true;
      }
      return false;
    },

    // Returns true if an import is running.
    isRunning() {
      if (this.status.status === 'importing'
        || this.status.status === 'stopping') {
        return true;
      }
      return false;
    },

    isSuccessful() {
      return this.status.status === 'finished';
    },

    isFailed() {
      return (
        this.status.status === 'stopped'
        || this.status.status === 'failed'
      );
    },

    // Returns true if an import has finished (failed or successful).
    isDone() {
      if (this.status.status === 'finished'
        || this.status.status === 'stopped'
        || this.status.status === 'failed'
      ) {
        return true;
      }
      return false;
    },

    pollStatus() {
      // Clear any running status polls.
      clearInterval(this.pollID);

      // Poll for the status as long as the import is running.
      this.pollID = setInterval(() => {
        this.$api.getImportStatus().then((data) => {
          this.isProcessing = false;
          this.isLoading = false;
          this.status = data;
          this.getLogs();

          if (!this.isRunning()) {
            clearInterval(this.pollID);
          }
        }, () => {
          this.isProcessing = false;
          this.isLoading = false;
          this.status = { status: 'none' };
          clearInterval(this.pollID);
        });
        return true;
      }, 250);
    },

    getLogs() {
      this.$api.getImportLogs().then((data) => {
        this.logs = data.split('\n').map((line) => line.replace(/\s+importer\.go:\d+:\s*/, ' *: '));
        this.$nextTick(() => {
          // vue.$refs doesn't work as the logs textarea is rendered dynamically.
          const ref = document.getElementById('import-log');
          if (ref) {
            ref.scrollTop = ref.scrollHeight;
          }
        });
      });
    },

    // Cancel a running import or clears a finished import.
    stopImport() {
      this.isProcessing = true;
      this.$api.stopImport().then(() => {
        this.pollStatus();
        this.form.file = null;
      });
    },

    renderExample() {
      const h = 'email,name,attributes\n'
        + 'user1@mail.com,"User One","{""age"": 42, ""planet"": ""Mars""}"\n'
        + 'user2@mail.com,"User Two","{""age"": 24, ""job"": ""Time Traveller""}"';

      this.example = h;
    },

    resetForm() {
      this.form.mode = 'subscribe';
      this.form.overwriteUserInfo = false;
      this.form.overwriteSubStatus = false;
      this.form.file = null;
      this.form.lists = [];
      this.form.subStatus = 'unconfirmed';
      this.form.delim = ',';
    },

    onUpload() {
      if (this.form.mode === 'subscribe' && this.form.overwriteSubStatus) {
        this.$utils.confirm(this.$t('import.subscribeWarning'), this.onSubmit, this.resetForm);
        return;
      }

      this.onSubmit();
    },

    onSubmit() {
      this.isProcessing = true;

      // Prepare the upload payload.
      const params = new FormData();
      params.set('params', JSON.stringify({
        mode: this.form.mode,
        subscription_status: this.form.subStatus,
        delim: this.form.delim,
        lists: this.form.lists.map((l) => l.id),
        overwrite_userinfo: this.form.overwriteUserInfo,
        overwrite_subscription_status: this.form.overwriteSubStatus,
      }));
      params.set('file', this.form.file);

      // Post.
      this.$api.importSubscribers(params).then(() => {
        // On file upload, show a confirmation.
        this.$utils.toast(this.$t('import.importStarted'));

        // Start polling status.
        this.pollStatus();
      }, () => {
        this.isProcessing = false;
        this.form.file = null;
      });
    },
  },

  computed: {
    ...mapState(useMainStore, ['lists']),

    // Import progress bar value.
    progress() {
      if (!this.status || !this.status.total > 0) {
        return 0;
      }
      return Math.ceil((this.status.imported / this.status.total) * 100);
    },
  },

  mounted() {
    this.renderExample();
    this.pollStatus();

    const ids = this.$utils.parseQueryIDs(this.$route.query.list_id);
    if (ids.length > 0 && this.lists.results) {
      this.$nextTick(() => {
        this.form.lists = this.lists.results.filter((l) => ids.indexOf(l.id) > -1);
      });
    }
  },
};
</script>

<style scoped lang="scss">
.import-page { display: flex; flex-direction: column; gap: 1.5rem; }

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
