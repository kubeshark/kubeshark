process.env.NODE_ENV = process.env.NODE_ENV || "test";

const { findArgsFromCli } = require("../lib/args");

// Make sure this is called before "paths" is imported.
findArgsFromCli();

const { log } = require("../lib/logger");
const { getCraPaths, test } = require("../lib/cra");
const { overrideJest } = require("../lib/features/jest/override");
const { loadCracoConfigAsync } = require("../lib/config");
const { validateCraVersion } = require("../lib/validate-cra-version");

log("Override started with arguments: ", process.argv);
log("For environment: ", process.env.NODE_ENV);

const context = {
    env: process.env.NODE_ENV
};

loadCracoConfigAsync(context).then(cracoConfig => {
    validateCraVersion(cracoConfig);

    context.paths = getCraPaths(cracoConfig);

    overrideJest(cracoConfig, context);
    test(cracoConfig);
});
