const path = require("path");

const { isFunction, isArray, deepMergeWithArray } = require("../../utils");
const { log } = require("../../logger");
const { applyJestConfigPlugins } = require("../plugins");
const { projectRoot } = require("../../paths");

const BABEL_TRANSFORM_ENTRY_KEY = "^.+\\.(js|jsx|mjs|cjs|ts|tsx)$";

function overrideBabelTransform(jestConfig, cracoConfig, transformKey) {
    // The cracoConfig needs to be available within the jest-babel-transform in order to honor its settings.
    // This approach is based on https://github.com/facebook/jest/issues/1468#issuecomment-384825178
    jestConfig.globals = jestConfig.globals || {};
    jestConfig.globals._cracoConfig = cracoConfig;

    jestConfig.transform[transformKey] = require.resolve("./jest-babel-transform");

    log("Overrided Jest Babel transformer.");
}

function configureBabel(jestConfig, cracoConfig) {
    const { addPresets, addPlugins } = cracoConfig.jest.babel;

    if (addPresets || addPlugins) {
        if (cracoConfig.babel) {
            const { presets, plugins } = cracoConfig.babel;

            if (isArray(presets) || isArray(plugins)) {
                if (jestConfig.transform[BABEL_TRANSFORM_ENTRY_KEY]) {
                    overrideBabelTransform(jestConfig, cracoConfig, BABEL_TRANSFORM_ENTRY_KEY);
                } else {
                    throw new Error(`craco: Cannot find Jest transform entry for Babel ${BABEL_TRANSFORM_ENTRY_KEY}.`);
                }
            }
        }
    }
}

function giveTotalControl(jestConfig, configureJest, context) {
    if (isFunction(configureJest)) {
        jestConfig = configureJest(jestConfig, context);

        if (!jestConfig) {
            throw new Error("craco: 'jest.configure' function didn't returned a Jest config object.");
        }
    } else {
        // TODO: ensure is otherwise a plain object, if not, log an error.
        jestConfig = deepMergeWithArray({}, jestConfig, configureJest);
    }

    log("Merged Jest config with 'jest.configure'.");

    return jestConfig;
}

function mergeJestConfig(cracoConfig, craJestConfigProvider, context) {
    const customResolve = relativePath =>
        require.resolve(path.join(cracoConfig.reactScriptsVersion, relativePath), { paths: [projectRoot] });

    let jestConfig = craJestConfigProvider(customResolve, projectRoot, false);

    if (cracoConfig.jest) {
        configureBabel(jestConfig, cracoConfig);

        const jestContext = {
            ...context,
            resolve: customResolve,
            rootDir: projectRoot
        };

        if (cracoConfig.jest.configure) {
            jestConfig = giveTotalControl(jestConfig, cracoConfig.jest.configure, jestContext);
        }

        jestConfig = applyJestConfigPlugins(cracoConfig, jestConfig, jestContext);

        log("Merged Jest config.");
    }

    return jestConfig;
}

module.exports = {
    mergeJestConfig
};
