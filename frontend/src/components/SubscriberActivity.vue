<template>
  <div class="subscriber-activity">
    <div v-if="isLoading" class="has-text-centered">
      <PvProgressSpinner style="width:2rem;height:2rem" />
    </div>

    <div v-else>
      <!-- Summary Stats -->
      <div class="grid">
        <div class="col-4">
          <div class="box has-text-centered">
            <p class="heading">{{ $t('globals.terms.campaigns') }}</p>
            <p class="title">{{ activity.campaignViews ? activity.campaignViews.length : 0 }}</p>
          </div>
        </div>
        <div class="col-4">
          <div class="box has-text-centered">
            <p class="heading">{{ $t('campaigns.views') }}</p>
            <p class="title">{{ totalViews }}</p>
          </div>
        </div>
        <div class="col-4">
          <div class="box has-text-centered">
            <p class="heading">{{ $t('campaigns.clicks') }}</p>
            <p class="title">{{ totalClicks }}</p>
          </div>
        </div>
      </div>

      <!-- Campaign Views Section -->
      <div class="section-header mb-4">
        <h5 class="title is-5">
          {{ $t('campaigns.views') }}
        </h5>
      </div>

      <div v-if="activity.campaignViews && activity.campaignViews.length > 0">
        <PvDataTable :value="activity.campaignViews" :hoverable="true" sort-field="lastViewedAt" :sort-order="-1"
          :paginator="true" :rows="10" class="campaign-views-table">
          <PvColumn field="subject" :header="$tc('globals.terms.campaign', 1)" sortable>
            <template #body="{ data }">
              <div v-if="data.uuid">
                <router-link :to="{ name: 'campaign', params: { id: data.id } }">
                  {{ data.name }}
                </router-link>
                <p class="is-size-7 has-text-grey">{{ data.subject }}</p>
              </div>
              <div v-else>
                <em class="has-text-grey">{{ $t('subscribers.activity.campaignDeleted') }}</em>
              </div>
            </template>
          </PvColumn>

          <PvColumn field="viewCount" :header="$t('campaigns.views')" sortable>
            <template #body="{ data }">
              <span class="tag is-light">{{ data.viewCount }}</span>
            </template>
          </PvColumn>

          <PvColumn field="lastViewedAt" :header="$t('globals.fields.createdAt')" sortable>
            <template #body="{ data }">
              <span v-if="data.lastViewedAt">
                {{ $utils.niceDate(data.lastViewedAt, true) }}
              </span>
            </template>
          </PvColumn>
        </PvDataTable>
      </div>
      <div v-else class="has-text-centered has-text-grey p-6">
        <p class="mt-2">{{ $t('globals.messages.emptyState') }}</p>
      </div>

      <!-- Link Clicks Section -->
      <div class="section-header mb-4 mt-6">
        <h5 class="title is-5">
          {{ $t('campaigns.clicks') }}
        </h5>
      </div>

      <div v-if="activity.linkClicks && activity.linkClicks.length > 0">
        <PvDataTable :value="activity.linkClicks" :hoverable="true" sort-field="lastClickedAt" :sort-order="-1"
          :paginator="true" :rows="10" class="link-clicks-table">
          <PvColumn field="url" :header="$t('globals.terms.url')" sortable body-class="link-click-url">
            <template #body="{ data }">
              <a :href="data.url" target="_blank" rel="noopener noreferrer">
                {{ data.url }}
              </a>
            </template>
          </PvColumn>

          <PvColumn field="campaignName" :header="$tc('globals.terms.campaign', 1)" sortable>
            <template #body="{ data }">
              <div v-if="data.campaignUuid">
                <router-link :to="{ name: 'campaign', params: { id: data.campaignId } }">
                  {{ data.campaignSubject || data.campaignName }}
                </router-link>
              </div>
              <div v-else>
                &mdash;
              </div>
            </template>
          </PvColumn>

          <PvColumn field="clickCount" :header="$t('campaigns.clicks')" sortable>
            <template #body="{ data }">
              <span class="tag is-light">{{ data.clickCount }}</span>
            </template>
          </PvColumn>

          <PvColumn field="lastClickedAt" :header="$t('globals.fields.createdAt')" sortable>
            <template #body="{ data }">
              <span v-if="data.lastClickedAt">
                {{ $utils.niceDate(data.lastClickedAt, true) }}
              </span>
            </template>
          </PvColumn>
        </PvDataTable>
      </div>
      <div v-else class="has-text-centered has-text-grey p-6">
        <p class="mt-2">{{ $t('globals.messages.emptyState') }}</p>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';

export default defineComponent({
  props: {
    subscriberId: {
      type: Number,
      required: true,
    },
  },

  data() {
    return {
      isLoading: false,
      activity: {
        campaignViews: [],
        linkClicks: [],
      },
    };
  },

  computed: {
    totalViews() {
      if (!this.activity.campaignViews) return 0;
      return this.activity.campaignViews.reduce((sum, v) => sum + (v.viewCount || 0), 0);
    },

    totalClicks() {
      if (!this.activity.linkClicks) return 0;
      return this.activity.linkClicks.reduce((sum, c) => sum + (c.clickCount || 0), 0);
    },
  },

  mounted() {
    this.getActivity();
  },

  methods: {
    getActivity() {
      this.isLoading = true;
      this.$api.getSubscriberActivity(this.subscriberId).then((data) => {
        this.activity = data;
        this.isLoading = false;
      }).catch(() => {
        this.isLoading = false;
      });
    },
  },
});
</script>
