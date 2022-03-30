import { createApp, App } from 'vue';
import 'bootstrap';
import 'bootstrap/dist/css/bootstrap.min.css';
// import App from './App.vue';
import Login from './blocks/LoginBlock.vue';
import FlashPanel from './components/FlashPanel.vue';

type EPData = {
  app: App;
  props?: Record<string, string>;
};
type EPLookup = Record<string, EPData>;

const keys: EPLookup = {};
//keys['app'] = { app: App };
keys['login'] = { app: Login as unknown as App };
keys['flash'] = {
  app: FlashPanel as unknown as App,
  props: { id: 'flashPanel' },
};

console.log('main ran');

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
