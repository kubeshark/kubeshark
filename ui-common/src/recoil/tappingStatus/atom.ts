import { atom } from "recoil";
import {TappingStatusPod} from "./index";

const tappingStatusAtom = atom({
    key: "tappingStatusAtom",
    default: null as TappingStatusPod[]
});

export default tappingStatusAtom;
