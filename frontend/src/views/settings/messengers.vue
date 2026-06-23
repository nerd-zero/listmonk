<template>
  <div>
    <div class="items messengers">
      <div class="block box" v-for="(item, n) in data.messengers" :key="n">
        <div class="field">
          <div class="flex items-center gap-2">
            <PvToggleSwitch v-model="item.enabled" name="enabled" />
            <span>{{ $t('globals.buttons.enabled') }}</span>
          </div>
        </div>
        <div class="field">
          <a @click.prevent="$utils.confirm(null, () => removeMessenger(n))" href="#" class="is-size-7">
            <i class="pi pi-trash" />
            {{ $t('globals.buttons.delete') }}
          </a>
        </div>

        <div :class="{ disabled: !item.enabled }">
          <div class="field">
            <label class="block mb-1 text-sm font-medium">{{ $t('globals.fields.name') }}</label>
            <PvInputText v-model="item.name" name="name" placeholder="mymessenger" :maxlength="200" />
            <small class="block mt-1 text-color-secondary">{{ $t('settings.messengers.nameHelp') }}</small>
          </div>

          <div class="field">
            <label class="block mb-1 text-sm font-medium">{{ $t('settings.messengers.url') }}</label>
            <PvInputText v-model="item.root_url" name="root_url" placeholder="https://postback.messenger.net/path"
              :maxlength="200" type="url" :pattern="'https?://.*'" class="w-full" />
            <small class="block mt-1 text-color-secondary">{{ $t('settings.messengers.urlHelp') }}</small>
          </div>

          <div class="field">
            <label class="block mb-1 text-sm font-medium">{{ $t('settings.messengers.username') }}</label>
            <PvInputText v-model="item.username" name="username" :maxlength="200" />
          </div>

          <div class="field">
            <label class="block mb-1 text-sm font-medium">{{ $t('settings.messengers.password') }}</label>
            <PvPassword v-model="item.password" name="password" :feedback="false"
              :placeholder="$t('globals.messages.passwordChange')" :maxlength="200" />
            <small class="block mt-1 text-color-secondary">{{ $t('globals.messages.passwordChange') }}</small>
          </div>

          <div class="columns">
            <div class="column is-4">
              <div class="field">
                <label class="block mb-1 text-sm font-medium">{{ $t('settings.messengers.maxConns') }}</label>
                <PvInputNumber v-model="item.max_conns" name="max_conns" placeholder="25" :min="1" :max="65535" />
                <small class="block mt-1 text-color-secondary">{{ $t('settings.messengers.maxConnsHelp') }}</small>
              </div>
            </div>
            <div class="column is-4">
              <div class="field">
                <label class="block mb-1 text-sm font-medium">{{ $t('settings.messengers.retries') }}</label>
                <PvInputNumber v-model="item.max_msg_retries" name="max_msg_retries" placeholder="2" :min="1" :max="1000" />
                <small class="block mt-1 text-color-secondary">{{ $t('settings.messengers.retriesHelp') }}</small>
              </div>
            </div>
            <div class="column is-4">
              <div class="field">
                <label class="block mb-1 text-sm font-medium">{{ $t('settings.messengers.timeout') }}</label>
                <PvInputText v-model="item.timeout" name="timeout" placeholder="5s" :pattern="regDuration"
                  :maxlength="10" />
                <small class="block mt-1 text-color-secondary">{{ $t('settings.messengers.timeoutHelp') }}</small>
              </div>
            </div>
          </div>
        </div>
      </div><!-- block -->
    </div><!-- mail-servers -->

    <PvButton @click="addMessenger" icon="pi pi-plus" severity="primary" :label="$t('globals.buttons.addNew')" />
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

  methods: {
    addMessenger() {
      this.data.messengers.push({
        enabled: true,
        root_url: '',
        name: '',
        username: '',
        password: '',
        max_conns: 25,
        max_msg_retries: 2,
        timeout: '5s',
      });

      this.$nextTick(() => {
        const items = document.querySelectorAll('.messengers input[name="name"]');
        items[items.length - 1].focus();
      });
    },

    removeMessenger(i) {
      this.data.messengers.splice(i, 1);
    },
  },
};
</script>
