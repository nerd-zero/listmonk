<template>
  <div class="items">
    <div class="grid">
      <div class="col-6">
        <div class="field">
          <div class="flex items-center gap-2">
            <PvToggleSwitch v-model="data['privacy.disable_tracking']" name="privacy.disable_tracking" />
            <span>{{ $t('settings.privacy.disableTracking') }}</span>
          </div>
          <small class="block mt-1 text-color-secondary">{{ $t('settings.privacy.disableTrackingHelp') }}</small>
        </div>
      </div>
      <div class="col-6" :class="{ disabled: data['privacy.disable_tracking'] }">
        <div class="field">
          <div class="flex items-center gap-2">
            <PvToggleSwitch v-model="data['privacy.individual_tracking']" :disabled="data['privacy.disable_tracking']"
              name="privacy.individual_tracking" />
            <span>{{ $t('settings.privacy.individualSubTracking') }}</span>
          </div>
          <small class="block mt-1 text-color-secondary">{{ $t('settings.privacy.individualSubTrackingHelp') }}</small>
        </div>
      </div>
    </div>

    <div class="field">
      <div class="flex items-center gap-2">
        <PvToggleSwitch v-model="data['privacy.unsubscribe_header']" name="privacy.unsubscribe_header" />
        <span>{{ $t('settings.privacy.listUnsubHeader') }}</span>
      </div>
      <small class="block mt-1 text-color-secondary">{{ $t('settings.privacy.listUnsubHeaderHelp') }}</small>
    </div>

    <div class="field">
      <div class="flex items-center gap-2">
        <PvToggleSwitch v-model="data['privacy.allow_blocklist']" name="privacy.allow_blocklist" />
        <span>{{ $t('settings.privacy.allowBlocklist') }}</span>
      </div>
      <small class="block mt-1 text-color-secondary">{{ $t('settings.privacy.allowBlocklistHelp') }}</small>
    </div>

    <div class="field">
      <div class="flex items-center gap-2">
        <PvToggleSwitch v-model="data['privacy.allow_preferences']" name="privacy.allow_blocklist" />
        <span>{{ $t('settings.privacy.allowPrefs') }}</span>
      </div>
      <small class="block mt-1 text-color-secondary">{{ $t('settings.privacy.allowPrefsHelp') }}</small>
    </div>

    <div class="field">
      <div class="flex items-center gap-2">
        <PvToggleSwitch v-model="data['privacy.allow_export']" name="privacy.allow_export" />
        <span>{{ $t('settings.privacy.allowExport') }}</span>
      </div>
      <small class="block mt-1 text-color-secondary">{{ $t('settings.privacy.allowExportHelp') }}</small>
    </div>

    <div class="field">
      <div class="flex items-center gap-2">
        <PvToggleSwitch v-model="data['privacy.allow_wipe']" name="privacy.allow_wipe" />
        <span>{{ $t('settings.privacy.allowWipe') }}</span>
      </div>
      <small class="block mt-1 text-color-secondary">{{ $t('settings.privacy.allowWipeHelp') }}</small>
    </div>

    <div class="field">
      <div class="flex items-center gap-2">
        <PvToggleSwitch v-model="data['privacy.record_optin_ip']" name="privacy.record_optin_ip" />
        <span>{{ $t('settings.privacy.recordOptinIP') }}</span>
      </div>
      <small class="block mt-1 text-color-secondary">{{ $t('settings.privacy.recordOptinIPHelp') }}</small>
    </div>

    <hr />

    <PvTabs v-model:value="tab">
      <PvTabList>
        <PvTab value="0">{{ `${$t('settings.privacy.domainBlocklist')} (${numBlocked})` }}</PvTab>
        <PvTab value="1">{{ `${$t('settings.privacy.domainAllowlist')} (${numAllowed})` }}</PvTab>
      </PvTabList>
      <PvTabPanels>
        <PvTabPanel value="0">
          <div class="field">
            <PvTextarea v-model="data['privacy.domain_blocklist']" name="privacy.domain_blocklist" rows="6" class="w-full" />
            <small class="block mt-1 text-color-secondary">{{ $t('settings.privacy.domainBlocklistHelp') }}</small>
          </div>
        </PvTabPanel>
        <PvTabPanel value="1">
          <div class="field">
            <PvTextarea v-model="data['privacy.domain_allowlist']" name="privacy.domain_allowlist" rows="6" class="w-full" />
            <small class="block mt-1 text-color-secondary">{{ $t('settings.privacy.domainAllowlistHelp') }}</small>
          </div>
        </PvTabPanel>
      </PvTabPanels>
    </PvTabs>
  </div>
</template>

<script setup lang="ts">
import {
  ref, computed, watch, onMounted,
} from 'vue';
import { useGlobal } from '../../composables/useGlobal';

const props = defineProps<{ form?: any }>();
const { $utils } = useGlobal();
const data = props.form;
const tab = ref('0');

function countItems(str: string) {
  return (str || '').split('\n').filter((line: string) => line.trim()).length;
}

const numBlocked = computed(() => countItems(props.form?.['privacy.domain_blocklist']));
const numAllowed = computed(() => countItems(props.form?.['privacy.domain_allowlist']));

watch(tab, (t) => { $utils.setPref('settings.privacyDomainTab', t); });

onMounted(() => {
  tab.value = String($utils.getPref('settings.privacyDomainTab') || 0);
});
</script>
