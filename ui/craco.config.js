// this workaround fix a warning of mini-css-extract-plugin throws "Conflicting order" during build
// https://github.com/facebook/create-react-app/issues/5372
module.exports = {
    webpack: {
        configure: (webpackConfig) => {
            const instanceOfMiniCssExtractPlugin = webpackConfig.plugins.find(
                (plugin) => plugin.options && plugin.options.ignoreOrder != null,
            );
            if(instanceOfMiniCssExtractPlugin)
                instanceOfMiniCssExtractPlugin.options.ignoreOrder = true;

            return webpackConfig;
        },
    }
}
