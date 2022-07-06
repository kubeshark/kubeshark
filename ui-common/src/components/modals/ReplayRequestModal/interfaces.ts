export interface IReplayRequestData {
    method: string;
    hostPort: string;
    path: string;
    postData: string;
    headers: {
        key: string;
        value: unknown;
    }[]
    params: {
        key: string;
        value: unknown;
    }[]
}
