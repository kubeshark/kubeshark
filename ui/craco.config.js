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


            return webpackConfig;
        }
    }
}