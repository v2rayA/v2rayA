import Vue from "vue";
import VueRouter from "vue-router";
import store from "@/store";

Vue.use(VueRouter);

const routes = [
  {
    path: "/",
    redirect: "/status"
  },
  {
    path: "/status",
    component: () => import("@/views/status")
  },
  {
    path: "/about",
    name: "about",
    // route level code-splitting
    // this generates a separate chunk (about.[hash].js) for this route
    // which is lazy-loaded when the route is visited.
    component: () =>
      import(/* webpackChunkName: "about" */ "../views/About.vue")
  }
];

const router = new VueRouter({
  mode: "history",
  routes
});

const title = "V2RayA";
router.afterEach(to => {
  if (to.meta.title) {
    document.title = `${to.meta.title} - ${title}`;
  } else {
    document.title = `${title}`;
  }
  store.commit("NAV", to.path.split("/")[1]);
});

export default router;
