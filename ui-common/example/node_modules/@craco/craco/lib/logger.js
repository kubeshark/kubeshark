const { getArgs } = require("./args");

function log(...rest) {
    if (getArgs().verbose) {
        console.log("craco: ", ...rest);
    }
}

function logError(...rest) {
    console.error("craco:  ***", ...rest, "***");
}

module.exports = {
    log,
    logError
};
