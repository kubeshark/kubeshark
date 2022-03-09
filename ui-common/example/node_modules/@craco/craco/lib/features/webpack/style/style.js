const { overrideCss } = require("./css");
const { overrideSass } = require("./sass");
const { overridePostcss } = require("./postcss");

function overrideStyle(cracoConfig, webpackConfig, context) {
    if (cracoConfig.style) {
        webpackConfig = overrideCss(cracoConfig.style, webpackConfig, context);
        webpackConfig = overrideSass(cracoConfig.style, webpackConfig, context);
        webpackConfig = overridePostcss(cracoConfig.style, webpackConfig, context);
    }

    return webpackConfig;
}

module.exports = {
    overrideStyle
};
