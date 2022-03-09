import type { TraceMap } from '@jridgewell/trace-mapping';
import type { DecodedSourceMap, RawSourceMap, Options } from './types';
/**
 * A SourceMap v3 compatible sourcemap, which only includes fields that were
 * provided to it.
 */
export default class SourceMap {
    file?: string | null;
    mappings: RawSourceMap['mappings'] | DecodedSourceMap['mappings'];
    sourceRoot?: string;
    names: string[];
    sources: (string | null)[];
    sourcesContent?: (string | null)[];
    version: 3;
    constructor(map: TraceMap, options: Options);
    toString(): string;
}
