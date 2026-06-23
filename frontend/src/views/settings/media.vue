<template>
  <div class="items">
    <div class="grid">
      <div class="col">
        <div class="field">
          <label class="block mb-1 text-sm font-medium">{{ $t('settings.media.provider') }}</label>
          <PvSelect v-model="data['upload.provider']" name="upload.provider"
            :options="[{ label: 'filesystem', value: 'filesystem' }, { label: 's3', value: 's3' }]"
            option-label="label" option-value="value" />
        </div>
      </div>
      <div class="col-10">
        <div class="field">
          <label class="block mb-1 text-sm font-medium">{{ $t('settings.media.upload.extensions') }}</label>
          <PvAutoComplete v-model="data['upload.extensions']" name="tags"
            :suggestions="[]" multiple placeholder="jpg, png, gif .." />
        </div>
      </div>
    </div>
    <hr />

    <div class="block" v-if="data['upload.provider'] === 'filesystem'">
      <div class="field">
        <label class="block mb-1 text-sm font-medium">{{ $t('settings.media.upload.path') }}</label>
        <PvInputText v-model="data['upload.filesystem.upload_path']" name="app.upload_path"
          placeholder="/home/listmonk/uploads" :maxlength="200" required />
        <small class="block mt-1 text-color-secondary">{{ $t('settings.media.upload.pathHelp') }}</small>
      </div>

      <div class="field">
        <label class="block mb-1 text-sm font-medium">{{ $t('settings.media.upload.uri') }}</label>
        <PvInputText v-model="data['upload.filesystem.upload_uri']" name="app.upload_uri" placeholder="/uploads"
          :maxlength="200" required pattern="^\/(.+?)" />
        <small class="block mt-1 text-color-secondary">{{ $t('settings.media.upload.uriHelp') }}</small>
      </div>
    </div><!-- filesystem -->

    <div class="block" v-if="data['upload.provider'] === 's3'">
      <div class="field">
        <label class="block mb-1 text-sm font-medium">{{ $t('settings.media.s3.region') }}</label>
        <PvInputText v-model="data['upload.s3.aws_default_region']" @input="onS3URLChange"
          name="upload.s3.aws_default_region" :maxlength="200" placeholder="ap-south-1" />
      </div>

      <div class="field">
        <label class="block mb-1 text-sm font-medium">{{ $t('settings.media.s3.key') }}</label>
        <PvInputText v-model="data['upload.s3.aws_access_key_id']" name="upload.s3.aws_access_key_id" :maxlength="200" />
      </div>

      <div class="field">
        <label class="block mb-1 text-sm font-medium">{{ $t('settings.media.s3.secret') }}</label>
        <PvPassword v-model="data['upload.s3.aws_secret_access_key']" name="upload.s3.aws_secret_access_key"
          :feedback="false" :maxlength="200" />
        <small class="block mt-1 text-color-secondary">Enter a value to change.</small>
      </div>

      <div class="field">
        <label class="block mb-1 text-sm font-medium">{{ $t('settings.media.s3.bucketType') }}</label>
        <PvSelect v-model="data['upload.s3.bucket_type']" name="upload.s3.bucket_type"
          :options="[{ label: $t('settings.media.s3.bucketTypePrivate'), value: 'private' }, { label: $t('settings.media.s3.bucketTypePublic'), value: 'public' }]"
          option-label="label" option-value="value" />
      </div>

      <div class="field">
        <label class="block mb-1 text-sm font-medium">{{ $t('settings.media.s3.bucket') }}</label>
        <PvInputText v-model="data['upload.s3.bucket']" @input="onS3URLChange" name="upload.s3.bucket" :maxlength="200"
          placeholder="" />
      </div>

      <div class="field">
        <label class="block mb-1 text-sm font-medium">{{ $t('settings.media.s3.bucketPath') }}</label>
        <PvInputText v-model="data['upload.s3.bucket_path']" name="upload.s3.bucket_path" :maxlength="200"
          placeholder="/" />
        <small class="block mt-1 text-color-secondary">{{ $t('settings.media.s3.bucketPathHelp') }}</small>
      </div>

      <div class="field">
        <label class="block mb-1 text-sm font-medium">{{ $t('settings.media.s3.uploadExpiry') }}</label>
        <PvInputText v-model="data['upload.s3.expiry']" name="upload.s3.expiry" placeholder="14d" :pattern="regDuration"
          :maxlength="10" />
        <small class="block mt-1 text-color-secondary">{{ $t('settings.media.s3.uploadExpiryHelp') }}</small>
      </div>

      <div class="field">
        <label class="block mb-1 text-sm font-medium">{{ $t('settings.media.s3.url') }}</label>
        <PvInputText v-model="data['upload.s3.url']" name="upload.s3.url" required
          placeholder="https://s3.$region.amazonaws.com" :maxlength="200" type="url" pattern="https?://.*" />
        <small class="block mt-1 text-color-secondary">{{ $t('settings.media.s3.urlHelp') }}</small>
      </div>

      <div class="field">
        <label class="block mb-1 text-sm font-medium">{{ $t('settings.media.s3.publicURL') }}</label>
        <PvInputText v-model="data['upload.s3.public_url']" name="upload.s3.public_url"
          placeholder="https://files.yourdomain.com" :maxlength="200" pattern="(https?://.*|/.+)" />
        <small class="block mt-1 text-color-secondary">{{ $t('settings.media.s3.publicURLHelp') }}</small>
      </div>
    </div><!-- s3 -->
  </div>
</template>

<script>
import { regDuration } from '../../constants';

export default {
  props: {
    form: {
      type: Object, default: () => { },
    },
  },

  data() {
    return {
      data: this.form,
      regDuration,
      extensions: [],
    };
  },

  methods: {
    onS3URLChange() {
      // If a custom non-AWS URL has been entered, don't update it automatically.
      if (this.data['upload.s3.url'] !== '' && !this.data['upload.s3.url'].match(/amazonaws\.com/)) {
        return;
      }
      this.data['upload.s3.url'] = `https://s3.${this.data['upload.s3.aws_default_region']}.amazonaws.com`;
    },
  },
};
</script>
