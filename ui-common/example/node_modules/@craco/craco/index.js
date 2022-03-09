const {
    getLoader,
    getLoaders,
    removeLoaders,
    addBeforeLoader,
    addBeforeLoaders,
    addAfterLoader,
    addAfterLoaders,
    loaderByName
} = require("./lib/loaders");
const { getPlugin, pluginByName, addPlugins, removePlugins } = require("./lib/webpack-plugins");
const { when, whenDev, whenProd, whenTest } = require("./lib/user-config-utils");
const { throwUnexpectedConfigError, gitHubIssueUrl } = require("./lib/plugin-utils");
const { ESLINT_MODES } = require("./lib/features/webpack/eslint");
const { POSTCSS_MODES } = require("./lib/features/webpack/style/postcss");
const { createJestConfig } = require("./lib/features/jest/api");
const { createWebpackDevConfig, createWebpackProdConfig } = require("./lib/features/webpack/api");
const { createDevServerConfigProviderProxy } = require("./lib/features/dev-server/api");

module.exports = {
    getLoader,
    getLoaders,
    removeLoaders,
    addBeforeLoader,
    addBeforeLoaders,
    addAfterLoader,
    addAfterLoaders,
    loaderByName,
    getPlugin,
    pluginByName,
    addPlugins,
    removePlugins,
    when,
    whenDev,
    whenProd,
    whenTest,
    throwUnexpectedConfigError,
    gitHubIssueUrl,
    ESLINT_MODES,
    POSTCSS_MODES,
    createJestConfig,
    createWebpackDevConfig,
    createWebpackProdConfig,
    createDevServerConfigProviderProxy
};
