import { createApp } from 'vue';
import 'bootstrap';
import "bootstrap/dist/css/bootstrap.min.css";
import App from './App.vue';
import Login from './blocks/LoginBlock.vue';

type EPLookup = Record<string, typeof App>;

const keys: EPLookup = {};
keys['app'] = App;
keys['login'] = Login;

console.log("main ran");

const appMounts = document.querySelectorAll('[data-entryp]');
appMounts.forEach((mp) => {
    const ep = mp.getAttribute('data-entryp');
    console.log("loading ", ep);
    if (ep && ep in keys) {
        createApp(keys[ep]).mount(mp);
    } else {
        console.log(`${ep}: key was not found.`);
    }
});
