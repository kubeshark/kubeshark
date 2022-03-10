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
    }
}