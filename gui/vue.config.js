var webpack = require("webpack");

module.exports = {
  configureWebpack: config => {
    config.resolve.alias["vue$"] = "vue/dist/vue.esm.js";
    return {
      plugins: [
        new webpack.DefinePlugin({
          apiRoot: '`${localStorage["backendAddress"]}/api`'
        })
      ]
    };
  },

  productionSourceMap: false,

  devServer: {
    port: 8081
  },

  // publicPath:process.env.NODE_ENV === 'production'
  // ? '/V2RayA/'
  // : '/',
  outputDir: "../web",

  pwa: {
    name: "V2RayA",
    themeColor: "#FFDD57",
    msTileColor: "#000000",
    appleMobileWebAppCapable: "yes",
    appleMobileWebAppStatusBarStyle: "white",
    workboxOptions: {
      skipWaiting: true
    }
  },

  lintOnSave: false
};
