<template>
  <div class="items">
    <div class="settings-card">
      <div class="grid">
        <div class="col-2">
          <div class="field">
            <div class="flex items-center gap-2">
              <PvToggleSwitch v-model="data.scrub.enabled" name="scrub.enabled" />
              <span>{{ $t('globals.buttons.enabled') }}</span>
            </div>
          </div>
        </div>

        <div class="col-10" :class="{ disabled: !data.scrub.enabled }">
          <div class="field">
            <label class="block mb-1 text-sm font-medium">{{ $t('settings.scrub.url') }}</label>
            <PvInputText v-model="data.scrub.url" name="scrub.url"
              placeholder="https://api.thescrub.app" :maxlength="300"
              :disabled="!data.scrub.enabled" class="w-full" />
            <small class="block mt-1 text-color-secondary">{{ $t('settings.scrub.urlHelp') }}</small>
          </div>

          <div class="field mt-4">
            <label class="block mb-1 text-sm font-medium">{{ $t('settings.scrub.apiKey') }}</label>
            <PvPassword v-model="data.scrub.api_key" name="scrub.api_key"
              :maxlength="300" :feedback="false"
              :placeholder="$t('settings.scrub.apiKeyPlaceholder')"
              :disabled="!data.scrub.enabled" class="w-full" />
            <small class="block mt-1 text-color-secondary">{{ $t('settings.scrub.apiKeyHelp') }}</small>
          </div>

          <div class="field mt-4">
            <label class="block mb-1 text-sm font-medium">{{ $t('settings.scrub.integrationId') }}</label>
            <PvInputNumber v-model="data.scrub.integration_id" name="scrub.integration_id"
              :use-grouping="false" :min="0" :disabled="!data.scrub.enabled" class="w-full" />
            <small class="block mt-1 text-color-secondary">{{ $t('settings.scrub.integrationIdHelp') }}</small>
          </div>

          <div class="field mt-4">
            <PvButton severity="primary" :loading="isTesting"
              :disabled="!data.scrub.enabled || !data.scrub.url || !data.scrub.api_key"
              icon="pi pi-link" :label="$t('settings.scrub.testConnection')"
              @click="testConnection" />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useGlobal } from '../../composables/useGlobal';

const props = defineProps<{ form?: any }>();
const { $api, $utils } = useGlobal();
const { t } = useI18n();
const data = props.form;
const isTesting = ref(false);

async function testConnection() {
  isTesting.value = true;
  try {
    await $api.testScrub({ url: data.scrub.url, api_key: data.scrub.api_key });
    $utils.toast(t('settings.scrub.testSuccess'), 'is-success');
  } catch (e: any) {
    $utils.toast(e.response?.data?.message || t('settings.scrub.testError'), 'is-danger');
  } finally {
    isTesting.value = false;
  }
}
</script>

<style scoped lang="scss">
:deep(.p-password) { width: 100%; }
:deep(.p-password-input) { width: 100%; }
</style>
