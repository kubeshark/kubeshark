const babelJest = require("babel-jest");

const { isArray } = require("../../utils");
/**
 * To check if support jsx-runtime
 * Copy from https://github.com/facebook/create-react-app/blob/2b1161b34641bb4d2f269661cd636bbcd4888406/packages/react-scripts/config/jest/babelTransform.js#L12
 */
const hasJsxRuntime = (() => {
    if (process.env.DISABLE_NEW_JSX_TRANSFORM === "true") {
        return false;
    }

    try {
        require.resolve("react/jsx-runtime");
        return true;
    } catch (e) {
        return false;
    }
})();

function createJestBabelTransform(cracoConfig) {
    const craBabelTransformer = {
        presets: [
            [
                "babel-preset-react-app",
                {
                    runtime: hasJsxRuntime ? "automatic" : "classic"
                }
            ]
        ],
        babelrc: false,
        configFile: false
    };

    if (cracoConfig) {
        const { addPresets, addPlugins } = cracoConfig.jest.babel;

        if (cracoConfig.babel) {
            if (addPresets) {
                const { presets } = cracoConfig.babel;

                if (isArray(presets)) {
                    craBabelTransformer.presets = craBabelTransformer.presets.concat(presets);
                }
            }

            if (addPlugins) {
                const { plugins } = cracoConfig.babel;

                if (isArray(plugins)) {
                    craBabelTransformer.plugins = plugins;
                }
            }
        }
    }
    return babelJest.createTransformer(craBabelTransformer);
}

module.exports = {
    createJestBabelTransform
};
