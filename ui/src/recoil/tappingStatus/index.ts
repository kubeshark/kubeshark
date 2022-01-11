import atom from "./atom"

interface TappingStatusPod {
    name: string;
    namespace: string;
    isTapped: boolean;
}

interface TappingStatus {
    pods: TappingStatusPod[];
}

export type {TappingStatus, TappingStatusPod};

export default atom;
