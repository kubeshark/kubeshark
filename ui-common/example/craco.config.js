//const stylesResourcesLoader = require('craco-style-resources-loader');
module.exports = {
  webpack: {
      configure: (webpackConfig) => {
          const instanceOfMiniCssExtractPlugin = webpackConfig.plugins.find(
              (plugin) => plugin.options && plugin.options.ignoreOrder != null,
          );
          if(instanceOfMiniCssExtractPlugin)
              instanceOfMiniCssExtractPlugin.options.ignoreOrder = true;

          return webpackConfig;
      }
  },
  // plugins: [
  //   {
  //     plugin: stylesResourcesLoader,
  //     options: {
  //       patterns: ['./node_modules/liraz-test/dist/index.css','./node_modules/liraz-test/dist/*.svg'],
  //     },
  //   },
  // ]
}