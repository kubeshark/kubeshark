"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.TypeScriptLoader = void 0;
const ts_node_1 = require("ts-node");
const typescript_compile_error_1 = require("./typescript-compile-error");
function TypeScriptLoader(options) {
    return (path, content) => {
        try {
            // cosmiconfig requires the transpiled configuration to be CJS
            (0, ts_node_1.register)(Object.assign(Object.assign({}, options), { compilerOptions: { module: "commonjs" } })).compile(content, path);
            const result = require(path);
            // `default` is used when exporting using export default, some modules
            // may still use `module.exports` or if in TS `export = `
            return result.default || result;
        }
        catch (error) {
            if (error instanceof Error) {
                // Coerce generic error instance into typed error with better logging.
                throw typescript_compile_error_1.TypeScriptCompileError.fromError(error);
            }
            throw error;
        }
    };
}
exports.TypeScriptLoader = TypeScriptLoader;
