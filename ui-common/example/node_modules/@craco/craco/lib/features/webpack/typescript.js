const { log } = require("../../logger");

function disableTypeChecking(webpackConfig) {
    webpackConfig.plugins = webpackConfig.plugins.filter(
        plugin => plugin.constructor.name !== "ForkTsCheckerWebpackPlugin"
    );

    log("Disabled TypeScript type checking.");

    return webpackConfig;
}

function overrideTypeScript(cracoConfig, webpackConfig) {
    if (cracoConfig.typescript) {
        const { enableTypeChecking } = cracoConfig.typescript;

        if (enableTypeChecking === false) {
            disableTypeChecking(webpackConfig);
        }
    }

    return webpackConfig;
}

module.exports = {
    overrideTypeScript
};
