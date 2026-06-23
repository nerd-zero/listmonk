<template>
  <nav>
    <ul>
      <li>
        <router-link :to="{ name: 'dashboard' }" :class="{ 'is-active': activeItem.dashboard }">
          <i class="pi pi-home" />
          <span>{{ $t('menu.dashboard') }}</span>
        </router-link>
      </li><!-- dashboard -->

      <li data-cy="lists">
        <a :class="{ 'is-active': activeGroup.lists }" @click="toggleGroup('lists', !activeGroup.lists)">
          <i class="pi pi-list" />
          <span>{{ $t('globals.terms.lists') }}</span>
        </a>
        <ul v-if="activeGroup.lists">
          <li>
            <router-link :to="{ name: 'lists' }" :class="{ 'is-active': activeItem.lists }" data-cy="all-lists">
              <i class="pi pi-list" />
              <span>{{ $t('menu.allLists') }}</span>
            </router-link>
          </li>
          <li class="forms">
            <router-link :to="{ name: 'forms' }" :class="{ 'is-active': activeItem.forms }">
              <i class="pi pi-file" />
              <span>{{ $t('menu.forms') }}</span>
            </router-link>
          </li>
        </ul>
      </li><!-- lists -->

      <li v-if="$can('subscribers:*')" data-cy="subscribers">
        <a :class="{ 'is-active': activeGroup.subscribers }" @click="toggleGroup('subscribers', !activeGroup.subscribers)">
          <i class="pi pi-users" />
          <span>{{ $t('globals.terms.subscribers') }}</span>
        </a>
        <ul v-if="activeGroup.subscribers">
          <li v-if="$can('subscribers:get_all', 'subscribers:get')">
            <router-link :to="{ name: 'subscribers' }" :class="{ 'is-active': activeItem.subscribers }" data-cy="all-subscribers">
              <i class="pi pi-users" />
              <span>{{ $t('menu.allSubscribers') }}</span>
            </router-link>
          </li>
          <li v-if="$can('subscribers:import')">
            <router-link :to="{ name: 'import' }" :class="{ 'is-active': activeItem.import }" data-cy="import">
              <i class="pi pi-upload" />
              <span>{{ $t('menu.import') }}</span>
            </router-link>
          </li>
          <li v-if="$can('bounces:get')">
            <router-link :to="{ name: 'bounces' }" :class="{ 'is-active': activeItem.bounces }" data-cy="bounces">
              <i class="pi pi-envelope" />
              <span>{{ $t('globals.terms.bounces') }}</span>
            </router-link>
          </li>
        </ul>
      </li><!-- subscribers -->

      <li v-if="$can('campaigns:*')" data-cy="campaigns">
        <a :class="{ 'is-active': activeGroup.campaigns }" @click="toggleGroup('campaigns', !activeGroup.campaigns)">
          <i class="pi pi-send" />
          <span>{{ $t('globals.terms.campaigns') }}</span>
        </a>
        <ul v-if="activeGroup.campaigns">
          <li v-if="$can('campaigns:get')">
            <router-link :to="{ name: 'campaigns' }" :class="{ 'is-active': activeItem.campaigns }" data-cy="all-campaigns">
              <i class="pi pi-send" />
              <span>{{ $t('menu.allCampaigns') }}</span>
            </router-link>
          </li>
          <li v-if="$can('campaigns:manage')">
            <router-link :to="{ name: 'campaign', params: { id: 'new' } }" :class="{ 'is-active': activeItem.campaign }" data-cy="new-campaign">
              <i class="pi pi-plus" />
              <span>{{ $t('menu.newCampaign') }}</span>
            </router-link>
          </li>
          <li v-if="$can('media:*')">
            <router-link :to="{ name: 'media' }" :class="{ 'is-active': activeItem.media }" data-cy="media">
              <i class="pi pi-image" />
              <span>{{ $t('menu.media') }}</span>
            </router-link>
          </li>
          <li v-if="$can('templates:get')">
            <router-link :to="{ name: 'templates' }" :class="{ 'is-active': activeItem.templates }" data-cy="templates">
              <i class="pi pi-file" />
              <span>{{ $t('globals.terms.templates') }}</span>
            </router-link>
          </li>
          <li v-if="$can('campaigns:get_analytics')">
            <router-link :to="{ name: 'campaignAnalytics' }" :class="{ 'is-active': activeItem.campaignAnalytics }" data-cy="analytics">
              <i class="pi pi-chart-line" />
              <span>{{ $t('globals.terms.analytics') }}</span>
            </router-link>
          </li>
        </ul>
      </li><!-- campaigns -->

      <li v-if="$can('users:*', 'roles:*')" data-cy="users">
        <a :class="{ 'is-active': activeGroup.users }" @click="toggleGroup('users', !activeGroup.users)">
          <i class="pi pi-users" />
          <span>{{ $t('globals.terms.users') }}</span>
        </a>
        <ul v-if="activeGroup.users">
          <li v-if="$can('users:get')">
            <router-link :to="{ name: 'users' }" :class="{ 'is-active': activeItem.users }" data-cy="users">
              <i class="pi pi-users" />
              <span>{{ $t('globals.terms.users') }}</span>
            </router-link>
          </li>
          <li v-if="$can('roles:get')">
            <router-link :to="{ name: 'userRoles' }" :class="{ 'is-active': activeItem.userRoles }" data-cy="userRoles">
              <i class="pi pi-file" />
              <span>{{ $t('users.userRoles') }}</span>
            </router-link>
          </li>
          <li v-if="$can('roles:get')">
            <router-link :to="{ name: 'listRoles' }" :class="{ 'is-active': activeItem.listRoles }" data-cy="listRoles">
              <i class="pi pi-list" />
              <span>{{ $t('users.listRoles') }}</span>
            </router-link>
          </li>
        </ul>
      </li><!-- users -->

      <li v-if="$can('settings:*')" data-cy="settings">
        <a :class="{ 'is-active': activeGroup.settings }" @click="toggleGroup('settings', !activeGroup.settings)">
          <i class="pi pi-cog" />
          <span>{{ $t('menu.settings') }}</span>
        </a>
        <ul v-if="activeGroup.settings">
          <li v-if="$can('settings:get')">
            <router-link :to="{ name: 'settings' }" :class="{ 'is-active': activeItem.settings }" data-cy="all-settings">
              <i class="pi pi-cog" />
              <span>{{ $t('menu.settings') }}</span>
            </router-link>
          </li>
          <li v-if="$can('settings:maintain')">
            <router-link :to="{ name: 'maintenance' }" :class="{ 'is-active': activeItem.maintenance }" data-cy="maintenance">
              <i class="pi pi-wrench" />
              <span>{{ $t('menu.maintenance') }}</span>
            </router-link>
          </li>
          <li v-if="$can('settings:get')">
            <router-link :to="{ name: 'logs' }" :class="{ 'is-active': activeItem.logs }" data-cy="logs">
              <i class="pi pi-list" />
              <span>{{ $t('menu.logs') }}</span>
            </router-link>
          </li>
        </ul>
      </li><!-- settings -->
    </ul>
  </nav>
</template>

<script>
import { mapState } from 'vuex';

export default {
  name: 'Navigation',

  props: {
    activeItem: { type: Object, default: () => { } },
    activeGroup: { type: Object, default: () => { } },
    isMobile: Boolean,
  },

  methods: {
    toggleGroup(group, state) {
      this.$emit('toggleGroup', group, state);
    },

    doLogout() {
      this.$emit('doLogout');
    },
  },

  computed: {
    ...mapState(['profile']),
  },

  mounted() {
    // A hack to close the open accordion burger menu items on click.
    // Buefy does not have a way to do this.
    if (this.isMobile) {
      document.querySelectorAll('.navbar li a[href]').forEach((e) => {
        e.onclick = () => {
          document.querySelector('.navbar-burger').click();
        };
      });
    }
  },
};

</script>
