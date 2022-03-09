const merge = require("webpack-merge");

const { isFunction, isArray } = require("../../utils");
const { log } = require("../../logger");
const { overrideBabel } = require("./babel");
const { overrideEsLint } = require("./eslint");
const { overrideStyle } = require("./style/style");
const { overrideTypeScript } = require("./typescript");
const { applyWebpackConfigPlugins } = require("../plugins");
const {
    addPlugins: addWebpackPlugins,
    removePlugins: removeWebpackPlugins,
    pluginByName
} = require("../../webpack-plugins");

function addAlias(webpackConfig, webpackAlias) {
    // TODO: ensure is a plain object, if not, log an error.
    webpackConfig.resolve.alias = Object.assign(webpackConfig.resolve.alias || {}, webpackAlias);

    log("Added webpack alias.");
}

function addPlugins(webpackConfig, webpackPlugins) {
    if (isArray(webpackPlugins)) {
        addWebpackPlugins(webpackConfig, webpackPlugins);

        log("Added webpack plugins.");
    } else {
        throw new Error(`craco: 'webpack.plugins.add' needs to be a an array of plugins`);
    }
}

function removePluginsFromWebpackConfig(webpackConfig, remove) {
    if (!remove) {
        return;
    }

    if (isArray(remove)) {
        for (const pluginName of remove) {
            const { hasRemovedAny } = removeWebpackPlugins(webpackConfig, pluginByName(pluginName));

            if (hasRemovedAny) {
                log(`Removed webpack plugin ${pluginName}.`);
            } else {
                log(`Did not remove webpack plugin ${pluginName}.`);
            }
        }

        log("Removed webpack plugins.");
    } else {
        throw new Error(`craco: 'webpack.plugins.remove' needs to be a an array of plugin names`);
    }
}

function giveTotalControl(webpackConfig, configureWebpack, context) {
    if (isFunction(configureWebpack)) {
        webpackConfig = configureWebpack(webpackConfig, context);

        if (!webpackConfig) {
            throw new Error("craco: 'webpack.configure' function didn't returned a webpack config object.");
        }
    } else {
        // TODO: ensure is otherwise a plain object, if not, log an error.
        webpackConfig = merge(webpackConfig, configureWebpack);
    }

    log("Merged webpack config with 'webpack.configure'.");

    return webpackConfig;
}

function mergeWebpackConfig(cracoConfig, webpackConfig, context) {
    let resultingWebpackConfig = webpackConfig;

    resultingWebpackConfig = overrideBabel(cracoConfig, resultingWebpackConfig, context);
    resultingWebpackConfig = overrideEsLint(cracoConfig, resultingWebpackConfig, context);
    resultingWebpackConfig = overrideStyle(cracoConfig, resultingWebpackConfig, context);
    resultingWebpackConfig = overrideTypeScript(cracoConfig, resultingWebpackConfig, context);

    if (cracoConfig.webpack) {
        const { alias, plugins, configure } = cracoConfig.webpack;

        if (alias) {
            addAlias(resultingWebpackConfig, alias);
        }

        if (plugins) {
            // we still support the old format of plugin: [] where the array is a list of the plugins to add
            if (isArray(plugins)) {
                addPlugins(resultingWebpackConfig, plugins);
            } else {
                const { add, remove } = plugins;

                if (remove) {
                    removePluginsFromWebpackConfig(resultingWebpackConfig, remove);
                }

                // Add after removing to preserve any plugins explicitely added via the Craco config
                if (add) {
                    addPlugins(resultingWebpackConfig, add);
                }
            }
        }

        if (configure) {
            resultingWebpackConfig = giveTotalControl(resultingWebpackConfig, configure, context);
        }
    }

    resultingWebpackConfig = applyWebpackConfigPlugins(cracoConfig, resultingWebpackConfig, context);

    return resultingWebpackConfig;
}

module.exports = {
    mergeWebpackConfig
};
