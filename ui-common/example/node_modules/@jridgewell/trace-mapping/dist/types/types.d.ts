export interface SourceMapV3 {
    file?: string | null;
    names: string[];
    sourceRoot?: string;
    sources: (string | null)[];
    sourcesContent?: (string | null)[];
    version: 3;
}
declare type Column = number;
declare type SourcesIndex = number;
declare type SourceLine = number;
declare type SourceColumn = number;
declare type NamesIndex = number;
export declare type SourceMapSegment = [Column] | [Column, SourcesIndex, SourceLine, SourceColumn] | [Column, SourcesIndex, SourceLine, SourceColumn, NamesIndex];
export interface EncodedSourceMap extends SourceMapV3 {
    mappings: string;
}
export interface DecodedSourceMap extends SourceMapV3 {
    mappings: SourceMapSegment[][];
}
export declare type OriginalMapping = {
    source: string | null;
    line: number;
    column: number;
    name: string | null;
};
export declare type InvalidMapping = {
    source: null;
    line: null;
    column: null;
    name: null;
};
export declare type SourceMapInput = string | EncodedSourceMap | DecodedSourceMap;
export declare type Needle = {
    line: number;
    column: number;
};
export declare type EachMapping = {
    generatedLine: number;
    generatedColumn: number;
    source: null;
    originalLine: null;
    originalColumn: null;
    name: null;
} | {
    generatedLine: number;
    generatedColumn: number;
    source: string | null;
    originalLine: number;
    originalColumn: number;
    name: string | null;
};
export declare abstract class SourceMap {
    version: SourceMapV3['version'];
    file: SourceMapV3['file'];
    names: SourceMapV3['names'];
    sourceRoot: SourceMapV3['sourceRoot'];
    sources: SourceMapV3['sources'];
    sourcesContent: SourceMapV3['sourcesContent'];
    resolvedSources: SourceMapV3['sources'];
}
export {};
