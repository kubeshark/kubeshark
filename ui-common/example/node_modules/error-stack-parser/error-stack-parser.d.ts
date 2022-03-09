// Type definitions for ErrorStackParser v2.0.0
// Project: https://github.com/stacktracejs/error-stack-parser
// Definitions by: Eric Wendelin <https://www.eriwen.com>
// Definitions: https://github.com/DefinitelyTyped/DefinitelyTyped

import StackFrame = require("stackframe");

declare module ErrorStackParser {
    /**
     * Given an Error object, extract the most information from it.
     *
     * @param {Error} error object
     * @return {Array} of StackFrames
     */
    export function parse(error: Error): StackFrame[];
}

export = ErrorStackParser;
