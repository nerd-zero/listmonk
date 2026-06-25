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

<script lang="ts">
import { defineComponent } from 'vue';
import { mapState } from 'pinia';
import { useMainStore } from '../../store';
import CodeEditor from '../../components/CodeEditor.vue';

export default defineComponent({
  components: {
    'code-editor': CodeEditor,
  },

  props: {
    form: {
      type: Object, default: () => { },
    },
  },

  data() {
    return {
      data: this.form,
      tab: '0',
    };
  },

  mounted() {
    this.tab = String(this.$utils.getPref('settings.apperanceTab') || '0');
  },

  watch: {
    tab(t) {
      this.$utils.setPref('settings.apperanceTab', t);
    },
  },

  computed: {
    ...mapState(useMainStore, ['settings']),
  },
});

</script>
