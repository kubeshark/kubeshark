import { atom } from "recoil";

const loggedInUserStateAtom = atom({
    key: "loggedInUserState",
    default: {
            "username": "",
            "role": "",
            "workspace": {
                "id": "",
                "name": "",
                "namespaces": []
        }
    }
});

export default loggedInUserStateAtom;
