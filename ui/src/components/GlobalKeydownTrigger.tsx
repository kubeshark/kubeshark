import React, {useEffect} from "react";

interface GlobalKeydownTriggerProps {
    onKeyDown: (key: string) => void;
}

const GlobalKeydownTrigger: React.FC<GlobalKeydownTriggerProps> = ({onKeyDown}) => {
    // @ts-ignore
    useEffect(() => {
        const keyDownHandler = e => onKeyDown(e.key);   
        document.addEventListener("keydown", keyDownHandler);
        return () => {
            //clean up listener on unmount or on callback change
            document.removeEventListener("keydown", keyDownHandler);
        }
    }, [onKeyDown]);

    return <></>
};

export default GlobalKeydownTrigger;
