module.exports = {
  root: true,

  env: {
    node: true,
  },

  rules: {
    "no-console": "off",
    "no-debugger": "off",
  },

  parserOptions: {
    parser: "@babel/eslint-parser",
  },

  extends: ["plugin:vue/recommended", "@vue/prettier"],

  globals: {
    apiRoot: true,
  },
};
