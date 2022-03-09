process.env.NODE_ENV = "production";

const { findArgsFromCli } = require("../lib/args");

// Make sure this is called before "paths" is imported.
findArgsFromCli();

const { log } = require("../lib/logger");
const { getCraPaths, build } = require("../lib/cra");
const { loadCracoConfigAsync } = require("../lib/config");
const { overrideWebpackProd } = require("../lib/features/webpack/override");
const { validateCraVersion } = require("../lib/validate-cra-version");

log("Override started with arguments: ", process.argv);
log("For environment: ", process.env.NODE_ENV);

const context = {
    env: process.env.NODE_ENV
};

loadCracoConfigAsync(context).then(cracoConfig => {
    validateCraVersion(cracoConfig);

    context.paths = getCraPaths(cracoConfig);

    overrideWebpackProd(cracoConfig, context);
    build(cracoConfig);
});
