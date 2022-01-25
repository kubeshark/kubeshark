import React from "react";

export interface Props {
    checked: boolean;
    onToggle: (checked:boolean) => any;
}

const Radio: React.FC<Props> = ({checked, onToggle}) => {

    return (
        <div>
            <input style={{cursor: "pointer"}} type="radio" checked={checked} onChange={(event) => onToggle(event.target.checked)}/>
        </div>
    );
};

export default Radio;
