const { mergeWebpackConfig } = require("./merge-webpack-config");
const {
    loadWebpackProdConfig,
    overrideWebpackProdConfig,
    loadWebpackDevConfig,
    overrideWebpackDevConfig
} = require("../../cra");

function overrideWebpackDev(cracoConfig, context) {
    const craWebpackConfig = loadWebpackDevConfig(cracoConfig);
    const resultingWebpackConfig = mergeWebpackConfig(cracoConfig, craWebpackConfig, context);

    overrideWebpackDevConfig(cracoConfig, resultingWebpackConfig);
}

function overrideWebpackProd(cracoConfig, context) {
    const craWebpackConfig = loadWebpackProdConfig(cracoConfig);
    const resultingWebpackConfig = mergeWebpackConfig(cracoConfig, craWebpackConfig, context);

    overrideWebpackProdConfig(cracoConfig, resultingWebpackConfig);
}

module.exports = {
    overrideWebpackDev,
    overrideWebpackProd
};
