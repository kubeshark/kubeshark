const { isString } = require("../../utils");

function setEnvironmentVariable(envProperty, value) {
    if (!isString(value)) {
        process.env[envProperty] = value.toString();
    } else {
        process.env[envProperty] = value;
    }
}

function setEnvironmentVariables(cracoConfig) {
    if (cracoConfig.devServer) {
        const { open, https, host, port } = cracoConfig.devServer;

        if (open === false) {
            setEnvironmentVariable("BROWSER", "none");
        }

        if (https) {
            setEnvironmentVariable("HTTPS", "true");
        }

        if (host) {
            setEnvironmentVariable("HOST", host);
        }

        if (port) {
            setEnvironmentVariable("PORT", port);
        }
    }
}

module.exports = {
    setEnvironmentVariables
};
