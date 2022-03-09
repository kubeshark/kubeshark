const { cosmiconfigSync } = require("cosmiconfig");
const { default: tsLoader } = require("cosmiconfig-typescript-loader");

const path = require("path");

const { getArgs } = require("./args");
const { log } = require("./logger");
const { projectRoot } = require("./paths");
const { deepMergeWithArray, isArray, isFunction, isString } = require("./utils");
const { applyCracoConfigPlugins } = require("./features/plugins");
const { POSTCSS_MODES } = require("./features/webpack/style/postcss");
const { ESLINT_MODES } = require("./features/webpack/eslint");

const DEFAULT_CONFIG = {
    reactScriptsVersion: "react-scripts",
    style: {
        postcss: {
            mode: POSTCSS_MODES.extends
        }
    },
    eslint: {
        mode: ESLINT_MODES.extends
    },
    jest: {
        babel: {
            addPresets: true,
            addPlugins: true
        }
    }
};

const moduleName = "craco";
const explorer = cosmiconfigSync(moduleName, {
    searchPlaces: [
        "package.json",
        `${moduleName}.config.ts`,
        `${moduleName}.config.js`,
        `${moduleName}.config.cjs`,
        `.${moduleName}rc.ts`,
        `.${moduleName}rc.js`,
        `.${moduleName}rc`
    ],
    loaders: {
        ".ts": tsLoader()
    }
});

function ensureConfigSanity(cracoConfig) {
    if (isArray(cracoConfig.plugins)) {
        cracoConfig.plugins.forEach((x, index) => {
            if (!x.plugin) {
                throw new Error(`craco: Malformed plugin at index: ${index} of 'plugins'.`);
            }
        });
    }
}

function processCracoConfig(cracoConfig, context) {
    let resultingCracoConfig = deepMergeWithArray({}, DEFAULT_CONFIG, cracoConfig);
    ensureConfigSanity(resultingCracoConfig);

    return applyCracoConfigPlugins(resultingCracoConfig, context);
}

function getConfigPath() {
    const args = getArgs();

    if (args.config.isProvided) {
        return path.resolve(projectRoot, args.config.value);
    } else {
        const packageJsonPath = path.join(projectRoot, "package.json");

        const package = require(packageJsonPath);

        if (package.cracoConfig && isString(package.cracoConfig)) {
            // take it as the path to the config file if it's path-like, otherwise assume it contains the config content below
            return path.resolve(projectRoot, package.cracoConfig);
        } else {
            const result = explorer.search(projectRoot);

            if (result === null) {
                throw new Error(
                    "craco: Config file not found. check if file exists at root (craco.config.ts, craco.config.js, .cracorc.js, .cracorc.json, .cracorc.yaml, .cracorc)"
                );
            }

            return result.filepath;
        }
    }
}

function getConfigAsObject(context) {
    const configFilePath = getConfigPath();
    log("Config file path resolved to: ", configFilePath);
    const result = explorer.load(configFilePath);

    const configAsObject = isFunction(result.config) ? result.config(context) : result.config;

    if (!configAsObject) {
        throw new Error("craco: Config function didn't return a config object.");
    }
    return configAsObject;
}

function loadCracoConfig(context) {
    const configAsObject = getConfigAsObject(context);

    if (configAsObject instanceof Promise) {
        throw new Error(
            "craco: Config function returned a promise. Use `loadCracoConfigAsync` instead of `loadCracoConfig`."
        );
    }

    return processCracoConfig(configAsObject, context);
}

// The "build", "start", and "test" scripts use this to wait for any promises to resolve before they run.
async function loadCracoConfigAsync(context) {
    const configAsObject = await getConfigAsObject(context);

    if (!configAsObject) {
        throw new Error("craco: Async config didn't return a config object.");
    }

    return processCracoConfig(configAsObject, context);
}

module.exports = {
    loadCracoConfig,
    loadCracoConfigAsync,
    processCracoConfig
};
