import type { SourceMapSegmentObject } from './types';
/**
 * A "leaf" node in the sourcemap tree, representing an original, unmodified
 * source file. Recursive segment tracing ends at the `OriginalSource`.
 */
export default class OriginalSource {
    content: string | null;
    source: string;
    constructor(source: string, content: string | null);
    /**
     * Tracing a `SourceMapSegment` ends when we get to an `OriginalSource`,
     * meaning this line/column location originated from this source file.
     */
    originalPositionFor(line: number, column: number, name: string): SourceMapSegmentObject;
}
