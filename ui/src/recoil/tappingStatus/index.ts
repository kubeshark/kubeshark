import atom from "./atom";
import tappingStatusDetails from './details';

interface TappingStatusPod {
    name: string;
    namespace: string;
    isTapped: boolean;
}

interface TappingStatus {
    pods: TappingStatusPod[];
}

export type {TappingStatus, TappingStatusPod};
export {tappingStatusDetails};

export default atom;
