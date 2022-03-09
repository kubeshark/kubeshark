import { atom } from "recoil";

const wsConnectionAtom = atom({
    key: "wsConnectionAtom",
    default: 0
});

type closeWsConnectionCallback = {closeCallback : () => {}}
//closeCallback
export const closeWsConnectionCallbackAtom = atom({
    key: "closeWsConnectionCallbackAtom",
    default:  {} as closeWsConnectionCallback
})

export default wsConnectionAtom;
