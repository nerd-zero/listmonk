<template>
  <form @submit.prevent="onSubmit">
    <div class="modal-card content" style="width: auto">
      <header class="modal-card-head">
        <PvTag v-if="isEditing" :class="[data.status, 'is-pulled-right']"
          :value="$t(`subscribers.status.${data.status}`)" />
        <h4 v-if="isEditing">
          {{ data.name }}
        </h4>
        <h4 v-else>
          {{ $t('subscribers.newSubscriber') }}
        </h4>

        <p v-if="isEditing" class="has-text-grey is-size-7">
          {{ $t('globals.fields.id') }}: <span data-cy="id"><copy-text :text="`${data.id}`" /></span>
          {{ $t('globals.fields.uuid') }}: <copy-text :text="data.uuid" />
        </p>
      </header>

      <section expanded class="modal-card-body">
        <div class="field">
          <label class="block mb-1 text-sm font-medium">{{ $t('subscribers.email') }}</label>
          <PvInputText :maxlength="200" v-model="form.email" name="email" ref="focus"
            :placeholder="$t('subscribers.email')" required />
        </div>

        <div class="grid">
          <div class="col-8">
            <div class="field">
              <label class="block mb-1 text-sm font-medium">{{ $t('globals.fields.name') }}</label>
              <PvInputText :maxlength="200" v-model="form.name" name="name"
                :placeholder="$t('globals.fields.name')" />
            </div>
          </div>
          <div class="col-4">
            <div class="field">
              <label class="block mb-1 text-sm font-medium">{{ $t('globals.fields.status') }}</label>
              <PvSelect v-model="form.status" name="status" :placeholder="$t('globals.fields.status')" required
                :options="statusOptions" option-label="label" option-value="value" />
              <small class="block mt-1 text-color-secondary">{{ $t('subscribers.blocklistedHelp') }}</small>
            </div>
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
              <div class="grid">
                <div class="col-7">
                  <div class="field">
                    <small class="block mt-1 text-color-secondary">{{ $t('subscribers.preconfirmHelp') }}</small>
                    <div class="flex items-center gap-2">
                      <PvCheckbox v-model="form.preconfirm" :binary="true" :disabled="!hasOptinList" />
                      <span>{{ $t('subscribers.preconfirm') }}</span>
                    </div>
                  </div>
                </div>
                <div v-if="$can('subscribers:manage') && isEditing" class="col-5 has-text-right">
                  <a href="#" @click.prevent="sendOptinConfirmation" :class="{ 'is-disabled': !hasOptinList }">
                    <i class="pi pi-envelope" />
                    {{ $t('subscribers.sendOptinConfirm') }}</a>
                </div>
              </div>
            </PvTabPanel><!-- lists -->

            <!-- subscriptions -->
            <PvTabPanel value="1">
              <template v-if="data.lists">
                <PvDataTable :value="data.lists" hoverable sort-field="createdAt" class="subscriptions">
                  <PvColumn field="name" :header="$tc('globals.terms.list', 1)">
                    <template #body="{ data: row }">
                      <div>
                        <router-link v-if="!row.restricted" :to="`/lists/${row.id}`">
                          {{ row.name }}
                        </router-link>
                        <span v-else class="has-text-grey-light is-italic">{{ row.name }}</span>
                        <br />
                        <PvTag :class="row.optin" :data-cy="`optin-${row.optin}`">
                          <i :class="row.optin === 'double' ? 'pi pi-check-circle' : 'pi pi-times-circle'" />
                          {{ ' ' }}
                          {{ $t(`lists.optins.${row.optin}`) }}
                        </PvTag>{{ ' ' }}
                      </div>
                    </template>
                  </PvColumn>

                  <PvColumn field="status" :header="$t('globals.fields.status')" class="status">
                    <template #body="{ data: row }">
                      <PvTag :class="`status-${row.subscriptionStatus}`"
                        :value="$t(`subscribers.status.${row.subscriptionStatus}`)" />
                      <template v-if="row.optin === 'double' && row.subscriptionMeta.optinIp">
                        <br /><span class="is-size-7">{{ row.subscriptionMeta.optinIp }}</span>
                      </template>
                    </template>
                  </PvColumn>

                  <PvColumn field="createdAt" :header="$t('globals.fields.createdAt')">
                    <template #body="{ data: row }">
                      {{ $utils.niceDate(row.subscriptionCreatedAt, true) }}
                    </template>
                  </PvColumn>

                  <PvColumn field="updatedAt" :header="$t('globals.fields.updatedAt')">
                    <template #body="{ data: row }">
                      {{ $utils.niceDate(row.subscriptionCreatedAt, true) }}
                    </template>
                  </PvColumn>
                </PvDataTable>
              </template>
            </PvTabPanel><!-- subscriptions -->

            <!-- bounces -->
            <PvTabPanel value="2" class="bounces">
              <a href="#" class="is-size-6 is-pulled-right" disabed="true" @click.prevent="deleteBounces"
                v-if="isBounceVisible">
                <i class="pi pi-trash" />
                {{ $t('globals.buttons.delete') }}
              </a>

              <PvDataTable :value="bounces" hoverable sort-field="createdAt" class="bounces">
                <PvColumn field="campaign" :header="$tc('globals.terms.campaign', 1)">
                  <template #body="{ data: row }">
                    <div v-if="row.campaign">
                      <router-link :to="{ name: 'bounces', query: { campaign_id: row.campaign.id } }">
                        {{ row.campaign.name }}
                      </router-link>
                    </div>
                  </template>
                </PvColumn>

                <PvColumn field="createdAt" :header="$t('globals.fields.createdAt')">
                  <template #body="{ data: row }">
                    {{ $utils.niceDate(row.createdAt, true) }}
                  </template>
                </PvColumn>

                <PvColumn field="action" :header="$t('globals.fields.type')">
                  <template #body="{ data: row }">
                    <span class="is-pulled-right">
                      <a href="#" @click.prevent="toggleMeta(row.id)">
                        {{ row.source }}
                        <i :class="visibleMeta[row.id] ? 'pi pi-chevron-up' : 'pi pi-chevron-down'" />
                      </a>
                    </span>
                    <span class="is-clearfix" />
                    <pre v-if="visibleMeta[row.id]">{{ row.meta }}</pre>
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

        <div class="field mt-6">
          <small class="block mt-1 text-color-secondary">{{ $t('subscribers.attribsHelp') + ' ' + egAttribs }}</small>
          <div>
            <h5>{{ $t('globals.terms.attribs') }}</h5>
            <PvTextarea v-model="form.strAttribs" name="attribs" rows="4" />
            <a href="https://listmonk.app/docs/concepts" target="_blank" rel="noopener noreferrer" class="is-size-7">
              {{ $t('globals.buttons.learnMore') }} <i class="pi pi-external-link" />
            </a>
          </div>
        </div>
      </section>
      <footer class="modal-card-foot has-text-right">
        <PvButton @click="$parent.close()" :label="$t('globals.buttons.close')" />
        <PvButton v-if="$can('subscribers:manage')" type="submit" severity="primary"
          :loading="loading.subscribers" :label="$t('globals.buttons.save')" />
      </footer>
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

  data() {
    return {
      // Binds form input values. This is populated by subscriber props passed
      // from the parent component in mounted().
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
    toggleBounces() {
      this.isBounceVisible = !this.isBounceVisible;
    },

    toggleMeta(id) {
      let v = false;
      if (!this.visibleMeta[id]) {
        v = true;
      }
      this.visibleMeta[id] = v;
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

        // List IDs.
        lists: this.form.lists.map((l) => l.id),
      };

      this.$api.createSubscriber(data).then((d) => {
        this.$emit('finished');
        this.$parent.close();
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

        // List IDs.
        lists: this.form.lists.map((l) => l.id),
      };

      this.$api.updateSubscriber(data).then((d) => {
        this.$emit('finished');
        this.$parent.close();
        this.$utils.toast(this.$t('globals.messages.updated', { name: d.name }));
      });
    },

    sendOptinConfirmation() {
      this.$api.sendSubscriberOptin(this.form.id).then(() => {
        this.$utils.toast(this.$t('subscribers.sentOptinConfirm'));
      });
    },

    validateAttribs(str) {
      // Parse and validate attributes JSON.
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

        // Deep-copy the lists array on to the form.
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
