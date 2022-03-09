const { getLoaders, loaderByName } = require("../../../loaders");
const { log, logError } = require("../../../logger");
const { isString, isFunction, deepMergeWithArray } = require("../../../utils");

function setLoaderProperty(match, key, valueProviders) {
    if (isString(match.loader)) {
        match.parent[match.index] = {
            loader: match.loader,
            [key]: valueProviders.whenString()
        };
    } else {
        match.loader[key] = valueProviders.whenObject();
    }
}

function applyLoaderOptions(match, loaderOptions, context) {
    if (isFunction(loaderOptions)) {
        setLoaderProperty(match, "options", {
            whenString: () => loaderOptions({}, context),
            whenObject: () => loaderOptions(match.loader.options || {}, context)
        });

        if (!match.loader.options) {
            throw new Error("craco: 'style.sass.loaderOptions' function didn't return a loader config object.");
        }
    } else {
        // TODO: ensure is otherwise a plain object, if not, log an error.
        setLoaderProperty(match, "options", {
            whenString: () => loaderOptions,
            whenObject: () => deepMergeWithArray({}, match.loader.options || {}, loaderOptions)
        });
    }

    log("Applied Sass loaders options.");
}

function overrideLoader(match, sassOptions, context) {
    const { loaderOptions } = sassOptions;

    if (loaderOptions) {
        applyLoaderOptions(match, loaderOptions, context);

        log("Overrided Sass loader.");
    }
}

function overrideSass(styleConfig, webpackConfig, context) {
    if (styleConfig.sass) {
        const { hasFoundAny, matches } = getLoaders(webpackConfig, loaderByName("sass-loader"));

        if (!hasFoundAny) {
            logError("Cannot find any Sass loaders.");

            return webpackConfig;
        }

        matches.forEach(x => {
            overrideLoader(x, styleConfig.sass, context);
        });
    }

    return webpackConfig;
}

module.exports = {
    overrideSass
};
