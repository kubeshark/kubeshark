function pluginByName(targetPluginName) {
    return plugin => {
        return plugin.constructor.name === targetPluginName;
    };
}

function getPlugin(webpackConfig, matcher) {
    const matchingPlugin = webpackConfig.plugins.find(matcher);

    return {
        isFound: matchingPlugin !== undefined,
        match: matchingPlugin
    };
}

function addPlugins(webpackConfig, webpackPlugins) {
    const prependPlugins = [];
    const appendPlugins = [];

    for (const webpackPlugin of webpackPlugins) {
        if (Array.isArray(webpackPlugin)) {
            const [plugin, order] = webpackPlugin;
            if (order === "append") {
                appendPlugins.push(plugin);
            } else {
                // Existing behaviour is to prepend
                prependPlugins.push(plugin);
            }
            continue;
        }
        prependPlugins.push(webpackPlugin);
    }

    webpackConfig.plugins = [...prependPlugins, ...webpackConfig.plugins, ...appendPlugins];
}

function removePlugins(webpackConfig, matcher) {
    const prevCount = webpackConfig.plugins.length;
    webpackConfig.plugins = webpackConfig.plugins.filter(x => !matcher(x));
    const removedPluginsCount = prevCount - webpackConfig.plugins.length;

    return {
        hasRemovedAny: removedPluginsCount > 0,
        removedCount: removedPluginsCount
    };
}

module.exports = {
    getPlugin,
    pluginByName,
    addPlugins,
    removePlugins
};
