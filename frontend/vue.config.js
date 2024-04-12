const { defineConfig } = require('@vue/cli-service')
module.exports = defineConfig({
  lintOnSave: false,
  publicPath: process.env.NODE_ENV === 'production' ? 'dist' : '',
  devServer: {
    historyApiFallback: true,
    allowedHosts: "all",
  }
})
