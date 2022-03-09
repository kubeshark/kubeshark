const { loadDevServerUtils, overrideDevServerUtils } = require("../../cra");
const { log } = require("../../logger");

function overrideWebpackCompilerToDisableTypeScriptTypeChecking(craDevServersUtils) {
    if (craDevServersUtils.createCompiler) {
        const craCreateCompiler = craDevServersUtils.createCompiler;

        craDevServersUtils.createCompiler = args => {
            const newArgs = {
                ...args,
                useTypeScript: false
            };

            return craCreateCompiler(newArgs);
        };

        log("Overrided Webpack compiler to disable TypeScript type checking.");
    }

    return craDevServersUtils;
}

function overrideUtils(cracoConfig) {
    if (cracoConfig.typescript) {
        const { enableTypeChecking } = cracoConfig.typescript;

        if (enableTypeChecking === false) {
            const craDevServersUtils = loadDevServerUtils();
            const resultingDevServersUtils = overrideWebpackCompilerToDisableTypeScriptTypeChecking(craDevServersUtils);

            overrideDevServerUtils(resultingDevServersUtils);
        }
    }
}

module.exports = {
    overrideDevServerUtils: overrideUtils
};
