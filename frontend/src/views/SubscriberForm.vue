<template>
  <form class="lm-form" @submit.prevent="onSubmit">
    <div class="lm-form-header">
      <div class="lm-form-title-row">
        <h3 class="lm-form-title">{{ isEditing ? data.name : $t('subscribers.newSubscriber') }}</h3>
        <PvTag v-if="isEditing" :severity="data.status === 'enabled' ? 'success' : 'danger'" size="small"
          :value="$t(`subscribers.status.${data.status}`)" />
      </div>
      <p v-if="isEditing" class="lm-form-meta">
        ID: <copy-text :text="`${data.id}`" data-cy="id" /> &nbsp;·&nbsp;
        UUID: <copy-text :text="data.uuid" />
      </p>
    </div>

    <div class="lm-form-body">
      <div class="lm-field">
        <label class="lm-label">{{ $t('subscribers.email') }}</label>
        <PvInputText :maxlength="200" v-model="form.email" name="email" ref="focus"
          :placeholder="$t('subscribers.email')" class="w-full" required />
      </div>

      <div class="lm-field-row">
        <div class="lm-field">
          <label class="lm-label">{{ $t('globals.fields.name') }}</label>
          <PvInputText :maxlength="200" v-model="form.name" name="name"
            :placeholder="$t('globals.fields.name')" class="w-full" />
        </div>
        <div class="lm-field">
          <label class="lm-label">{{ $t('globals.fields.status') }}</label>
          <PvSelect v-model="form.status" name="status" :placeholder="$t('globals.fields.status')" required
            :options="statusOptions" option-label="label" option-value="value" class="w-full" />
          <small class="lm-help">{{ $t('subscribers.blocklistedHelp') }}</small>
        </div>
      </div>

      <PvTabs v-model:value="activeTab">
        <PvTabList>
          <PvTab value="0">{{ $t('globals.terms.lists') }}</PvTab>
          <PvTab value="1">{{ `${$tc('globals.terms.subscriptions', 2)} (${data.lists ? data.lists.length : 0})` }}</PvTab>
          <PvTab value="2" :disabled="bounces.length === 0">{{ `${$t('globals.terms.bounces')} (${bounces.length})` }}</PvTab>
          <PvTab value="3" :disabled="!isEditing">{{ $t('subscribers.activity') }}</PvTab>
        </PvTabList>
        <PvTabPanels>
          <!-- lists -->
          <PvTabPanel value="0">
            <list-selector :label="$t('subscribers.lists')" :placeholder="$t('subscribers.listsPlaceholder')"
              :message="$t('subscribers.listsHelp')" v-model="form.lists" :selected="form.lists" :all="lists.results" />
            <div class="lm-field-row lm-field-row--preconfirm">
              <div class="lm-field">
                <small class="lm-help">{{ $t('subscribers.preconfirmHelp') }}</small>
                <div class="check-row">
                  <PvCheckbox v-model="form.preconfirm" :binary="true" :disabled="!hasOptinList" />
                  <span class="check-label">{{ $t('subscribers.preconfirm') }}</span>
                </div>
              </div>
              <div v-if="$can('subscribers:manage') && isEditing" class="optin-action">
                <a href="#" @click.prevent="sendOptinConfirmation"
                  :class="['optin-link', { 'optin-link--disabled': !hasOptinList }]">
                  <i class="pi pi-envelope" />
                  {{ $t('subscribers.sendOptinConfirm') }}
                </a>
              </div>
            </div>
          </PvTabPanel><!-- lists -->

          <!-- subscriptions -->
          <PvTabPanel value="1">
            <template v-if="data.lists">
              <PvDataTable :value="data.lists" hoverable sort-field="createdAt" class="subscriptions">
                <PvColumn field="name" :header="$tc('globals.terms.list', 1)">
                  <template #body="{ data: row }">
                    <div class="sub-name-cell">
                      <router-link v-if="!row.restricted" :to="`/lists/${row.id}`">{{ row.name }}</router-link>
                      <span v-else class="sub-restricted">{{ row.name }}</span>
                      <PvTag :severity="row.optin === 'double' ? 'success' : 'secondary'" size="small"
                        :data-cy="`optin-${row.optin}`">
                        <i :class="row.optin === 'double' ? 'pi pi-check-circle' : 'pi pi-times-circle'" />
                        {{ ' ' }}{{ $t(`lists.optins.${row.optin}`) }}
                      </PvTag>
                    </div>
                  </template>
                </PvColumn>

                <PvColumn field="status" :header="$t('globals.fields.status')" class="status">
                  <template #body="{ data: row }">
                    <div class="sub-status-cell">
                      <PvTag :severity="subStatusSeverity(row.subscriptionStatus)" size="small"
                        :value="$t(`subscribers.status.${row.subscriptionStatus}`)" />
                      <span v-if="row.optin === 'double' && row.subscriptionMeta.optinIp" class="sub-ip">
                        {{ row.subscriptionMeta.optinIp }}
                      </span>
                    </div>
                  </template>
                </PvColumn>

                <PvColumn field="createdAt" :header="$t('globals.fields.createdAt')">
                  <template #body="{ data: row }">{{ $utils.niceDate(row.subscriptionCreatedAt, true) }}</template>
                </PvColumn>

                <PvColumn field="updatedAt" :header="$t('globals.fields.updatedAt')">
                  <template #body="{ data: row }">{{ $utils.niceDate(row.subscriptionCreatedAt, true) }}</template>
                </PvColumn>
              </PvDataTable>
            </template>
          </PvTabPanel><!-- subscriptions -->

          <!-- bounces -->
          <PvTabPanel value="2" class="bounces">
            <div class="bounces-header">
              <a v-if="isBounceVisible" href="#" class="delete-link" @click.prevent="deleteBounces">
                <i class="pi pi-trash" /> {{ $t('globals.buttons.delete') }}
              </a>
            </div>

            <PvDataTable :value="bounces" hoverable sort-field="createdAt">
              <PvColumn field="campaign" :header="$tc('globals.terms.campaign', 1)">
                <template #body="{ data: row }">
                  <router-link v-if="row.campaign" :to="{ name: 'bounces', query: { campaign_id: row.campaign.id } }">
                    {{ row.campaign.name }}
                  </router-link>
                </template>
              </PvColumn>

              <PvColumn field="createdAt" :header="$t('globals.fields.createdAt')">
                <template #body="{ data: row }">{{ $utils.niceDate(row.createdAt, true) }}</template>
              </PvColumn>

              <PvColumn field="action" :header="$t('globals.fields.type')">
                <template #body="{ data: row }">
                  <div class="bounce-meta-row">
                    <a href="#" class="bounce-source" @click.prevent="toggleMeta(row.id)">
                      {{ row.source }}
                      <i :class="visibleMeta[row.id] ? 'pi pi-chevron-up' : 'pi pi-chevron-down'" />
                    </a>
                  </div>
                  <pre v-if="visibleMeta[row.id]" class="bounce-meta-pre">{{ row.meta }}</pre>
                </template>
              </PvColumn>
            </PvDataTable>
          </PvTabPanel><!-- bounces -->

          <!-- activity -->
          <PvTabPanel value="3" class="activity">
            <subscriber-activity v-if="isEditing && data.id" :subscriber-id="data.id" />
          </PvTabPanel><!-- activity -->
        </PvTabPanels>
      </PvTabs>

      <div class="attribs-field">
        <h5 class="attribs-title">{{ $t('globals.terms.attribs') }}</h5>
        <small class="lm-help">{{ $t('subscribers.attribsHelp') + ' ' + egAttribs }}</small>
        <PvTextarea v-model="form.strAttribs" name="attribs" rows="4" class="w-full" />
        <a href="https://listmonk.app/docs/concepts" target="_blank" rel="noopener noreferrer" class="learn-more-link">
          {{ $t('globals.buttons.learnMore') }} <i class="pi pi-external-link" />
        </a>
      </div>
    </div>

    <div class="lm-form-footer">
      <PvButton @click="$emit('close')" :label="$t('globals.buttons.close')" severity="secondary" />
      <PvButton v-if="$can('subscribers:manage')" type="submit" severity="primary"
        :loading="loading.subscribers" :label="$t('globals.buttons.save')" />
    </div>
  </form>
</template>

<script>
import { mapState } from 'pinia';
import { useMainStore } from '../store';
import ListSelector from '../components/ListSelector.vue';
import CopyText from '../components/CopyText.vue';
import SubscriberActivity from '../components/SubscriberActivity.vue';

export default {
  components: {
    ListSelector,
    CopyText,
    SubscriberActivity,
  },

  props: {
    data: {
      type: Object,
      default: () => ({ lists: [] }),
    },
    isEditing: Boolean,
  },

  emits: ['finished', 'close'],

  data() {
    return {
      form: {
        lists: [],
        strAttribs: '{}',
        status: 'enabled',
        preconfirm: false,
      },
      isBounceVisible: false,
      bounces: [],
      visibleMeta: {},
      activeTab: '0',

      egAttribs: '{"job": "developer", "location": "Mars", "has_rocket": true}',

      statusOptions: [
        { label: this.$t('subscribers.status.enabled'), value: 'enabled' },
        { label: this.$t('subscribers.status.blocklisted'), value: 'blocklisted' },
      ],
    };
  },

  methods: {
    subStatusSeverity(status) {
      const map = {
        confirmed: 'success',
        unconfirmed: 'warn',
        unsubscribed: 'secondary',
        blocklisted: 'danger',
      };
      return map[status] || 'secondary';
    },

    toggleBounces() {
      this.isBounceVisible = !this.isBounceVisible;
    },

    toggleMeta(id) {
      this.visibleMeta[id] = !this.visibleMeta[id];
    },

    deleteBounces(sub) {
      this.$utils.confirm(
        null,
        () => {
          this.$api.deleteSubscriberBounces(this.form.id).then(() => {
            this.getBounces();
            this.$utils.toast(this.$t('globals.messages.deleted', { name: sub.name }));
          });
        },
      );
    },

    getBounces() {
      this.$api.getSubscriberBounces(this.form.id).then((data) => {
        this.bounces = data;
      });
    },

    onSubmit() {
      if (this.isEditing) {
        this.updateSubscriber();
        return;
      }

      this.createSubscriber();
    },

    createSubscriber() {
      let attribs = {};
      if (this.form.strAttribs) {
        attribs = this.validateAttribs(this.form.strAttribs);
        if (!attribs) {
          return;
        }
      }

      const data = {
        email: this.form.email,
        name: this.form.name,
        status: this.form.status,
        attribs,
        preconfirm_subscriptions: this.form.preconfirm,
        lists: this.form.lists.map((l) => l.id),
      };

      this.$api.createSubscriber(data).then((d) => {
        this.$emit('finished');
        this.$emit('close');
        this.$utils.toast(this.$t('globals.messages.created', { name: d.name }));
      });
    },

    updateSubscriber() {
      let attribs = {};
      if (this.form.strAttribs) {
        attribs = this.validateAttribs(this.form.strAttribs);
        if (!attribs) {
          return;
        }
      }

      const data = {
        id: this.form.id,
        email: this.form.email,
        name: this.form.name,
        status: this.form.status,
        preconfirm_subscriptions: this.form.preconfirm,
        attribs,
        lists: this.form.lists.map((l) => l.id),
      };

      this.$api.updateSubscriber(data).then((d) => {
        this.$emit('finished');
        this.$emit('close');
        this.$utils.toast(this.$t('globals.messages.updated', { name: d.name }));
      });
    },

    sendOptinConfirmation() {
      if (!this.hasOptinList) return;
      this.$api.sendSubscriberOptin(this.form.id).then(() => {
        this.$utils.toast(this.$t('subscribers.sentOptinConfirm'));
      });
    },

    validateAttribs(str) {
      let attribs = {};
      try {
        attribs = JSON.parse(str);
      } catch (e) {
        this.$utils.toast(
          `${this.$t('subscribers.invalidJSON')}: ${e.toString()}`,
          'is-danger',
          3000,
        );
        return null;
      }
      if (attribs instanceof Array) {
        this.$utils.toast('Attributes should be a map {} and not an array []', 'is-danger', 3000);
        return null;
      }

      return attribs;
    },
  },

  computed: {
    ...mapState(useMainStore, ['lists', 'loading']),

    hasOptinList() {
      return this.form.lists.some((l) => l.optin === 'double');
    },
  },

  mounted() {
    if (this.$props.isEditing) {
      this.form = {
        ...this.$props.data,
        strAttribs: JSON.stringify(this.$props.data.attribs, null, 4),
      };
    }

    if (this.form.id) {
      this.getBounces();
    }

    this.$nextTick(() => {
      this.$refs.focus.$el.focus();
    });
  },
};
</script>

<style scoped lang="scss">
:deep(.p-tag-secondary) {
  background: var(--lm-bg-subtle);
  color: var(--lm-text-secondary);
  border: 1px solid var(--lm-border);
}

.lm-field { display: flex; flex-direction: column; gap: 0.35rem; margin-bottom: 0; }
.lm-field-row { display: grid; grid-template-columns: 1fr 1fr; gap: 1rem; }
.lm-field-row--preconfirm { align-items: start; margin-top: 0.75rem; }

.lm-label { display: block; font-size: 0.8rem; font-weight: 600; color: var(--lm-text); }
.lm-help { display: block; font-size: 0.75rem; color: var(--lm-text-subtle); line-height: 1.4; }

.check-row { display: flex; align-items: center; gap: 0.5rem; }
.check-label { font-size: 0.875rem; color: var(--lm-text); }

.optin-action { display: flex; align-items: flex-end; padding-bottom: 0.1rem; }
.optin-link {
  font-size: 0.85rem; color: var(--lm-primary); text-decoration: none;
  display: inline-flex; align-items: center; gap: 0.35rem;
  &:hover { text-decoration: underline; }
  &--disabled { opacity: 0.4; pointer-events: none; }
}

.sub-name-cell { display: flex; flex-direction: column; gap: 0.25rem; }
.sub-restricted { color: var(--lm-text-subtle); font-style: italic; }
.sub-status-cell { display: flex; flex-direction: column; gap: 0.2rem; }
.sub-ip { font-size: 0.75rem; color: var(--lm-text-subtle); }

.bounces-header { display: flex; justify-content: flex-end; margin-bottom: 0.5rem; }
.delete-link { font-size: 0.85rem; color: var(--lm-danger); text-decoration: none; display: inline-flex; align-items: center; gap: 0.3rem; &:hover { text-decoration: underline; } }

.bounce-meta-row { display: flex; justify-content: flex-end; }
.bounce-source { font-size: 0.85rem; color: var(--lm-text-muted); text-decoration: none; display: inline-flex; align-items: center; gap: 0.3rem; &:hover { color: var(--lm-text); } }
.bounce-meta-pre { font-size: 0.75rem; margin-top: 0.4rem; background: var(--lm-bg); border-radius: 4px; padding: 0.5rem; overflow-x: auto; }

.attribs-field { display: flex; flex-direction: column; gap: 0.35rem; }
.attribs-title { font-size: 0.9rem; font-weight: 600; color: var(--lm-text); margin: 0; }
.learn-more-link { font-size: 0.78rem; color: var(--lm-primary); text-decoration: none; &:hover { text-decoration: underline; } }
</style>
