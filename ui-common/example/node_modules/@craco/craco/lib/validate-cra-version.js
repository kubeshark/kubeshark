const { getReactScriptVersion } = require("../lib/cra");

function validateCraVersion(cracoConfig) {
    const { isSupported, version } = getReactScriptVersion(cracoConfig);
    if (!isSupported) {
        throw new Error(
            `Your current version of react-scripts(${version}) is not supported by this version of CRACO. Please try updating react-scripts to the latest version:\n\n` +
                `   $ yarn upgrade react-scripts\n\n` +
                "Or:\n\n" +
                `   $ npm update react-scripts\n\n` +
                `If that doesn't work or if you can't, refer to the following table to choose the right version of CRACO.\n` +
                "https://github.com/gsoft-inc/craco/blob/master/packages/craco/README.md#backward-compatibility"
        );
    }
}

module.exports = {
    validateCraVersion
};
