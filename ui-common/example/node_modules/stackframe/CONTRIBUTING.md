## Submitting bugs
Please include the following information to help us reproduce and fix:

* What you did
* What you expected to happen
* What actually happened
* Browser and version
* Example code that reproduces the problem (if possible)
* *(l33t mode)* A failing test

## Making contributions
Want to be listed as a *Contributor*? Make a pull request with: 

* Unit and/or functional tests that validate changes you're making.
* Run unit tests in the latest IE, Firefox, Chrome, Safari and Opera and make sure they pass.
* Rebase your changes onto origin/HEAD if you can do so cleanly.
* If submitting additional functionality, provide an example of how to use it.
* Please keep code style consistent with surrounding code.

## Dev Setup
* Make sure you have [NodeJS v16.x](https://nodejs.org/) installed
* Run `npm ci` from the project directory

## Linting
* Run `npm run lint` to run ESLint

## Testing
* (Local) Run `npm test`. Make sure [Karma Local Config](karma.conf.js) has the browsers you want
* (Any browser, remotely) If you have a [Sauce Labs](https://saucelabs.com) account, you can run `npm run test-ci`
 Make sure the target browser is enabled in [Karma CI Config](karma.conf.ci.js)
 Otherwise, GitHub Actions will run all browsers if you submit a Pull Request.
