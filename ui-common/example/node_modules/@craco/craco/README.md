# CRACO [![Build Status](https://travis-ci.org/sharegate/craco.svg?branch=master)](https://travis-ci.org/sharegate/craco) [![PRs Welcome](https://img.shields.io/badge/PRs-welcome-green.svg)](https://github.com/sharegate/craco/pulls)

**C**reate **R**eact **A**pp **C**onfiguration **O**verride is an easy and comprehensible configuration layer for create-react-app.

Get all the benefits of create-react-app **and** customization without using 'eject' by adding a single configuration (e.g. `craco.config.js`) file at the root of your application and customize your eslint, babel, postcss configurations and many more.

All you have to do is create your app using [create-react-app](https://github.com/facebook/create-react-app/) and customize the configuration file.

## Support

- Create React App (CRA) 4.*
- Yarn
- Yarn Workspace
- NPM
- Lerna (with or without hoisting)
- Custom `react-scripts` version

## Documentation

- [Installation](#installation) - How to install and setup CRACO.
- [Configuration](#configuration) - How to customize your CRA installation with CRACO.
  - [Configuration File](#configuration-file)
  - [Configuration Helpers](#configuration-helpers)
  - [Exporting your Configuration](#exporting-your-configuration)
  - [Setting a Custom Location for the configuration file](#setting-a-custom-location-for-cracoconfigjs)
- [CRA toolchain for Beginners](#cra-toolchain-for-beginners)
  - [Notes on CRA Configurations and Problem Solving](#notes-on-cra-configurations-and-problem-solving)
  - [Ejecting CRA to Learn](#ejecting-cra-to-learn)
  - [Direct Versus Functional Config Definitions](#direct-object-literal-versus-functional-config-definitions)
- [API](#api) - CRACO APIs for Jest and Webpack.
  - [Jest API](#jest-api)
  - [Webpack API](#webpack-api)
- [Recipes](https://github.com/sharegate/craco/tree/master/recipes) – Short recipes for common use cases.
- [Available Plugins](https://github.com/sharegate/craco#community-maintained-plugins) - Plugins maintained by the community.
- [Develop a Plugin](#develop-a-plugin) - How to develop a plugin for CRACO.
- [Backward Compatibility](#backward-compatibility)
- [Debugging](#debugging)
- [License](#license)

## Preface

### Acknowledgements

We are grateful to [@timarney](https://github.com/timarney) the creator of [react-app-rewired](https://github.com/timarney/react-app-rewired) for his original idea.

The configuration style of this plugin has been greatly influenced by [Vue CLI](https://cli.vuejs.org/guide/).

### Fair Warning

By doing this you're breaking the ["guarantees"](https://github.com/facebookincubator/create-react-app/issues/99#issuecomment-234657710) that CRA provides. That is to say you now "own" the configs. **No support** will be provided. Proceed with caution.

## Installation

Install the plugin from **npm**:

```bash
$ yarn add @craco/craco

# OR

$ npm install @craco/craco --save
```

Create a `craco.config.js` file in the root directory and [configure CRACO](#configuration):

```
my-app
├── node_modules
├── craco.config.js
└── package.json
```

Update the existing calls to `react-scripts` in the `scripts` section of your `package.json` file to use the `craco` CLI:

```diff
/* package.json */

"scripts": {
-   "start": "react-scripts start",
+   "start": "craco start",
-   "build": "react-scripts build",
+   "build": "craco build"
-   "test": "react-scripts test",
+   "test": "craco test"
}
```

Start your app for development:

```bash
$ npm start
```

Or build your app:

```bash
$ npm run build
```

## Configuration

CRACO is configured with a `craco.config.ts`, `craco.config.js`, `.cracorc.ts`, `.cracorc.js` or `.cracorc` file, or [a file specified in `package.json`](#setting-a-custom-location-for-cracoconfigjs). This file is divided into sections representing the major parts of what makes up the default create react app. 

If there are multiple configuration files in the same directory, CRACO will only use one. The priority order is:

1. `package.json`
1. `craco.config.ts`
1. `craco.config.js`
1. `.cracorc.ts`
1. `.cracorc.js`
1. `.cracorc`

### Configuration File 

Below is a sample CRACO configuration file. Your final config file will be much shorter than this sample. See example CRACO configurations in [Recipes](https://github.com/sharegate/craco/tree/master/recipes).

Some sections have a `mode` property. When this is available there are 2 possible values:

- `extends`: the provided configuration will extends the CRA settings (**default mode**)
- `file`: the CRA settings will be reset and you will provide an official configuration file for the plugin ([postcss](https://github.com/michael-ciniawsky/postcss-load-config#postcssrc), [eslint](https://eslint.org/docs/user-guide/configuring#configuration-file-formats)) that will supersede any settings.

```javascript
const { when, whenDev, whenProd, whenTest, ESLINT_MODES, POSTCSS_MODES } = require("@craco/craco");

module.exports = {
    reactScriptsVersion: "react-scripts" /* (default value) */,
    style: {
        modules: {
            localIdentName: ""
        },
        css: {
            loaderOptions: { /* Any css-loader configuration options: https://github.com/webpack-contrib/css-loader. */ },
            loaderOptions: (cssLoaderOptions, { env, paths }) => { return cssLoaderOptions; }
        },
        sass: {
            loaderOptions: { /* Any sass-loader configuration options: https://github.com/webpack-contrib/sass-loader. */ },
            loaderOptions: (sassLoaderOptions, { env, paths }) => { return sassLoaderOptions; }
        },
        postcss: {
            mode: "extends" /* (default value) */ || "file",
            plugins: [require('plugin-to-append')], // Additional plugins given in an array are appended to existing config.
            plugins: (plugins) => [require('plugin-to-prepend')].concat(plugins), // Or you may use the function variant.
            env: {
                autoprefixer: { /* Any autoprefixer options: https://github.com/postcss/autoprefixer#options */ },
                stage: 3, /* Any valid stages: https://cssdb.org/#staging-process. */
                features: { /* Any CSS features: https://preset-env.cssdb.org/features. */ }
            },
            loaderOptions: { /* Any postcss-loader configuration options: https://github.com/postcss/postcss-loader. */ },
            loaderOptions: (postcssLoaderOptions, { env, paths }) => { return postcssLoaderOptions; }
        }
    },
    eslint: {
        enable: true /* (default value) */,
        mode: "extends" /* (default value) */ || "file",
        configure: { /* Any eslint configuration options: https://eslint.org/docs/user-guide/configuring */ },
        configure: (eslintConfig, { env, paths }) => { return eslintConfig; },
        pluginOptions: { /* Any eslint plugin configuration options: https://github.com/webpack-contrib/eslint-webpack-plugin#options. */ },
        pluginOptions: (eslintOptions, { env, paths }) => { return eslintOptions; }
    },
    babel: {
        presets: [],
        plugins: [],
        loaderOptions: { /* Any babel-loader configuration options: https://github.com/babel/babel-loader. */ },
        loaderOptions: (babelLoaderOptions, { env, paths }) => { return babelLoaderOptions; }
    },
    typescript: {
        enableTypeChecking: true /* (default value)  */
    },
    webpack: {
        alias: {},
        plugins: {
            add: [], /* An array of plugins */
            add: [
                plugin1,
                [plugin2, "append"],
                [plugin3, "prepend"], /* Specify if plugin should be appended or prepended */
            ], /* An array of plugins */
            remove: [],  /* An array of plugin constructor's names (i.e. "StyleLintPlugin", "ESLintWebpackPlugin" ) */
        },
        configure: { /* Any webpack configuration options: https://webpack.js.org/configuration */ },
        configure: (webpackConfig, { env, paths }) => { return webpackConfig; }
    },
    jest: {
        babel: {
            addPresets: true, /* (default value) */
            addPlugins: true  /* (default value) */
        },
        configure: { /* Any Jest configuration options: https://jestjs.io/docs/en/configuration. */ },
        configure: (jestConfig, { env, paths, resolve, rootDir }) => { return jestConfig; }
    },
    devServer: { /* Any devServer configuration options: https://webpack.js.org/configuration/dev-server/#devserver. */ },
    devServer: (devServerConfig, { env, paths, proxy, allowedHost }) => { return devServerConfig; },
    plugins: [
        {
            plugin: {
                overrideCracoConfig: ({ cracoConfig, pluginOptions, context: { env, paths } }) => { return cracoConfig; },
                overrideWebpackConfig: ({ webpackConfig, cracoConfig, pluginOptions, context: { env, paths } }) => { return webpackConfig; },
                overrideDevServerConfig: ({ devServerConfig, cracoConfig, pluginOptions, context: { env, paths, proxy, allowedHost } }) => { return devServerConfig; },
                overrideJestConfig: ({ jestConfig, cracoConfig, pluginOptions, context: { env, paths, resolve, rootDir } }) => { return jestConfig },
            },
            options: {}
        }
    ]
};
```

### Configuration Helpers

Usage for all "when" functions is the same, `whenDev, whenProd, whenTest` are shortcuts for `when`.

`when(condition, fct, [unmetValue])`

Usage:

```javascript
const { when, whenDev } = require("@craco/craco");

module.exports = {
    eslint: {
        mode: ESLINT_MODES.file,
        configure: {
            formatter: when(process.env.NODE_ENV === "CI", require("eslint-formatter-vso"))
        }
    },
    webpack: {
        plugins: [
            new ConfigWebpackPlugin(),
            ...whenDev(() => [new CircularDependencyPlugin()], [])
        ]
    }
};
```

### Exporting your Configuration

You can export your configuration as an **object literal**:

```javascript
/* craco.config.js */

module.exports = {
    ...
}
```

a **function**:

```javascript
/* craco.config.js */

module.exports = function({ env }) {
    return {
        ...
    };
}
```

a **promise** or an **async function**:

```javascript
/* craco.config.js */

module.exports = async function({ env }) {
    await ...

    return {
        ...
    };
}
```

### Setting a Custom Location for craco.config.js

Both options support a **relative** or an **absolute** path.

**1- package.json** _(Recommended)_

You can change the location of the `craco.config.js` file by specifying a value for `cracoConfig` in your `package.json` 
file.

```javascript
/* package.json */

{
    "cracoConfig": "config/craco-config-with-custom-name.js"
}
```

**2- CLI** _(For backward compatibility)_

You can also change the location of the `craco.config.js` file by specifying the `--config` CLI option. _This option 
doesn't support Babel with Jest_

```javascript
/* package.json */

{
    "scripts": {
        "start": "craco start --config config/craco-config-with-custom-name.js"
    }
}
```

## CRA Toolchain for Beginners

### Introduction to CRACO

Create React App ([CRA](https://github.com/facebook/create-react-app)) is intended to allow people to get started with 
writing React apps quickly. It does this by packaging several key components with a solid default configuration. 

After some initial experimentation, many people find the default CRA is not quite the right fit. Yet, selecting and configuring a toolchain featuring all of the components CRA already offers is overwhelming. 

CRACO allows you to enjoy the recognizable project structure of CRA while changing detailed configuration settings of 
each component. 

### Notes on CRA Configurations and Problem Solving  

Keep in mind that there are _some_ configuration settings available to CRA without CRACO. 

Getting exactly what you want may involve a combination of making changes your CRACO configuration file and by using 
some of the more limited _but still important_ settings available in Create React App. 

Before jumping into customizing your _CRACO_ configuration, step back and think about each part of the problem you're 
trying to solve. Be sure to review these resources on the CRA configuration, as it may save you time:

 - [Important Environment Variables that Configure CRA](https://create-react-app.dev/docs/advanced-configuration)
 - [Learn about using `postbuild` commands in `package.json`](https://stackoverflow.com/a/51818028/4028977)
 - [Proxying API or other Requests](https://create-react-app.dev/docs/proxying-api-requests-in-development/), or "how 
 to integrate CRA's dev server with a second backend": 
 [problem statement](https://github.com/facebook/create-react-app/issues/147)
 - [Search CRACO issues, for gotchas, hints and examples](https://github.com/gsoft-inc/craco/issues?q=is%3Aissue+sort%3Aupdated-desc)

### Ejecting CRA to Learn

Avoiding ejecting is a major goal for many CRACO users. However, if you're still learning toolchains and modern 
frontend workflows, it may be helpful to create a sample ejected CRA project to see how the default CRA app configures 
each of the components. 

While CRACO's sample configuration file inherits directly from CRA's default settings, seeing the default CRA config in
the ejected CRA file structure may give you useful perspective. 

You may even want to try testing a change in the ejected app to better understand how it would be done with your CRACO 
config-based project. 

### Direct (object literal) Versus Functional Config Definitions

The [sample CRACO config file]((#sample-craco-configuration-file)) is meant to show possibilities for configuring your CRA-based project. Each section 
contains a primary configuration area, `loaderOptions` or `configure`. These config areas are where you will make most 
of your detailed changes.

You, (or perhaps your IDE) may have noticed that the sections have duplicate keys, i.e. loaderOptions is listed twice
in the sample config file.

The reason for this is to allow you to choose between object literal or functionally defined configuration choices. 
There are a few reasons for this:

1. Sometimes it may be faster to test a minor change using keys. 
1. Other times a functional definition is necessary to get the right configuration.
1. While not common, a setting may **only** work if you use one or the other! See, 
[devServer port example](https://github.com/gsoft-inc/craco/issues/172#issuecomment-651505730)

#### A simple example of equivalent direct and functionally defined configuration settings:

##### Direct configuration (object literal)

```javascript
devServer: {
    writeToDisk: true
}
```

##### Functionally defined configuration

```javascript
devServer: (devServerConfig, { env, paths, proxy, allowedHost }) => {
    devServerConfig.writeToDisk = true; 
    return devServerConfig;
}
```

## API

To integrate with other tools, it's usefull to have access to the configuration generated by CRACO.

That's what CRACO APIs are for. The current API support Jest and Webpack.

### Jest API

Accept a `cracoConfig`, a `context` object and `options`. The generated Jest config object is returned.

> **Warning:** `createJestConfig` does NOT accept `cracoConfig` as a function. If your `craco.config.js` exposes a config 
function, you have to call it yourself before passing it to `createJestConfig`.

`createJestConfig(cracoConfig, context = {}, options = { verbose: false, config: null })`

Usage:

```javascript
/* jest.config.js */

const { createJestConfig } = require("@craco/craco");

const cracoConfig = require("./craco.config.js");
const jestConfig = createJestConfig(cracoConfig);

module.exports = jestConfig;
```

#### Examples

- [vscode-jest](https://github.com/sharegate/craco/tree/master/recipes/use-a-jest-config-file)

### Webpack API

You can create Webpack DevServer and Production configurations using `createWebpackDevConfig` and `createWebpackProdConfig`. 

Accept a `cracoConfig`, a `context` object and `options`. The generated Webpack config object is returned.

> **Warning:** Similar to `createJestConfig`, these functions do NOT accept `cracoConfig` as a function. If your 
`craco.config.js` exposes a config function, you have to call it yourself before passing it further.

`createWebpackDevConfig(cracoConfig, context = {}, options = { verbose: false, config: null })`

`createWebpackProdConfig(cracoConfig, context = {}, options = { verbose: false, config: null })`

Usage:

```javascript
/* webpack.config.js */

const { createWebpackDevConfig } = require("@craco/craco");

const cracoConfig = require("./craco.config.js");
const webpackConfig = createWebpackDevConfig(cracoConfig);

module.exports = webpackConfig;
```

## Develop a Plugin

### Hooks

There are four hooks available to a plugin:

- `overrideCracoConfig`: Let a plugin customize the config object before it's process by `craco`.
- `overrideWebpackConfig`: Let a plugin customize the `webpack` config that will be used by CRA.
- `overrideDevServerConfig`: Let a plugin customize the dev server config that will be used by CRA.
- `overrideJestConfig`: Let a plugin customize the `Jest` config that will be used by CRA.

**Important:**

Every function must return the updated config object.

#### overrideCracoConfig

The function `overrideCracoConfig` let a plugin override the config object **before** it's process by `craco`.

If a plugin define the function, it will be called with the config object read from the `craco.config.js` file provided 
by the consumer.

*The function must return a valid config object, otherwise `craco` will throw an error.*

The function will be called with a single object argument having the following structure:

```javascript
{
    cracoConfig: "The config object read from the craco.config.js file provided by the consumer",
    pluginOptions: "The plugin options provided by the consumer",
    context: {
        env: "The current NODE_ENV (development, production, etc..)",
        paths: "An object that contains all the paths used by CRA"
    }
}
```

##### Example

Plugin:

```javascript
/* craco-plugin-log-craco-config.js */

module.exports = {
    overrideCracoConfig: ({ cracoConfig, pluginOptions, context: { env, paths } }) => {
        if (pluginOptions.preText) {
            console.log(pluginOptions.preText);
        }

        console.log(JSON.stringify(cracoConfig, null, 4));

        // Always return the config object.
        return cracoConfig;
    }
};
```

Registration (in a `craco.config.js` file):

```javascript
const logCracoConfigPlugin = require("./craco-plugin-log-craco-config");

module.exports = {
    ...
    plugins: [
        { plugin: logCracoConfigPlugin, options: { preText: "Will log the craco config:" } }
    ]
};
```

#### overrideWebpackConfig

The function `overrideWebpackConfig` let a plugin override the `webpack` config object **after** it's been customized 
by `craco`.

*The function must return a valid config object, otherwise `craco` will throw an error.*

The function will be called with a single object argument having the following structure:

```javascript
{
    webpackConfig: "The webpack config object already customized by craco",
    cracoConfig: "The configuration object read from the craco.config.js file provided by the consumer",
    pluginOptions: "The plugin options provided by the consumer",
    context: {
        env: "The current NODE_ENV (development, production, etc..)",
        paths: "An object that contains all the paths used by CRA"
    }
}
```

##### Example

Plugin:

```javascript
/* craco-plugin-log-webpack-config.js */

module.exports = {
    overrideWebpackConfig: ({ webpackConfig, cracoConfig, pluginOptions, context: { env, paths } }) => {
        if (pluginOptions.preText) {
            console.log(pluginOptions.preText);
        }

        console.log(JSON.stringify(webpackConfig, null, 4));

        // Always return the config object.
        return webpackConfig;
    }
};
```

Registration (in a `craco.config.js` file):

```javascript
const logWebpackConfigPlugin = require("./craco-plugin-log-webpack-config");

module.exports = {
    ...
    plugins: [
        { plugin: logWebpackConfigPlugin, options: { preText: "Will log the webpack config:" } }
    ]
};
```

#### overrideDevServerConfig

The function `overrideDevServerConfig` let a plugin override the dev server config object **after** it's been customized
by `craco`.

*The function must return a valid config object, otherwise `craco` will throw an error.*

The function will be called with a single object argument having the following structure:

```javascript
{
    devServerConfig: "The dev server config object already customized by craco",
    cracoConfig: "The configuration object read from the craco.config.js file provided by the consumer",
    pluginOptions: "The plugin options provided by the consumer",
    context: {
        env: "The current NODE_ENV (development, production, etc..)",
        paths: "An object that contains all the paths used by CRA",
        allowedHost: "Provided by CRA"
    }
}
```

##### Example

Plugin:

```javascript
/* craco-plugin-log-dev-server-config.js */

module.exports = {
    overrideDevServerConfig: ({ devServerConfig, cracoConfig, pluginOptions, context: { env, paths, allowedHost } }) => {
        if (pluginOptions.preText) {
            console.log(pluginOptions.preText);
        }

        console.log(JSON.stringify(devServerConfig, null, 4));

        // Always return the config object.
        return devServerConfig;
    }
};
```

Registration (in a `craco.config.js` file):

```javascript
const logDevServerConfigPlugin = require("./craco-plugin-log-dev-server-config");

module.exports = {
    ...
    plugins: [
        { plugin: logDevServerConfigPlugin, options: { preText: "Will log the dev server config:" } }
    ]
};
```

#### overrideJestConfig

The function `overrideJestConfig` let a plugin override the `Jest` config object **after** it's been customized by 
`craco`.

*The function must return a valid config object, otherwise `craco` will throw an error.*

The function will be called with a single object argument having the following structure:

```javascript
{
    jestConfig: "The Jest config object already customized by craco",
    cracoConfig: "The configuration object read from the craco.config.js file provided by the consumer",
    pluginOptions: "The plugin options provided by the consumer",
    context: {
        env: "The current NODE_ENV (development, production, etc..)",
        paths: "An object that contains all the paths used by CRA",
        resolve: "Provided by CRA",
        rootDir: "Provided by CRA"
    }
}
```

##### Example

Plugin:

```javascript
/* craco-plugin-log-jest-config.js */

module.exports = {
    overrideJestConfig: ({ jestConfig, cracoConfig, pluginOptions, context: { env, paths, resolve, rootDir } }) => {
        if (pluginOptions.preText) {
            console.log(pluginOptions.preText);
        }

        console.log(JSON.stringify(jestConfig, null, 4));

        // Always return the config object.
        return jestConfig;
    }
};
```

Registration (in a `craco.config.js` file):

```javascript
const logJestConfigPlugin = require("./craco-plugin-log-jest-config");

module.exports = {
    ...
    plugins: [
        { plugin: logJestConfigPlugin, options: { preText: "Will log the Jest config:" } }
    ]
};
```

### Utility Functions

A few utility functions are provided by CRACO to help you develop a plugin:
 - `getLoader`
 - `getLoaders`
 - `removeLoaders`
 - `addBeforeLoader`
 - `addBeforeLoaders`
 - `addAfterLoader`
 - `addAfterLoaders`
 - `getPlugin`
 - `removePlugins`
 - `addPlugins`
 - `throwUnexpectedConfigError`
 
```javascript
const { getLoader, getLoaders, removeLoaders, loaderByName, getPlugin, removePlugins, addPlugins, pluginByName, throwUnexpectedConfigError } = require("@craco/craco");
```

#### getLoader

Retrieve the **first** loader that match the specified criteria from the webpack config.

Returns:

```javascript
{
    isFound: true | false,
    match: {
        loader,
        parent,
        index
    }
}
```

Usage:

```javascript
const { getLoader, loaderByName } = require("@craco/craco");

const { isFound, match } = getLoader(webpackConfig, loaderByName("eslint-loader"));

if (isFound) {
    // do stuff...
}
```

#### getLoaders

Retrieve **all** the loaders that match the specified criteria from the webpack config.

Returns:

```javascript
{
    hasFoundAny: true | false,
    matches: [
        {
            loader,
            parent,
            index
        }
    ]
}
```

Usage:

```javascript
const { getLoaders, loaderByName } = require("@craco/craco");

const { hasFoundAny, matches } = getLoaders(webpackConfig, loaderByName("babel-loader"));

if (hasFoundAny) {
    matches.forEach(x => {
        // do stuff...
    });
}
```

#### removeLoaders

Remove **all** the loaders that match the specified criteria from the webpack config.

Returns:

```javascript
{
    hasRemovedAny: true | false,
    removedCount: int
}
```

Usage:

```javascript
const { removeLoaders, loaderByName } = require("@craco/craco");

removeLoaders(webpackConfig, loaderByName("eslint-loader"));
```

#### addBeforeLoader

Add a new *loader* **before** the loader that match specified criteria to the webpack config.

Returns:

```javascript
{
    isAdded: true | false
}
```

Usage:

```javascript
const { addBeforeLoader, loaderByName } = require("@craco/craco");

const myNewWebpackLoader = {
    loader: require.resolve("tslint-loader")
};

addBeforeLoader(webpackConfig, loaderByName("eslint-loader"), myNewWebpackLoader);
```

#### addBeforeLoaders

Add a new *loader* **before** all the loaders that match specified criteria to the webpack config.

Returns:

```javascript
{
    isAdded: true | false,
    addedCount: int
}
```

Usage:

```javascript
const { addBeforeLoaders, loaderByName } = require("@craco/craco");

const myNewWebpackLoader = {
    loader: require.resolve("tslint-loader")
};

addBeforeLoaders(webpackConfig, loaderByName("eslint-loader"), myNewWebpackLoader);
```

#### addAfterLoader

Add a new *loader* **after** the loader that match specified criteria to the webpack config.

Returns:

```javascript
{
    isAdded: true | false
}
```

Usage:

```javascript
const { addAfterLoader, loaderByName } = require("@craco/craco");

const myNewWebpackLoader = {
    loader: require.resolve("tslint-loader")
};

addAfterLoader(webpackConfig, loaderByName("eslint-loader"), myNewWebpackLoader);
```

#### addAfterLoaders

Add a new *loader* **after** all the loaders that match specified criteria to the webpack config.

Returns:

```javascript
{
    isAdded: true | false,
    addedCount: int
}
```

Usage:

```javascript
const { addAfterLoaders, loaderByName } = require("@craco/craco");

const myNewWebpackLoader = {
    loader: require.resolve("tslint-loader")
};

addAfterLoaders(webpackConfig, loaderByName("eslint-loader"), myNewWebpackLoader);
```

#### getPlugin

Retrieve the **first** plugin that match the specified criteria from the webpack config.

Returns:

```javascript
{
    isFound: true | false,
    match: {...} // the webpack plugin
}
```

Usage:

```javascript
const { getPlugin, pluginByName } = require("@craco/craco");

const { isFound, match } = getPlugin(webpackConfig, pluginByName("ESLintWebpackPlugin"));

if (isFound) {
    // do stuff...
}
```

#### removePlugins

Remove **all** the plugins that match the specified criteria from the webpack config.

Returns:

```javascript
{
    hasRemovedAny:: true | false,
    removedCount:: int
}
```

Usage:

```javascript
const { removePlugins, pluginByName } = require("@craco/craco");

removePlugins(webpackConfig, pluginByName("ESLintWebpackPlugin"));
```

#### addPlugins

Add new *plugins* to the webpack config.

Usage:

```javascript
const { addPlugins } = require("@craco/craco");

const myNewWebpackPlugin = require.resolve("ESLintWebpackPlugin");

addPlugins(webpackConfig, [myNewWebpackPlugin]);
addPlugins(webpackConfig, [ [myNewWebpackPlugin, "append"] ]);
addPlugins(webpackConfig, [ [myNewWebpackPlugin, "prepend"] ]);
```


#### throwUnexpectedConfigError

Throw an error if the webpack configuration changes and does not match your expectations. (For example, `getLoader` 
cannot find a loader and `isFound` is `false`.) `create-react-app` might update the structure of their webpack config, 
so it is very important to show a helpful error message when something breaks.

Raises an error and crashes Node.js:

```bash
$ yarn start
yarn run v1.12.3
$ craco start
/path/to/your/app/craco.config.js:23
            throw new Error(
            ^

    Error: Can't find eslint-loader in the webpack config!

    This error probably occurred because you updated react-scripts or craco. Please try updating craco-less to the latest version:

       $ yarn upgrade craco-less

    Or:

       $ npm update craco-less

    If that doesn't work, craco-less needs to be fixed to support the latest version.
    Please check to see if there's already an issue in the ndbroadbent/craco-less repo:

       * https://github.com/ndbroadbent/craco-less/issues?q=is%3Aissue+webpack+eslint-loader

    If not, please open an issue and we'll take a look. (Or you can send a PR!)

    You might also want to look for related issues in the craco and create-react-app repos:

       * https://github.com/sharegate/craco/issues?q=is%3Aissue+webpack+eslint-loader
       * https://github.com/facebook/create-react-app/issues?q=is%3Aissue+webpack+eslint-loader

    at throwUnexpectedConfigError (/path/to/your/app/craco.config.js:23:19)
    ...
```

Usage:

```javascript
const { getLoader, loaderByName, throwUnexpectedConfigError } = require("@craco/craco");

// Create a helper function if you need to call this multiple times
const throwError = (message, githubIssueQuery) =>
    throwUnexpectedConfigError({
        packageName: "craco-less",
        githubRepo: "ndbroadbent/craco-less",
        message,
        githubIssueQuery,
    });

const { isFound, match } = getLoader(webpackConfig, loaderByName("eslint-loader"));

if (!isFound) {
    throwError("Can't find eslint-loader in the webpack config!", "webpack+eslint-loader")
}
```

Options:

```javascript
{
    message: "An error message explaining what went wrong",
    packageName: "NPM package name",
    githubRepo: "GitHub repo where people can open an issue. Format: username/repo",
    githubIssueQuery: "Search string to find related issues"
}
```

> Only `message` is required.

## Backward Compatibility

CRACO is not meant to be backward compatible with older versions of react-scripts. This package will only support the latest version. If your project uses an old react-scripts version, refer to the following table to select the appropriate CRACO version.

| react-scripts Version |CRACO Version|
| --------------------- | -----------:|
| react-scripts < 4.0.0 |       5.8.0 |

## Debugging

### Verbose Logging

To activate **verbose** logging specify the CLI option `--verbose`

```javascript
/* package.json */

{
    "scripts": {
        "start": "craco start --verbose"
    }
}
```

## License

Copyright © 2020, Groupe Sharegate inc. This code is licensed under the Apache License, Version 2.0. You may obtain a 
copy of this license at https://github.com/gsoft-inc/gsoft-license/blob/master/LICENSE.
