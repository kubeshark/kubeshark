const { getLoaders, loaderByName } = require("../../../loaders");
const { log, logError } = require("../../../logger");
const { isArray, isFunction, deepMergeWithArray } = require("../../../utils");
const { projectRoot } = require("../../../paths");

const POSTCSS_MODES = {
    extends: "extends",
    file: "file"
};

const CRA_PLUGINS = presetEnv => {
    // prettier-ignore
    return [
        require("postcss-flexbugs-fixes"),
        require("postcss-preset-env")(presetEnv),
        require(require.resolve("postcss-normalize", { paths: [projectRoot] }))
    ];
};

const CRA_PRESET_ENV = {
    autoprefixer: {
        flexbox: "no-2009"
    },
    stage: 3
};

function usePostcssConfigFile(match) {
    if (match.loader.options) {
        const ident = match.loader.options.ident;
        const sourceMap = match.loader.options.sourceMap;

        match.loader.options = {
            ident: ident,
            sourceMap: sourceMap
        };

        log("Overwrited PostCSS config to use a config file.");
    }
}

function extendsPostcss(match, { plugins, env }) {
    if (isArray(plugins) || env) {
        let postcssPlugins;

        if (env) {
            const mergedPreset = deepMergeWithArray({}, CRA_PRESET_ENV, env);
            postcssPlugins = CRA_PLUGINS(mergedPreset);

            log("Merged PostCSS env preset.");
        } else {
            let craPlugins = [];

            if (match.loader.options) {
                craPlugins = match.loader.options.plugins();
            }

            postcssPlugins = craPlugins || [];
        }

        if (plugins) {
            postcssPlugins = typeof plugins === "function" ? plugins(postcssPlugins) : postcssPlugins.concat(plugins);

            log("Added PostCSS plugins.");
        }

        if (match.loader.options) {
            match.loader.options.plugins = () => postcssPlugins;
        } else {
            match.loader.options = {
                plugins: () => postcssPlugins
            };
        }
    }
}

function applyLoaderOptions(match, loaderOptions, context) {
    if (isFunction(loaderOptions)) {
        match.loader.options = loaderOptions(match.loader.options || {}, context);

        if (!match.loader.options) {
            throw new Error("craco: 'style.postcss.loaderOptions' function didn't return a loader config object.");
        }
    } else {
        // TODO: ensure is otherwise a plain object, if not, log an error.
        match.loader.options = deepMergeWithArray({}, match.loader.options || {}, loaderOptions);
    }

    log("Applied PostCSS loaders options.");
}

function overrideLoader(match, postcssConfig, context) {
    const { mode, loaderOptions } = postcssConfig;

    if (mode === POSTCSS_MODES.file) {
        usePostcssConfigFile(match);
    } else {
        extendsPostcss(match, postcssConfig);
    }

    if (loaderOptions) {
        applyLoaderOptions(match, loaderOptions, context);
    }

    log("Overrided PostCSS loader.");
}

function overridePostcss(cracoConfig, webpackConfig, context) {
    if (cracoConfig.postcss) {
        const { hasFoundAny, matches } = getLoaders(webpackConfig, loaderByName("postcss-loader"));

        if (!hasFoundAny) {
            logError("Cannot find any PostCSS loaders.");

            return webpackConfig;
        }

        matches.forEach(x => {
            overrideLoader(x, cracoConfig.postcss, context);
        });
    }

    return webpackConfig;
}

module.exports = {
    overridePostcss,
    POSTCSS_MODES
};
