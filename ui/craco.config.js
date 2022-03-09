// this workaround fix a warning of mini-css-extract-plugin throws "Conflicting order" during build
// https://github.com/facebook/create-react-app/issues/5372
// const {CracoAliasPlugin, configPaths, aliasDangerous} = require('react-app-rewire-alias')

// const aliasMap = configPaths('./tsconfig.paths.json')
// const path = require('path');

// module.exports = function override(config) {
//     aliasDangerous({
//       ...configPaths('tsconfig.paths.json')
//     })(config)
  
//     return config
//   }
module.exports = {
    webpack: {
        configure: (webpackConfig) => {
            const instanceOfMiniCssExtractPlugin = webpackConfig.plugins.find(
                (plugin) => plugin.options && plugin.options.ignoreOrder != null,
            );
            if(instanceOfMiniCssExtractPlugin)
                instanceOfMiniCssExtractPlugin.options.ignoreOrder = true;

            //webpackConfig.resolve.alias['react']= path.resolve(__dirname, 'node_modules/react'); // solve 2  react instances

            return webpackConfig;
        }
    }
}
// const path = require("path");
// const enableImportsFromExternalPaths = require("./enableImportsFromExternalPaths");

// Paths to the code you want to use
//const sharedLibOne = path.resolve(__dirname, "../../traffic-viewer/src/components/Pages/TrafficPage/TrafficPage");

// module.exports = {
//     plugins: [
//         {
//             plugin: {
//                 overrideWebpackConfig: ({ webpackConfig }) => {
//                     enableImportsFromExternalPaths(webpackConfig, [
//                         // Add the paths here
//                         sharedLibOne
//                     ]);

//                     webpackConfig.resolve.plugins.forEach(plugin => {
//                         if (plugin instanceof ModuleScopePlugin) {
//                           plugin.allowedFiles.add(path.resolve("./config.json"));
//                         })
//                     return webpackConfig;
//                 },
//             },
//         },
//     ],
// };

// const ModuleScopePlugin = require("react-dev-utils/ModuleScopePlugin");
// const path = require("path");

// module.exports = {
//   webpack: {
//     configure: webpackConfig => {
//       webpackConfig.resolve.plugins.forEach(plugin => {
//         if (plugin instanceof ModuleScopePlugin) {
//           plugin.allowedFiles.add(path.resolve("./config.json"));
//         }
//       });
//       return webpackConfig;
//     }
//   }
// };