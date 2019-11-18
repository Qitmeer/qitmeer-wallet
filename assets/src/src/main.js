import Vue from 'vue';
import ElementUI from 'element-ui';
import axios from 'axios';

import 'element-ui/lib/theme-chalk/index.css';
import './assets/iconfont/iconfont.css'
import './assets/main.css';
import App from './App.vue';
import router from './router';
import store from './store'


Vue.use(ElementUI);
Vue.prototype.$axios = axios.create({
  baseURL: window.QitmeerConfig.RPCAddr,
  timeout: 5000,
  headers: { "content-type": "application/json" },
  auth: {
    username: window.QitmeerConfig.RPCUser,
    password: window.QitmeerConfig.RPCPass
  },
  withCredentials: true,
  crossDomain: true,

});


new Vue({
  el: '#app',
  router,
  store,
  render: h => h(App)
});