<template>
  <section class="section">
    <div class="content has-text-grey has-text-centered">
      <p>
        <i :class="'pi ' + (!icon ? 'pi-plus' : icon)" style="font-size:2rem" />
      </p>
      <p>{{ !label ? $t('globals.messages.emptyState') : label }}</p>
    </div>
  </section>
</template>

<script setup lang="ts">
withDefaults(defineProps<{
  icon?: string;
  label?: string;
}>(), {
  icon: '',
  label: '',
});
</script>
