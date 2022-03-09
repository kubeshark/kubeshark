const fs = require("fs");
const { log } = require("./logger");

const projectRoot = fs.realpathSync(process.cwd());

log("Project root path resolved to: ", projectRoot);

module.exports = {
    projectRoot
};
