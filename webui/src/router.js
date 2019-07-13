import Vue from "vue";
import Router from "vue-router";
import WIP from "./views/WIP.vue";

Vue.use(Router);

export default new Router({
  routes: [
    {
      path: "/",
      name: "home",
      component: WIP
    }
  ]
});
