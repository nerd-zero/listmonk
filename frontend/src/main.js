import { createApp } from 'vue';
import { createI18n } from 'vue-i18n';
import PrimeVue from 'primevue/config';
import Aura from '@primeuix/themes/aura';
import ToastService from 'primevue/toastservice';
import ConfirmationService from 'primevue/confirmationservice';
import 'primeflex/primeflex.css';

// PrimeVue components registered globally (mirrors Buefy's global install).
import Button from 'primevue/button';
import InputText from 'primevue/inputtext';
import Textarea from 'primevue/textarea';
import Select from 'primevue/select';
import ToggleSwitch from 'primevue/toggleswitch';
import Tag from 'primevue/tag';
import DataTable from 'primevue/datatable';
import Column from 'primevue/column';
import Dialog from 'primevue/dialog';
import ProgressBar from 'primevue/progressbar';
import ProgressSpinner from 'primevue/progressspinner';
import Tabs from 'primevue/tabs';
import TabList from 'primevue/tablist';
import Tab from 'primevue/tab';
import TabPanels from 'primevue/tabpanels';
import TabPanel from 'primevue/tabpanel';
import Toast from 'primevue/toast';
import ConfirmDialog from 'primevue/confirmdialog';
import Tooltip from 'primevue/tooltip';
import Badge from 'primevue/badge';
import Chip from 'primevue/chip';
import InputNumber from 'primevue/inputnumber';
import Password from 'primevue/password';
import Checkbox from 'primevue/checkbox';
import RadioButton from 'primevue/radiobutton';
import Paginator from 'primevue/paginator';
import Menu from 'primevue/menu';
import Menubar from 'primevue/menubar';
import PanelMenu from 'primevue/panelmenu';
import Drawer from 'primevue/drawer';
import Message from 'primevue/message';
import InlineMessage from 'primevue/inlinemessage';
import FloatLabel from 'primevue/floatlabel';
import AutoComplete from 'primevue/autocomplete';
import MultiSelect from 'primevue/multiselect';
import DatePicker from 'primevue/datepicker';
import Divider from 'primevue/divider';
import Panel from 'primevue/panel';
import Card from 'primevue/card';
import Avatar from 'primevue/avatar';
import IconField from 'primevue/iconfield';
import InputIcon from 'primevue/inputicon';

import App from './App.vue';
import router from './router';
import store from './store';
import * as api from './api';
import Utils from './utils';

const i18n = createI18n({
  legacy: true,
  locale: 'en',
  messages: {},
});

const app = createApp(App);

app.use(router);
app.use(store);
app.use(i18n);

app.use(PrimeVue, {
  theme: {
    preset: Aura,
    options: {
      darkModeSelector: '.app-dark',
    },
  },
  ripple: true,
});
app.use(ToastService);
app.use(ConfirmationService);

// Register PrimeVue components globally.
app.component('PvButton', Button);
app.component('PvInputText', InputText);
app.component('PvTextarea', Textarea);
app.component('PvSelect', Select);
app.component('PvToggleSwitch', ToggleSwitch);
app.component('PvTag', Tag);
app.component('PvDataTable', DataTable);
app.component('PvColumn', Column);
app.component('PvDialog', Dialog);
app.component('PvProgressBar', ProgressBar);
app.component('PvProgressSpinner', ProgressSpinner);
app.component('PvTabs', Tabs);
app.component('PvTabList', TabList);
app.component('PvTab', Tab);
app.component('PvTabPanels', TabPanels);
app.component('PvTabPanel', TabPanel);
app.component('PvToast', Toast);
app.component('PvConfirmDialog', ConfirmDialog);
app.component('PvBadge', Badge);
app.component('PvChip', Chip);
app.component('PvInputNumber', InputNumber);
app.component('PvPassword', Password);
app.component('PvCheckbox', Checkbox);
app.component('PvRadioButton', RadioButton);
app.component('PvPaginator', Paginator);
app.component('PvMenu', Menu);
app.component('PvMenubar', Menubar);
app.component('PvPanelMenu', PanelMenu);
app.component('PvDrawer', Drawer);
app.component('PvMessage', Message);
app.component('PvInlineMessage', InlineMessage);
app.component('PvFloatLabel', FloatLabel);
app.component('PvAutoComplete', AutoComplete);
app.component('PvMultiSelect', MultiSelect);
app.component('PvDatePicker', DatePicker);
app.component('PvDivider', Divider);
app.component('PvPanel', Panel);
app.component('PvCard', Card);
app.component('PvAvatar', Avatar);
app.component('PvIconField', IconField);
app.component('PvInputIcon', InputIcon);

app.directive('tooltip', Tooltip);

// Setup the router lifecycle hooks.
router.beforeEach((to, from, next) => {
  if (to.matched.length === 0) {
    next('/404');
  } else {
    next();
  }
});

router.afterEach((to) => {
  const { te, tc } = i18n.global;
  const title = to.meta.title && te(to.meta.title) ? `${tc(to.meta.title, 0)} /` : '';
  document.title = `${title} listmonk`;
});

async function initConfig(instance) {
  const [profile, cfg] = await Promise.all([api.getUserProfile(), api.getServerConfig()]);

  const lang = await api.getLang(cfg.lang);
  i18n.global.locale = cfg.lang;
  i18n.global.setLocaleMessage(cfg.lang, lang);

  const props = instance.config.globalProperties;
  props.$utils = new Utils(i18n.global);
  props.$api = api;

  props.$can = (...perms) => {
    if (profile.userRole.id === 1) {
      return true;
    }
    return perms.some((perm) => {
      if (perm.endsWith('*')) {
        const group = `${perm.split(':')[0]}:`;
        return profile.userRole.permissions.some((p) => p.startsWith(group));
      }
      return profile.userRole.permissions.includes(perm);
    });
  };

  props.$canList = (id, perm) => {
    if (profile.userRole.id === 1) {
      return true;
    }
    const can = props.$can('lists:get_all', 'lists:manage_all');
    if (can) {
      return true;
    }
    return profile.listRole.lists.some((list) => list.id === id && list.permissions.includes(perm));
  };

  const currentRoute = router.currentRoute.value;
  const routeTitle = currentRoute.meta.title ? `${i18n.global.tc(currentRoute.meta.title, 0)} /` : '';
  document.title = `${routeTitle} listmonk`;

  instance.mount('#app');
}

initConfig(app);
