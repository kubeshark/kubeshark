// this workaround fix a warning of mini-css-extract-plugin throws "Conflicting order" during build
// https://github.com/facebook/create-react-app/issues/5372

const path = require("path")

module.exports = {
    webpack: {
        configure: (webpackConfig) => {
            const instanceOfMiniCssExtractPlugin = webpackConfig.plugins.find(
                (plugin) => plugin.options && plugin.options.ignoreOrder != null,
            );
            if(instanceOfMiniCssExtractPlugin)
                instanceOfMiniCssExtractPlugin.options.ignoreOrder = true;

            webpackConfig.resolve.alias['react']= path.resolve(__dirname, 'node_modules/react'); // solve 2  react instances
            webpackConfig.resolve.alias['@material-ui/styles']= path.resolve("node_modules", "@material-ui/styles");

            return webpackConfig;
        }
    }
}