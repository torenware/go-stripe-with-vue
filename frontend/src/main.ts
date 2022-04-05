import { createApp, App as BaseApp } from 'vue';
import 'bootstrap';
import 'bootstrap/dist/css/bootstrap.min.css';
import '@popperjs/core';
import '@stripe/stripe-js';

import App from './App.vue';
import Login from './blocks/LoginBlock.vue';
import FlashPanel from './components/FlashPanel.vue';
import TableTest from './blocks/TableTest.vue';
import SubscribersTable from './blocks/SubscribersTable.vue';
import NewSubscriber from './blocks/NewSubscriber.vue';

type EPData = {
  app: BaseApp;
  props?: Record<string, string>;
};
type EPLookup = Record<string, EPData>;

const keys: EPLookup = {};
keys['app'] = { app: App as unknown as BaseApp };
keys['login'] = { app: Login as unknown as BaseApp };
keys['subs'] = { app: SubscribersTable as unknown as BaseApp };
keys['new-sub'] = { app: NewSubscriber as unknown as BaseApp };
keys['table'] = { app: TableTest as unknown as BaseApp };
keys['flash'] = {
  app: FlashPanel as unknown as BaseApp,
  props: { id: 'flashPanel' },
};

const appMounts = document.querySelectorAll('[data-entryp]');
appMounts.forEach((mp) => {
  const ep = mp.getAttribute('data-entryp');
  console.log('loading ', ep);
  if (ep && ep in keys) {
    let { app, props } = keys[ep];
    props = props ? props : {};
    createApp(app, props).mount(mp);
  } else {
    console.log(`${ep}: key was not found.`);
  }
});
