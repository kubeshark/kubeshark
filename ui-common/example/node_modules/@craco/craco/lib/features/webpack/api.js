const { isFunction } = require("../../utils");
const { mergeWebpackConfig } = require("./merge-webpack-config");
const { setArgs } = require("../../args");
const { processCracoConfig } = require("../../config");
const { loadWebpackProdConfig, loadWebpackDevConfig, getCraPaths } = require("../../cra");

function createWebpackDevConfig(callerCracoConfig, callerContext, options) {
    return createWebpackConfig(callerCracoConfig, callerContext, loadWebpackDevConfig, "development", options);
}

function createWebpackProdConfig(callerCracoConfig, callerContext, options) {
    return createWebpackConfig(callerCracoConfig, callerContext, loadWebpackProdConfig, "production", options);
}

function createWebpackConfig(callerCracoConfig, callerContext = {}, loadWebpackConfig, env, options = {}) {
    if (!callerCracoConfig) {
        throw new Error("craco: 'cracoConfig' is required.");
    }
    if (isFunction(callerCracoConfig)) {
        throw new Error("craco: 'cracoConfig' should be an object.");
    }

    if (!process.env.NODE_ENV) {
        process.env.NODE_ENV = env;
    }

    setArgs(options);

    const context = {
        env: process.env.NODE_ENV,
        ...callerContext
    };

    const cracoConfig = processCracoConfig(callerCracoConfig, context);
    context.paths = getCraPaths(cracoConfig);

    const craWebpackConfig = loadWebpackConfig(cracoConfig);
    const resultingWebpackConfig = mergeWebpackConfig(cracoConfig, craWebpackConfig, context);

    return resultingWebpackConfig;
}

module.exports = {
    createWebpackDevConfig,
    createWebpackProdConfig
};
