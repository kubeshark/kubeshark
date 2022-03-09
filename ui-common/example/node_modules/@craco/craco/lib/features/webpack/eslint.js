const { log, logError } = require("../../logger");
const { isFunction, deepMergeWithArray } = require("../../utils");
const { getPlugin, removePlugins, pluginByName } = require("../../webpack-plugins");

const ESLINT_MODES = {
    extends: "extends",
    file: "file"
};

function disableEslint(webpackConfig) {
    const { hasRemovedAny } = removePlugins(webpackConfig, pluginByName("ESLintWebpackPlugin"));

    if (hasRemovedAny) {
        log("Disabled ESLint.");
    } else {
        logError("Couldn't disabled ESLint.");
    }
}

function extendsEslintConfig(plugin, eslintConfig, context) {
    const { configure } = eslintConfig;

    if (configure) {
        if (isFunction(configure)) {
            if (plugin.options) {
                plugin.options.baseConfig = configure(plugin.options.baseConfig || {}, context);
            } else {
                plugin.options = {
                    baseConfig: configure({}, context)
                };
            }

            if (!plugin.options.baseConfig) {
                throw new Error("craco: 'eslint.configure' function didn't return a config object.");
            }
        } else {
            // TODO: ensure is otherwise a plain object, if not, log an error.
            if (plugin.options) {
                plugin.options.baseConfig = deepMergeWithArray({}, plugin.options.baseConfig || {}, configure);
            } else {
                plugin.options = {
                    baseConfig: configure
                };
            }
        }

        log("Merged ESLint config with 'eslint.configure'.");
    }
}

function useEslintConfigFile(plugin) {
    if (plugin.options) {
        plugin.options.useEslintrc = true;
        delete plugin.options.baseConfig;
    } else {
        plugin.options = {
            useEslintrc: true
        };
    }

    log("Overrided ESLint config to use a config file.");
}

function enableEslintIgnoreFile(plugin) {
    if (plugin.options) {
        plugin.options.ignore = true;
    } else {
        plugin.options = {
            ignore: true
        };
    }

    log("Overrided ESLint config to enable an ignore file.");
}

function applyPluginOptions(plugin, pluginOptions, context) {
    if (isFunction(pluginOptions)) {
        plugin.options = pluginOptions(plugin.options || {}, context);

        if (!plugin.options) {
            throw new Error("craco: 'eslint.pluginOptions' function didn't return a config object.");
        }
    } else {
        // TODO: ensure is otherwise a plain object, if not, log an error.
        plugin.options = deepMergeWithArray(plugin.options || {}, pluginOptions);
    }

    log("Applied ESLint plugin options.");
}

function overrideEsLint(cracoConfig, webpackConfig, context) {
    if (cracoConfig.eslint) {
        const { isFound, match } = getPlugin(webpackConfig, pluginByName("ESLintWebpackPlugin"));
        if (!isFound) {
            logError("Cannot find ESLint plugin (ESLintWebpackPlugin).");
            return webpackConfig;
        }

        const { enable, mode, pluginOptions } = cracoConfig.eslint;

        if (enable === false) {
            disableEslint(webpackConfig);

            return webpackConfig;
        }

        enableEslintIgnoreFile(match);

        if (mode === ESLINT_MODES.file) {
            useEslintConfigFile(match);
        } else {
            extendsEslintConfig(match, cracoConfig.eslint, context);
        }

        if (pluginOptions) {
            applyPluginOptions(match, pluginOptions, context);
        }
    }

    return webpackConfig;
}

module.exports = {
    overrideEsLint,
    ESLINT_MODES
};
