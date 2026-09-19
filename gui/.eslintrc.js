module.exports = {
  root: true,

  env: {
    node: true,
  },

  rules: {
    "no-console": "off",
    "no-debugger": "off",
  },

  extends: ["plugin:vue/vue3-recommended", "@vue/prettier"],

  globals: {
    apiRoot: true,
  },
};
