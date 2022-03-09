export declare class TypeScriptCompileError extends Error {
    constructor(message: string);
    static fromError(error: Error): TypeScriptCompileError;
    /**
     * Support legacy usage of this method.
     * @deprecated
     */
    toObject(): {
        message: string;
        name: string;
        stack: string | undefined;
    };
}
