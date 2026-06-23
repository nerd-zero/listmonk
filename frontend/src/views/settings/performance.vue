<template>
  <div class="items">
    <div class="field">
      <label class="block mb-1 text-sm font-medium">{{ $t('settings.performance.concurrency') }}</label>
      <PvInputNumber v-model="data['app.concurrency']" name="app.concurrency" placeholder="5" :min="1"
        :max="10000" />
      <small class="block mt-1 text-color-secondary">{{ $t('settings.performance.concurrencyHelp') }}</small>
    </div>

    <div class="field">
      <label class="block mb-1 text-sm font-medium">{{ $t('settings.performance.messageRate') }}</label>
      <PvInputNumber v-model="data['app.message_rate']" name="app.message_rate" placeholder="5" :min="1"
        :max="100000" />
      <small class="block mt-1 text-color-secondary">{{ $t('settings.performance.messageRateHelp') }}</small>
    </div>

    <div class="field">
      <label class="block mb-1 text-sm font-medium">{{ $t('settings.performance.batchSize') }}</label>
      <PvInputNumber v-model="data['app.batch_size']" name="app.batch_size" placeholder="1000" :min="1"
        :max="100000" />
      <small class="block mt-1 text-color-secondary">{{ $t('settings.performance.batchSizeHelp') }}</small>
    </div>

    <div class="field">
      <label class="block mb-1 text-sm font-medium">{{ $t('settings.performance.maxErrThreshold') }}</label>
      <PvInputNumber v-model="data['app.max_send_errors']" name="app.max_send_errors" placeholder="1999"
        :min="0" :max="100000" />
      <small class="block mt-1 text-color-secondary">{{ $t('settings.performance.maxErrThresholdHelp') }}</small>
    </div>

    <div>
      <div class="columns">
        <div class="column is-6">
          <div class="field">
            <div class="flex items-center gap-2">
              <PvToggleSwitch v-model="data['app.message_sliding_window']" name="app.message_sliding_window" />
              <span>{{ $t('settings.performance.slidingWindow') }}</span>
            </div>
            <small class="block mt-1 text-color-secondary">{{ $t('settings.performance.slidingWindowHelp') }}</small>
          </div>
        </div>

        <div class="column is-3" :class="{ disabled: !data['app.message_sliding_window'] }">
          <div class="field">
            <label class="block mb-1 text-sm font-medium">{{ $t('settings.performance.slidingWindowRate') }}</label>
            <PvInputNumber v-model="data['app.message_sliding_window_rate']" name="sliding_window_rate"
              :disabled="!data['app.message_sliding_window']" placeholder="25" :min="1"
              :max="10000000" />
            <small class="block mt-1 text-color-secondary">{{ $t('settings.performance.slidingWindowRateHelp') }}</small>
          </div>
        </div>

        <div class="column is-3" :class="{ disabled: !data['app.message_sliding_window'] }">
          <div class="field">
            <label class="block mb-1 text-sm font-medium">{{ $t('settings.performance.slidingWindowDuration') }}</label>
            <PvInputText v-model="data['app.message_sliding_window_duration']" name="sliding_window_duration"
              :disabled="!data['app.message_sliding_window']" placeholder="1h" :pattern="regDuration" :maxlength="10" />
            <small class="block mt-1 text-color-secondary">{{ $t('settings.performance.slidingWindowDurationHelp') }}</small>
          </div>
        </div>
      </div>
    </div><!-- sliding window -->

    <div>
      <hr />
      <div class="columns">
        <div class="column is-4">
          <div class="field">
            <div class="flex items-center gap-2">
              <PvToggleSwitch v-model="data['app.cache_slow_queries']" name="app.cache_slow_queries" />
              <span>{{ $t('settings.performance.cacheSlowQueries') }}</span>
            </div>
            <small class="block mt-1 text-color-secondary">{{ $t('settings.performance.cacheSlowQueriesHelp') }}</small>
          </div>
        </div>
        <div class="column is-4" :class="{ disabled: !data['app.cache_slow_queries'] }">
          <div class="field">
            <label class="block mb-1 text-sm font-medium">{{ $t('settings.maintenance.cron') }}</label>
            <PvInputText v-model="data['app.cache_slow_queries_interval']" :disabled="!data['app.cache_slow_queries']"
              placeholder="0 3 * * *" />
          </div>
        </div>
        <div class="column">
          <br /><br />
          <a href="https://listmonk.app/docs/maintenance/performance/" target="_blank" rel="noopener noreferer">
            <i class="pi pi-external-link" /> {{ $t('globals.buttons.learnMore') }}
          </a>
        </div>
      </div>
    </div>
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
    };
  },
};
</script>
