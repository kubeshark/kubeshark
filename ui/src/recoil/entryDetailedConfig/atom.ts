import { atom } from "recoil";

const entryDetailedConfigAtom = atom({
    key: "entryDetailedConfigAtom",
    default: null
});

export type EntryDetailedConfig = {
    isReplayEnabled: boolean
}

export default entryDetailedConfigAtom;
