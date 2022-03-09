"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.TypeScriptCompileError = void 0;
class TypeScriptCompileError extends Error {
    constructor(message) {
        super(message);
        this.name = this.constructor.name;
        // https://github.com/Microsoft/TypeScript-wiki/blob/main/Breaking-Changes.md#extending-built-ins-like-error-array-and-map-may-no-longer-work
        Object.setPrototypeOf(this, new.target.prototype);
    }
    static fromError(error) {
        const errMsg = error.message.replace(/(TypeScript compiler encountered syntax errors while transpiling\. Errors:\s?)|(⨯ Unable to compile TypeScript:\s?)/, "");
        const message = `TypeScriptLoader failed to compile TypeScript:\n${errMsg}`;
        const newError = new TypeScriptCompileError(message);
        newError.stack = error.stack;
        return newError;
    }
    /**
     * Support legacy usage of this method.
     * @deprecated
     */
    toObject() {
        return {
            message: this.message,
            name: this.name,
            stack: this.stack,
        };
    }
}
exports.TypeScriptCompileError = TypeScriptCompileError;
