import {selector} from "recoil";
import tappingStatusAtom from "./atom";

const tappingStatusDetails = selector({
    key: 'tappingStatusDetails',
    get: ({get}) => {
        const tappingStatus = get(tappingStatusAtom);
        const uniqueNamespaces = Array.from(new Set(tappingStatus.map(pod => pod.namespace)));
        const amountOfPods = tappingStatus.length;
        const amountOfTappedPods = tappingStatus.filter(pod => pod.isTapped).length;
        const amountOfUntappedPods = amountOfPods - amountOfTappedPods;

        return {
            uniqueNamespaces,
            amountOfPods,
            amountOfTappedPods,
            amountOfUntappedPods,
        };
    },
});

export default tappingStatusDetails;
