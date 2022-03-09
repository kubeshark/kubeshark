const { isFunction } = require("../../utils");
const { getCraPaths } = require("../../cra");
const { mergeJestConfig } = require("./merge-jest-config");
const { loadJestConfigProvider } = require("../../cra");
const { setArgs } = require("../../args");
const { processCracoConfig } = require("../../config");

function createJestConfig(callerCracoConfig, callerContext = {}, options = {}) {
    if (!callerCracoConfig) {
        throw new Error("craco: 'cracoConfig' is required.");
    }
    if (isFunction(callerCracoConfig)) {
        throw new Error("craco: 'cracoConfig' should be an object.");
    }

    if (!process.env.NODE_ENV) {
        process.env.NODE_ENV = "development";
    }

    setArgs(options);

    const context = {
        env: process.env.NODE_ENV,
        ...callerContext
    };

    const cracoConfig = processCracoConfig(callerCracoConfig, context);
    context.paths = getCraPaths(cracoConfig);

    const craJestConfigProvider = loadJestConfigProvider(cracoConfig);

    return mergeJestConfig(cracoConfig, craJestConfigProvider, context);
}

module.exports = {
    createJestConfig
};
