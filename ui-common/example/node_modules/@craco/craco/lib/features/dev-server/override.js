const { overrideDevServerConfigProvider } = require("../../cra");
const { createConfigProviderProxy } = require("./create-config-provider-proxy");
const { setEnvironmentVariables } = require("./set-environment-variables");
const { overrideDevServerUtils } = require("./override-utils");

function overrideDevServer(cracoConfig, context) {
    overrideDevServerUtils(cracoConfig);
    setEnvironmentVariables(cracoConfig);

    const proxy = createConfigProviderProxy(cracoConfig, context);
    overrideDevServerConfigProvider(cracoConfig, proxy);
}

module.exports = {
    overrideDevServer
};
