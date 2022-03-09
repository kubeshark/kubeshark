import React from "react";

export interface Props {
    checked: boolean;
    onToggle: (checked:boolean) => any;
}

const Checkbox: React.FC<Props> = ({checked, onToggle}) => {

    return (
        <div>
            <input style={{cursor: "pointer"}} type="checkbox" checked={checked} onChange={(event) => onToggle(event.target.checked)}/>
        </div>
    );
};

export default Checkbox;
