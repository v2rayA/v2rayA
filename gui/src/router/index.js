import Vue from "vue";
import VueRouter from "vue-router";
import store from "@/store";

Vue.use(VueRouter);

const routes = [
  {
    path: "/",
    redirect: "/node"
  },
  {
    path: "/node",
    component: () => import("@/views/node.vue")
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
