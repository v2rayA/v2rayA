var webpack = require("webpack");

module.exports = {
  configureWebpack: config => {
    config.resolve.alias["vue$"]='vue/dist/vue.esm.js'
    if (process.env.NODE_ENV === "production") {
      // 为生产环境修改配置...
      return {
        plugins: [
          new webpack.DefinePlugin({
            apiRoot: "'http://localhost:2017/api'"
          })
        ]
      };
    } else {
      // 为开发环境修改配置...
      return {
        plugins: [
          new webpack.DefinePlugin({
            apiRoot: "'http://localhost:2017/api'"
          })
        ]
      };
    }
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
