import { atom } from "recoil";

const wsConnectionAtom = atom({
    key: "wsConnectionAtom",
    default: 0
});

export default wsConnectionAtom;
