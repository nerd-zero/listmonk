<template>
  <div id="app">
    <PvToast />
    <PvConfirmDialog />

    <nav class="navbar is-fixed-top" v-if="$root.isLoaded">
      <div class="navbar-brand">
        <div class="logo">
          <router-link :to="{ name: 'dashboard' }">
            <img class="full" src="@/assets/logo.svg" alt="" />
            <img class="favicon" src="@/assets/favicon.png" alt="" />
          </router-link>
        </div>
      </div>
      <div class="navbar-end">
        <navigation v-if="isMobile" :is-mobile="isMobile" :active-item="activeItem" :active-group="activeGroup"
          @toggle-group="toggleGroup" @do-logout="doLogout" />

        <a class="navbar-item" href="#" @click.prevent="emitPageRefresh" data-cy="btn-refresh"
          :aria-label="$t('globals.buttons.refresh')">
          <i class="pi pi-refresh" v-tooltip.bottom="$t('globals.buttons.refresh')" />
          <span class="is-hidden-tablet">{{ $t('globals.buttons.refresh') }}</span>
        </a>

        <div class="navbar-item has-dropdown is-hoverable user">
          <a class="navbar-link">
            <span class="user-avatar" v-if="profile.username">
              <img v-if="profile.avatar" :src="profile.avatar" alt="" />
              <span v-else>{{ profile.username[0].toUpperCase() }}</span>
            </span>
          </a>
          <div class="navbar-dropdown is-right">
            <router-link class="navbar-item user-name" to="/user/profile">
              <strong>{{ profile.username }}</strong>
              <div class="is-size-7">{{ profile.name }}</div>
            </router-link>

            <router-link class="navbar-item" to="/user/profile">
              <i class="pi pi-user" /> {{ $t('users.profile') }}
            </router-link>
            <a class="navbar-item" href="#" @click.prevent="doLogout">
              <i class="pi pi-sign-out" /> {{ $t('users.logout') }}
            </a>
          </div>
        </div>
      </div>
    </nav>

    <div class="wrapper" v-if="$root.isLoaded">
      <section class="sidebar">
        <aside class="sidebar-inner">
          <div>
            <nav>
              <navigation v-if="!isMobile" :is-mobile="isMobile" :active-item="activeItem" :active-group="activeGroup"
                @toggle-group="toggleGroup" />
            </nav>
          </div>
        </aside>
      </section>
      <!-- sidebar-->

      <!-- body //-->
      <div class="main">
        <div class="global-notices" v-if="isGlobalNotices">
          <div v-if="serverConfig.needs_restart" class="notification is-danger">
            {{ $t('settings.needsRestart') }}
            &mdash;
            <PvButton severity="primary" size="small"
              @click="$utils.confirm($t('settings.confirmRestart'), reloadApp)"
              :label="$t('settings.restart')" />
          </div>

          <template v-if="serverConfig.update">
            <div v-if="serverConfig.update.update.is_new" class="notification is-success">
              {{ $t('settings.updateAvailable', {
                version: `${serverConfig.update.update.release_version}
              (${$utils.getDate(serverConfig.update.update.release_date).format('DD MMM YY')})`,
              }) }}
              <a :href="serverConfig.update.update.url" target="_blank" rel="noopener noreferer">View</a>
            </div>

            <template v-if="serverConfig.update.messages && serverConfig.update.messages.length > 0">
              <div v-for="m in serverConfig.update.messages" class="notification"
                :class="{ [m.priority === 'high' ? 'is-danger' : 'is-info']: true }" :key="m.title">
                <h3 class="is-size-5" v-if="m.title"><strong>{{ m.title }}</strong></h3>
                <p v-if="m.description">{{ m.description }}</p>
                <a v-if="m.url" :href="m.url" target="_blank" rel="noopener noreferer">View</a>
              </div>
            </template>
          </template>

          <div v-if="serverConfig.has_legacy_user" class="notification is-danger">
            <i class="pi pi-exclamation-triangle" />
            Remove the <code>admin_username</code> and <code>admin_password</code> fields from the TOML
            configuration file or environment variables. If you are using APIs, create and use new API credentials
            before removing them. Visit
            <router-link :to="{ name: 'users' }">
              Admin -> Settings -> Users
            </router-link> dashboard. <a href="https://listmonk.app/docs/upgrade/#upgrading-to-v4xx" target="_blank"
              rel="noopener noreferer">Learn more.</a>
          </div>
        </div>

        <router-view :key="$route.fullPath" />
      </div>
    </div>

    <div v-if="!$root.isLoaded" class="flex justify-center p-8">
      <PvProgressSpinner />
    </div>
  </div>
</template>

<script>
import { mapState } from 'vuex';
import { useToast } from 'primevue/usetoast';
import { useConfirm } from 'primevue/useconfirm';
import { uris } from './constants';
import { setToastInstance, setConfirmInstance } from './toastService';

import Navigation from './components/Navigation.vue';

export default {
  name: 'App',

  components: {
    Navigation,
  },

  data() {
    return {
      activeItem: {},
      activeGroup: {},
      windowWidth: window.innerWidth,
    };
  },

  watch: {
    $route(to) {
      // Set the current route name to true for active+expanded keys in the
      // menu to pick up.
      this.activeItem = { [to.name]: true };
      if (to.meta.group) {
        this.activeGroup = { [to.meta.group]: true };
      } else {
        // Reset activeGroup to collapse menu items on navigating
        // to non group items from sidebar
        this.activeGroup = {};
      }
    },
  },

  methods: {
    toggleGroup(group, state) {
      this.activeGroup = state ? { [group]: true } : {};
    },

    emitPageRefresh() {
      this.$root.$emit('page.refresh');
    },

    reloadApp() {
      this.$api.reloadApp().then(() => {
        this.$utils.toast('Reloading app ...');

        // Poll until there's a 200 response, waiting for the app
        // to restart and come back up.
        const pollId = setInterval(() => {
          this.$api.getHealth().then(() => {
            clearInterval(pollId);
            document.location.reload();
          });
        }, 500);
      });
    },

    doLogout() {
      this.$api.logout().then(() => {
        document.location.href = uris.root;
      });
    },

    listenEvents() {
      const reMatchLog = /(.+?)\.go:\d+:(.+?)$/im;
      const evtSource = new EventSource(uris.errorEvents, { withCredentials: true });
      let numEv = 0;
      evtSource.onmessage = (e) => {
        if (numEv > 50) {
          return;
        }
        numEv += 1;

        const d = JSON.parse(e.data);
        if (d && d.type === 'error') {
          const msg = reMatchLog.exec(d.message.trim());
          this.$utils.toast(msg[2], 'is-danger', null, true);
        }
      };
    },
  },

  computed: {
    ...mapState(['serverConfig', 'profile']),

    isGlobalNotices() {
      return (this.serverConfig.needs_restart
        || this.serverConfig.has_legacy_user
        || (this.serverConfig.update
          && this.serverConfig.update.messages
          && this.serverConfig.update.messages.length > 0));
    },

    version() {
      return import.meta.env.VUE_APP_VERSION;
    },

    isMobile() {
      return this.windowWidth <= 768;
    },
  },

  setup() {
    setToastInstance(useToast());
    setConfirmInstance(useConfirm());
  },

  mounted() {
    // Lists is required across different views. On app load, fetch the lists
    // and have them in the store.
    this.$api.getLists({ minimal: true, per_page: 'all', status: 'active' });

    window.addEventListener('resize', () => {
      this.windowWidth = window.innerWidth;
    });

    this.listenEvents();
  },
};
</script>

<style lang="scss">
@import "assets/style.scss";
@import "assets/icons/fontello.css";
</style>
