import {useState} from "react";

export default function useToggle(initialState: boolean = false): [boolean, () => void] {

    const [isToggled, setToggled] = useState(initialState);

    return [isToggled, () => {
        setToggled(!isToggled)
    }];
}