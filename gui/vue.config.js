var webpack = require("webpack");
var path = require("path");
var WebpackIconfontPluginNodejs = require("webpack-iconfont-plugin-nodejs");
var dir = "src/assets/iconfont";
var TerserPlugin = require("terser-webpack-plugin");

module.exports = {
  configureWebpack: (config) => {
    config.resolve.alias["vue$"] = "vue/dist/vue.esm.js";
    return {
      optimization: {
        minimizer: [
          new TerserPlugin({
            minify: TerserPlugin.swcMinify,
            terserOptions: {
              compress: true,
            },
          }),
        ],
      },
      plugins: [
        new webpack.DefinePlugin({
          apiRoot: '`${localStorage["backendAddress"]}/api`',
        }),
        new WebpackIconfontPluginNodejs({
          cssPrefix: "icon",
          svgs: path.join(dir, "svgs/*.svg"),
          template: path.join(dir, "css-template.njk"),
          fontsOutput: path.join(dir, "fonts/"),
          cssOutput: path.join(dir, "fonts/font.css"),
          // htmlOutput: path.join(dir, "fonts/_font-preview.html"),
          jsOutput: path.join(dir, "fonts/fonts.js"),
          // formats: ['ttf', 'woff', 'svg'],
        }),
      ],
    };
  },

  productionSourceMap: false,

  devServer: {
    port: 8081,
  },

  // publicPath: "./static/",
  publicPath:
    process.env.publicPath ||
    (process.env.NODE_ENV === "production" ? "./static/" : "/"),
  outputDir: process.env.OUTPUT_DIR || "../web",

  // pwa: {
  //   name: "v2rayA",
  //   themeColor: "#FFDD57",
  //   msTileColor: "#fff",
  //   appleMobileWebAppCapable: "yes",
  //   appleMobileWebAppStatusBarStyle: "white",
  //   workboxOptions: {
  //     skipWaiting: true
  //   }
  // },

  lintOnSave: false,

  transpileDependencies: ["buefy"],
};
