<template>
  <div class="app-shell" :class="{ 'sidebar-collapsed': sidebarCollapsed, 'app-dark': isDark }">
    <PvToast position="bottom-right" />
    <PvConfirmDialog />

    <template v-if="isLoaded">
      <!-- ─── Sidebar ─── -->
      <aside class="app-sidebar">
        <div class="sidebar-header">
          <router-link :to="{ name: 'dashboard' }" class="sidebar-brand">
            <img src="@/assets/logo.svg" alt="listmonk" class="sidebar-logo" />
            <span class="sidebar-brand-name">listmonk</span>
          </router-link>
        </div>

        <nav class="sidebar-nav">
          <navigation :active-item="activeItem" />
        </nav>

        <div class="sidebar-footer">
          <PvAvatar
            v-if="profile.username"
            :label="profile.username[0].toUpperCase()"
            class="sidebar-avatar"
            shape="circle"
          />
          <span v-if="profile.username" class="sidebar-username">{{ profile.username }}</span>
          <button type="button" class="sidebar-logout" @click="doLogout" v-tooltip.top="$t('users.logout')">
            <i class="pi pi-sign-out" />
          </button>
        </div>
      </aside>

      <!-- ─── Main ─── -->
      <div class="app-main">
        <!-- Topbar -->
        <header class="app-topbar">
          <div class="topbar-start">
            <button type="button" class="topbar-btn" @click="sidebarCollapsed = !sidebarCollapsed">
              <i class="pi pi-bars" />
            </button>
          </div>

          <div class="topbar-end">
            <PvButton
              v-if="serverConfig.needs_restart"
              severity="danger"
              size="small"
              icon="pi pi-exclamation-triangle"
              :label="$t('settings.needsRestart')"
              @click="$utils.confirm($t('settings.confirmRestart'), reloadApp)"
            />

            <button
              type="button"
              class="topbar-btn"
              v-tooltip.bottom="$t('globals.buttons.refresh')"
              data-cy="btn-refresh"
              @click="triggerRefresh"
            >
              <i class="pi pi-refresh" />
            </button>

            <button type="button" class="topbar-btn" @click="toggleDark" v-tooltip.bottom="isDark ? 'Light mode' : 'Dark mode'">
              <i :class="['pi', isDark ? 'pi-sun' : 'pi-moon']" />
            </button>

            <button
              v-if="profile.username"
              type="button"
              class="topbar-user-btn"
              @click="(e) => $refs.userMenu.show(e)"
            >
              <PvAvatar
                :label="profile.username[0].toUpperCase()"
                class="topbar-avatar"
                shape="circle"
                size="small"
              />
              <span class="topbar-username">{{ profile.username }}</span>
              <i class="pi pi-chevron-down topbar-chevron" />
            </button>
            <PvMenu ref="userMenu" :model="userMenuItems" popup />
          </div>
        </header>

        <!-- System notices -->
        <template v-if="serverConfig.update">
          <div
            v-if="serverConfig.update.update && serverConfig.update.update.is_new"
            class="app-notice app-notice--success"
          >
            <i class="pi pi-arrow-up-right" />
            <span>
              {{ $t('settings.updateAvailable', { version: serverConfig.update.update.release_version }) }}
            </span>
            <a :href="serverConfig.update.update.url" target="_blank" rel="noopener noreferrer">View release</a>
          </div>
          <div
            v-for="m in (serverConfig.update.messages || [])"
            :key="m.title"
            class="app-notice"
            :class="m.priority === 'high' ? 'app-notice--danger' : 'app-notice--info'"
          >
            <i class="pi pi-info-circle" />
            <strong v-if="m.title">{{ m.title }}</strong>
            <span>{{ m.description }}</span>
            <a v-if="m.url" :href="m.url" target="_blank" rel="noopener noreferrer">View</a>
          </div>
        </template>
        <div v-if="serverConfig.has_legacy_user" class="app-notice app-notice--danger">
          <i class="pi pi-exclamation-triangle" />
          Remove <code>admin_username</code> / <code>admin_password</code> from config.
          <router-link :to="{ name: 'users' }">Go to Users</router-link>
        </div>

        <!-- Page content -->
        <div class="app-content">
          <router-view :key="$route.fullPath" />
        </div>
      </div>

      <!-- Mobile overlay -->
      <div class="sidebar-overlay" @click="sidebarCollapsed = true" />
    </template>

    <div v-if="!isLoaded" class="app-loading">
      <PvProgressSpinner style="width:48px;height:48px" stroke-width="3" />
    </div>
  </div>
</template>

<script>
import { mapState } from 'pinia';
import { useToast } from 'primevue/usetoast';
import { useConfirm } from 'primevue/useconfirm';
import { uris } from './constants';
import { useMainStore } from './store';
import { setToastInstance, setConfirmInstance } from './toastService';
import Navigation from './components/Navigation.vue';

export default {
  name: 'App',
  components: { Navigation },

  setup() {
    setToastInstance(useToast());
    setConfirmInstance(useConfirm());
  },

  data() {
    return {
      isLoaded: false,
      isDark: false,
      sidebarCollapsed: window.innerWidth < 992,
      activeItem: {},
    };
  },

  computed: {
    ...mapState(useMainStore, ['serverConfig', 'profile']),

    userMenuItems() {
      return [
        {
          label: this.$t('users.profile'),
          icon: 'pi pi-user',
          command: () => this.$router.push('/user/profile'),
        },
        { separator: true },
        {
          label: this.$t('users.logout'),
          icon: 'pi pi-sign-out',
          command: () => this.doLogout(),
        },
      ];
    },
  },

  watch: {
    $route(to) {
      this.activeItem = { [to.name]: true };
      if (window.innerWidth < 992) this.sidebarCollapsed = true;
    },
  },

  methods: {
    toggleDark() {
      this.isDark = !this.isDark;
      document.documentElement.classList.toggle('app-dark', this.isDark);
    },

    triggerRefresh() {
      useMainStore().refresh();
    },

    doLogout() {
      this.$api.logout().then(() => { document.location.href = uris.root; });
    },

    reloadApp() {
      this.$api.reloadApp().then(() => {
        this.$utils.toast('Reloading…');
        const poll = setInterval(() => {
          this.$api.getHealth().then(() => { clearInterval(poll); document.location.reload(); });
        }, 500);
      });
    },

    listenEvents() {
      const re = /(.+?)\.go:\d+:(.+?)$/im;
      const src = new EventSource(uris.errorEvents, { withCredentials: true });
      let n = 0;
      src.onmessage = (e) => {
        if (n > 50) return;
        n += 1;
        const d = JSON.parse(e.data);
        if (d?.type === 'error') {
          const m = re.exec(d.message.trim());
          if (m) this.$utils.toast(m[2], 'is-danger', null, true);
        }
      };
    },
  },

  mounted() {
    this.isLoaded = true;
    this.$api.getLists({ minimal: true, per_page: 'all', status: 'active' });
    this.listenEvents();
    this.activeItem = { [this.$route.name]: true };
    window.addEventListener('resize', () => {
      if (window.innerWidth >= 992) this.sidebarCollapsed = false;
    });
  },
};
</script>

<style lang="scss">
@import "assets/style.scss";
@import "assets/icons/fontello.css";
</style>
