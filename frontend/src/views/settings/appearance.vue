<template>
  <div class="items">
    <PvTabs :animated="false" v-model:value="tab">
      <PvTabList>
        <PvTab value="0">{{ $t('settings.appearance.adminName') }}</PvTab>
        <PvTab value="1">{{ $t('settings.appearance.publicName') }}</PvTab>
      </PvTabList>
      <PvTabPanels>
        <PvTabPanel value="0">
          <div class="block">
            {{ $t('settings.appearance.adminHelp') }}
          </div>

          <div class="field">
            <label class="block mb-1 text-sm font-medium">{{ $t('settings.appearance.customCSS') }}</label>
            <code-editor lang="css" v-model="data['appearance.admin.custom_css']" name="body" key="editor-admin-css" />
          </div>

          <div class="field">
            <label class="block mb-1 text-sm font-medium">{{ $t('settings.appearance.customJS') }}</label>
            <code-editor lang="javascript" v-model="data['appearance.admin.custom_js']" name="body"
              key="editor-admin-js" />
          </div>
        </PvTabPanel><!-- admin -->

        <PvTabPanel value="1">
          <div class="block">
            {{ $t('settings.appearance.publicHelp') }}
          </div>

          <div class="field">
            <label class="block mb-1 text-sm font-medium">{{ $t('settings.appearance.customCSS') }}</label>
            <code-editor lang="css" v-model="data['appearance.public.custom_css']" name="body" key="editor-public-css" />
          </div>

          <div class="field">
            <label class="block mb-1 text-sm font-medium">{{ $t('settings.appearance.customJS') }}</label>
            <code-editor lang="javascript" v-model="data['appearance.public.custom_js']" name="body"
              key="editor-public-js" />
          </div>
        </PvTabPanel><!-- public -->
      </PvTabPanels>
    </PvTabs>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue';
import { useGlobal } from '../../composables/useGlobal';
import CodeEditor from '../../components/CodeEditor.vue';

const props = defineProps<{ form?: any }>();
const { $utils } = useGlobal();
const data = props.form;
const tab = ref('0');

onMounted(() => {
  tab.value = String($utils.getPref('settings.apperanceTab') || '0');
});

watch(tab, (t) => { $utils.setPref('settings.apperanceTab', t); });
</script>
