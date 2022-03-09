const { getLoaders, loaderByName } = require("../../loaders");
const { log, logError } = require("../../logger");
const { isFunction, isArray, deepMergeWithArray } = require("../../utils");

// TODO: CRA use a cacheIdentifier, should we update it with the new plugins?

function addPresets(loader, babelPresets) {
    if (isArray(babelPresets)) {
        if (loader.options) {
            if (loader.options.presets) {
                loader.options.presets = loader.options.presets.concat(babelPresets);
            } else {
                loader.options.presets = babelPresets;
            }
        } else {
            loader.options = {
                presets: babelPresets
            };
        }
    }

    log("Added Babel presets.");
}

function addPlugins(loader, babelPlugins) {
    if (isArray(babelPlugins)) {
        if (loader.options) {
            if (loader.options.plugins) {
                loader.options.plugins = loader.options.plugins.concat(babelPlugins);
            } else {
                loader.options.plugins = babelPlugins;
            }
        } else {
            loader.options = {
                plugins: babelPlugins
            };
        }
    }

    log("Added Babel plugins.");
}

function applyLoaderOptions(loader, loaderOptions, context) {
    if (isFunction(loaderOptions)) {
        loader.options = loaderOptions(loader.options || {}, context);

        if (!loader.options) {
            throw new Error("craco: 'babel.loaderOptions' function didn't return a loader config object.");
        }
    } else {
        // TODO: ensure is otherwise a plain object, if not, log an error.
        loader.options = deepMergeWithArray({}, loader.options || {}, loaderOptions);
    }

    log("Applied Babel loader options.");
}

function overrideLoader(match, cracoConfig, context) {
    const { presets, plugins, loaderOptions } = cracoConfig.babel;

    if (presets) {
        addPresets(match.loader, presets);
    }

    if (plugins) {
        addPlugins(match.loader, plugins);
    }

    if (loaderOptions) {
        applyLoaderOptions(match.loader, loaderOptions, context);
    }
}

function overrideBabel(cracoConfig, webpackConfig, context) {
    if (cracoConfig.babel) {
        const { hasFoundAny, matches } = getLoaders(webpackConfig, loaderByName("babel-loader"));

        if (!hasFoundAny) {
            logError("Cannot find any Babel loaders.");

            return webpackConfig;
        }

        matches.forEach(x => {
            overrideLoader(x, cracoConfig, context);
        });
    }

    return webpackConfig;
}

module.exports = {
    overrideBabel
};
