import { atom } from "recoil";
import TrafficViewerApi from "../../components/TrafficViewer/TrafficViewerApi";

const TrafficViewerApiAtom = atom({
    key: "TrafficViewerApiAtom",
    default: {} as TrafficViewerApi
});

export default TrafficViewerApiAtom;
