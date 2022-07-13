const { defineConfig } = require('cypress')

module.exports = defineConfig({
  watchForFileChanges: false,
  viewportWidth: 1920,
  viewportHeight: 1080,
  video: false,
  screenshotOnRunFailure: false,
  defaultCommandTimeout: 6000,
  env: {
    testUrl: 'http://localhost:8899/',
    redactHeaderContent: 'User-Header[REDACTED]',
    redactBodyContent: '{ "User": "[REDACTED]" }',
    greenFilterColor: 'rgb(210, 250, 210)',
    redFilterColor: 'rgb(250, 214, 220)',
    bodyJsonClass: '.hljs',
    mizuWidth: 1920,
    normalMizuHeight: 1080,
    hugeMizuHeight: 3500,
  },
  e2e: {
    // We've imported your old cypress plugins here.
    // You may want to clean this up later by importing these.
    // setupNodeEvents(on, config) {
    //   return require('./cypress/plugins/index.js')(on, config)
    // },
    specPattern: 'cypress/e2e/tests/*.js',
    supportFile: false
  },
})
