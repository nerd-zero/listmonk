<template>
  <section class="media-files">
    <h1 class="title is-4">
      {{ $t('media.title') }}
      <span v-if="media.results && media.results.length > 0">({{ media.results.length }})</span>
      <span class="has-text-grey-light"> / {{ serverConfig.media_provider }}</span>
    </h1>

    <div v-if="isProcessing || loading.media" class="flex justify-center p-8">
      <PvProgressSpinner />
    </div>

    <section class="wrap gallery mt-6">
      <div class="grid mb-4">
        <div class="col">
          <form @submit.prevent="onQueryMedia" class="search">
            <div>
              <div class="field">
                <div class="p-inputgroup">
                  <PvInputText v-model="queryParams.query" name="query" ref="query" data-cy="query" />
                  <PvButton type="submit" severity="primary" icon="pi pi-search" data-cy="btn-query" />
                </div>
              </div>
            </div>
          </form>
        </div>
        <div v-if="$can('media:manage')" class="col-auto">
          <PvButton @click="onToggleForm" icon="pi pi-upload" data-cy="btn-toggle-upload"
            :label="$t('media.upload')" />
        </div>
      </div>

      <div v-if="$can('media:manage') && showUploadForm" class="mb-6">
        <form @submit.prevent="onSubmit" data-cy="upload">
          <div>
            <div class="field">
              <label class="block mb-1 text-sm font-medium">{{ $t('media.upload') }}</label>
              <!-- TODO: PrimeVue FileUpload used in custom mode to replicate drag-drop multi-file upload -->
              <PvFileUpload
                mode="advanced"
                :multiple="true"
                :auto="false"
                :custom-upload="true"
                @select="onFilesSelect"
                @remove="onFileRemove"
                :show-upload-button="false"
                :show-cancel-button="false"
              >
                <template #empty>
                  <div class="has-text-centered section">
                    <p>
                      <i class="pi pi-upload" style="font-size:2rem" />
                    </p>
                    <p>{{ $t('media.uploadHelp') }}</p>
                  </div>
                </template>
              </PvFileUpload>
            </div>
            <div class="tags" v-if="form.files.length > 0">
              <PvTag v-for="(f, i) in form.files" :key="i" :value="f.name" class="mr-1 mb-1">
                <template #default>
                  {{ f.name }}
                  <i class="pi pi-times ml-1" style="cursor:pointer" @click="removeUploadFile(i)" />
                </template>
              </PvTag>
            </div>
            <div class="buttons mt-3">
              <PvButton type="submit" severity="primary" icon="pi pi-upload"
                :disabled="form.files.length === 0" :loading="isProcessing"
                :label="$tc('media.upload')" />
            </div>
          </div>
        </form>
      </div>

      <!-- Pagination -->
      <div v-if="media.total > media.perPage" class="pagination-wrapper mt-5">
        <PvPaginator
          :rows="media.perPage"
          :total-records="media.total"
          :first="(media.page - 1) * media.perPage"
          @page="(e) => onPageChange(e.page + 1)"
        />
      </div>

      <div v-if="loading.media" class="has-text-centered py-6">
        <PvProgressSpinner />
      </div>
      <div v-else-if="media.results && media.results.length > 0" class="grid">
        <div v-for="item in media.results" :key="item.id" class="item">
          <div class="thumb">
            <a @click="(e) => onMediaSelect(item, e)" :href="item.url" target="_blank" rel="noopener noreferer"
              class="thumb-link">
              <div class="thumb-container">
                <img v-if="item.thumbUrl" :src="item.thumbUrl" :title="item.filename" :alt="item.filename" />
                <div v-else class="thumb-placeholder">
                  <span class="file-ext">
                    {{ item.filename.split(".").pop().toUpperCase() }}
                  </span>
                </div>
              </div>
            </a>
            <div class="actions">
              <a href="#" @click.prevent="$utils.confirm(null, () => onDeleteMedia(item.id))" data-cy="btn-delete"
                :aria-label="$t('globals.buttons.delete')" class="delete-btn">
                <i class="pi pi-trash" />
              </a>
            </div>
          </div>
          <div class="info">
            <p class="filename" :title="item.filename">{{ item.filename }}</p>
            <p class="date">{{ $utils.niceDate(item.createdAt, false) }}</p>
          </div>
        </div>
      </div>

      <!-- Empty State -->
      <div v-else-if="!loading.media">
        <empty-placeholder />
      </div>

      <!-- Pagination -->
      <div v-if="media.total > media.perPage" class="pagination-wrapper mt-5">
        <PvPaginator
          :rows="media.perPage"
          :total-records="media.total"
          :first="(media.page - 1) * media.perPage"
          @page="(e) => onPageChange(e.page + 1)"
        />
      </div>
    </section>
  </section>
</template>

<script>
import { mapState } from 'pinia';
import { useMainStore } from '../store';
import EmptyPlaceholder from '../components/EmptyPlaceholder.vue';

export default {
  components: {
    EmptyPlaceholder,
  },

  name: 'Media',

  props: {
    isModal: Boolean,
    type: { type: String, default: '' },
  },

  data() {
    return {
      form: {
        files: [],
      },
      toUpload: 0,
      uploaded: 0,
      showUploadForm: false,

      queryParams: {
        page: 1,
        query: '',
      },
    };
  },

  methods: {
    removeUploadFile(i) {
      this.form.files.splice(i, 1);
    },

    onFilesSelect(event) {
      this.form.files = event.files ? [...event.files] : [];
    },

    onFileRemove(event) {
      this.form.files = this.form.files.filter((f) => f !== event.file);
    },

    getMedia() {
      this.$api.getMedia({
        page: this.queryParams.page,
        query: this.queryParams.query,
      });
    },

    onToggleForm() {
      this.showUploadForm = !this.showUploadForm;
      this.$utils.setPref('media.upload', this.showUploadForm);
    },

    onQueryMedia() {
      this.queryParams.page = 1;
      this.getMedia();
    },

    onMediaSelect(m, e) {
      // If the component is open in the modal mode, close the modal and
      // fire the selection event.
      // Otherwise, do nothing and let the image open like a normal link.
      if (this.isModal) {
        e.preventDefault();
        this.$emit('selected', m);
        this.$parent.close();
      }
    },

    onSubmit() {
      this.toUpload = this.form.files.length;

      // Upload N files with N requests.
      for (let i = 0; i < this.toUpload; i += 1) {
        const params = new FormData();
        params.set('file', this.form.files[i]);
        this.$api.uploadMedia(params).then(() => {
          this.onUploaded();
        }, () => {
          this.onUploaded();
        });
      }
    },

    onDeleteMedia(id) {
      this.$api.deleteMedia(id).then(() => {
        this.getMedia();
      });
    },

    onUploaded() {
      this.uploaded += 1;
      if (this.uploaded >= this.toUpload) {
        this.toUpload = 0;
        this.uploaded = 0;
        this.form.files = [];

        this.getMedia();
      }
    },

    onPageChange(p) {
      this.queryParams.page = p;
      this.getMedia();
    },
  },

  watch: {
    refreshTick() { this.getMedia(); },
  },

  computed: {
    ...mapState(useMainStore, ['refreshTick', 'loading', 'media', 'serverConfig']),

    isProcessing() {
      if (this.toUpload > 0 && this.uploaded < this.toUpload) {
        return true;
      }
      return false;
    },
  },

  mounted() {
    this.$api.getMedia();

    if (this.$utils.getPref('media.upload')) {
      this.showUploadForm = true;
    }
  },
};
</script>
