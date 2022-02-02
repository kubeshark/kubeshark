import { atom } from "recoil";
import {Roles} from "./index";

const loggedInUserStateAtom = atom({
    key: "loggedInUserState",
    default: {
            "username": "",
            "role": "",
            "workspace": {
                "id": "",
                "name": "",
                "namespaces": [
                ]
        }
    }
});

export default loggedInUserStateAtom;
