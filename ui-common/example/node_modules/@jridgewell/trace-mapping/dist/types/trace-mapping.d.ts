import type { SourceMapV3, DecodedSourceMap, EncodedSourceMap, InvalidMapping, OriginalMapping, SourceMapSegment, SourceMapInput, Needle, SourceMap, EachMapping } from './types';
export type { SourceMapSegment, SourceMapInput, DecodedSourceMap, EncodedSourceMap, InvalidMapping, OriginalMapping as Mapping, OriginalMapping, EachMapping, } from './types';
/**
 * Returns the encoded (VLQ string) form of the SourceMap's mappings field.
 */
export declare let encodedMappings: (map: TraceMap) => EncodedSourceMap['mappings'];
/**
 * Returns the decoded (array of lines of segments) form of the SourceMap's mappings field.
 */
export declare let decodedMappings: (map: TraceMap) => Readonly<DecodedSourceMap['mappings']>;
/**
 * A low-level API to find the segment associated with a generated line/column (think, from a
 * stack trace). Line and column here are 0-based, unlike `originalPositionFor`.
 */
export declare let traceSegment: (map: TraceMap, line: number, column: number) => Readonly<SourceMapSegment> | null;
/**
 * A higher-level API to find the source/line/column associated with a generated line/column
 * (think, from a stack trace). Line is 1-based, but column is 0-based, due to legacy behavior in
 * `source-map` library.
 */
export declare let originalPositionFor: (map: TraceMap, needle: Needle) => OriginalMapping | InvalidMapping;
/**
 * Iterates each mapping in generated position order.
 */
export declare let eachMapping: (map: TraceMap, cb: (mapping: EachMapping) => void) => void;
/**
 * A helper that skips sorting of the input map's mappings array, which can be expensive for larger
 * maps.
 */
export declare let presortedDecodedMap: (map: DecodedSourceMap, mapUrl?: string) => TraceMap;
export declare class TraceMap implements SourceMap {
    version: SourceMapV3['version'];
    file: SourceMapV3['file'];
    names: SourceMapV3['names'];
    sourceRoot: SourceMapV3['sourceRoot'];
    sources: SourceMapV3['sources'];
    sourcesContent: SourceMapV3['sourcesContent'];
    resolvedSources: string[];
    private _encoded;
    private _decoded;
    private _binarySearchMemo;
    constructor(map: SourceMapInput, mapUrl?: string | null);
}
