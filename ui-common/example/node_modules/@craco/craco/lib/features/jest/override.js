const { overrideJestConfigProvider } = require("../../cra");
const { mergeJestConfig } = require("./merge-jest-config");
const { log } = require("../../logger");
const { loadJestConfigProvider } = require("../../cra");

function overrideJest(cracoConfig, context) {
    if (cracoConfig.jest) {
        const craJestConfigProvider = loadJestConfigProvider(cracoConfig);

        const proxy = () => {
            return mergeJestConfig(cracoConfig, craJestConfigProvider, context);
        };

        overrideJestConfigProvider(cracoConfig, proxy);

        log("Overrided Jest config.");
    }
}

module.exports = {
    overrideJest
};
