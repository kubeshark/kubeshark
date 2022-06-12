export interface GraphData {
    nodes: Node[];
    edges: Edge[];
}

export interface Node {
    id: number;
    value: number;
    label: string;
    title?: string;
    color?: object;
}

export interface Edge {
    from: number;
    to: number;
    value: number;
    label: string;
    title?: string;
    color?: object;
}

export interface ServiceMapNode {
    id: number;
    name: string;
    entry: Entry;
    count: number;
    resolved: boolean;
}

export interface ServiceMapEdge {
    source: ServiceMapNode;
    destination: ServiceMapNode;
    count: number;
    protocol: Protocol;
}

export interface ServiceMapGraph {
    nodes: ServiceMapNode[];
    edges: ServiceMapEdge[];
}

export interface Entry {
    ip: string;
    port: string;
    name: string;
}

export interface Protocol {
    name: string;
    abbr: string;
    macro: string;
    version: string;
    backgroundColor: string;
    foregroundColor: string;
    fontSize: number;
    referenceLink: string;
    ports: string[];
    priority: number;
}