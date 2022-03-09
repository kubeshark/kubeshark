import type { TraceMap } from '@jridgewell/trace-mapping';
import type OriginalSource from './original-source';
import type { SourceMapSegmentObject } from './types';
declare type Sources = OriginalSource | SourceMapTree;
declare const INVALID_MAPPING: undefined;
declare const SOURCELESS_MAPPING: null;
declare type MappingSource = SourceMapSegmentObject | typeof INVALID_MAPPING | typeof SOURCELESS_MAPPING;
/**
 * traceMappings is only called on the root level SourceMapTree, and begins the process of
 * resolving each mapping in terms of the original source files.
 */
export declare let traceMappings: (tree: SourceMapTree) => TraceMap;
/**
 * SourceMapTree represents a single sourcemap, with the ability to trace
 * mappings into its child nodes (which may themselves be SourceMapTrees).
 */
export declare class SourceMapTree {
    map: TraceMap;
    sources: Sources[];
    constructor(map: TraceMap, sources: Sources[]);
    /**
     * originalPositionFor is only called on children SourceMapTrees. It recurses down
     * into its own child SourceMapTrees, until we find the original source map.
     */
    originalPositionFor(line: number, column: number, name: string): MappingSource;
}
export {};
