const MiniCssExtractPlugin = require("mini-css-extract-plugin");

module.exports = {
    webpack: {
        configure: (webpackConfig) => {
            const instanceOfMiniCssExtractPlugin = webpackConfig.plugins.find(
                (plugin) => plugin instanceof MiniCssExtractPlugin,
            );
            instanceOfMiniCssExtractPlugin.options.ignoreOrder = true;

            return webpackConfig;
        },
    }
}
