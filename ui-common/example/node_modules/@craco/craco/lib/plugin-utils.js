function gitHubIssueUrl(repo, query) {
    return `https://github.com/${repo}/issues?q=is%3Aissue${query ? `+${query}` : ""}`;
}

function showNpmPackageUrl(packageName) {
    return `\n   * https://www.npmjs.com/package/${packageName}\n\n`;
}

function showGitHubIssueUrl(repo, query) {
    return (
        `Please check to see if there's already an issue in the ${repo} repo:\n\n` +
        `   * ${gitHubIssueUrl(repo, query)}\n\n` +
        "If not, please open an issue and we'll take a look. (Or you can send a PR!)\n\n"
    );
}

function showPackageUpdateInstructions(packageName, repo, query) {
    return (
        `Please try updating ${packageName} to the latest version:\n\n` +
        `   $ yarn upgrade ${packageName}\n\n` +
        "Or:\n\n" +
        `   $ npm update ${packageName}\n\n` +
        `If that doesn't work, ${packageName} needs to be fixed to support the latest version.\n` +
        (repo ? showGitHubIssueUrl(repo, query) : showNpmPackageUrl(packageName))
    );
}

function throwUnexpectedConfigError({ message, packageName, githubRepo: repo, githubIssueQuery: query }) {
    throw new Error(
        `${message}\n\n` +
            "This error probably occurred because you updated react-scripts or craco. " +
            (packageName
                ? showPackageUpdateInstructions(packageName, repo, query)
                : "You will need to update this plugin to work with the latest version.\n\n") +
            "You might also want to look for related issues in the " +
            "craco and create-react-app repos:\n\n" +
            `   * ${gitHubIssueUrl("sharegate/craco", query)}\n` +
            `   * ${gitHubIssueUrl("facebook/create-react-app", query)}\n`
    );
}

module.exports = {
    gitHubIssueUrl,
    throwUnexpectedConfigError
};
