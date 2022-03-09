const merge = require("webpack-merge");

const { isFunction } = require("../../utils");
const { log } = require("../../logger");
const { applyDevServerConfigPlugins } = require("../plugins");
const { loadDevServerConfigProvider } = require("../../cra");

function createProxy(cracoConfig, craDevServerConfigProvider, context) {
    const proxy = (proxy, allowedHost) => {
        let devServerConfig = craDevServerConfigProvider(proxy, allowedHost);

        if (isFunction(cracoConfig.devServer)) {
            devServerConfig = cracoConfig.devServer(devServerConfig, {
                ...context,
                proxy,
                allowedHost
            });

            if (!devServerConfig) {
                throw new Error("craco: 'devServer' function didn't return a config object.");
            }
        } else {
            // TODO: ensure is otherwise a plain object, if not, log an error.
            devServerConfig = merge(devServerConfig, cracoConfig.devServer || {});
        }

        devServerConfig = applyDevServerConfigPlugins(cracoConfig, devServerConfig, {
            ...context,
            proxy,
            allowedHost
        });

        log("Merged DevServer config.");

        return devServerConfig;
    };

    return proxy;
}

function createConfigProviderProxy(cracoConfig, context) {
    const craDevServerConfigProvider = loadDevServerConfigProvider(cracoConfig);
    const proxy = createProxy(cracoConfig, craDevServerConfigProvider, context);

    return proxy;
}

module.exports = {
    createConfigProviderProxy
};
