#!/usr/bin/env node

const spawn = require("cross-spawn");

const args = process.argv.slice(2);
const scriptIndex = args.findIndex(x => x === "build" || x === "start" || x === "test");
const script = scriptIndex === -1 ? args[0] : args[scriptIndex];

switch (script) {
    case "build":
    case "start":
    case "test": {
        const nodeArgs = scriptIndex > 0 ? args.slice(0, scriptIndex) : [];
        const scriptPath = require.resolve(`../scripts/${script}`);
        const scriptArgs = args.slice(scriptIndex + 1);
        const processArgs = nodeArgs.concat(scriptPath).concat(scriptArgs);

        const child = spawn.sync("node", processArgs, { stdio: "inherit" });

        if (child.signal) {
            if (child.signal === "SIGKILL") {
                console.log(`
                    The build failed because the process exited too early.
                    This probably means the system ran out of memory or someone called
                    \`kill -9\` on the process.
                `);
            } else if (child.signal === "SIGTERM") {
                console.log(`
                    The build failed because the process exited too early.
                    Someone might have called  \`kill\` or \`killall\`, or the system could
                    be shutting down.
                `);
            }

            process.exit(1);
        }

        process.exit(child.status);
        break;
    }
    default:
        console.log(`Unknown script "${script}".`);
        console.log("Perhaps you need to update craco?");
        break;
}
