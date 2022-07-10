export interface KeyValuePair {
    key: string;
    value: unknown;
}

export interface IReplayRequestData {
    method: string;
    hostPort: string;
    path: string;
    postData: string;
    headers: KeyValuePair[]
    params: KeyValuePair[]
}
