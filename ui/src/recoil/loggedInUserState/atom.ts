import { atom } from "recoil";
import {Roles} from "./index";

const loggedInUserStateAtom = atom({
    key: "loggedInUserState",
    default: Roles.unAuthorise
});

export default loggedInUserStateAtom;
