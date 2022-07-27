import React from "react";

export interface Props {
    checked: boolean;
    onToggle: (checked: boolean) => any;
    disabled?: boolean;
}

const Radio: React.FC<Props> = ({ checked, onToggle, disabled, ...props }) => {
    return (
        <div>
            <input style={!disabled ? { cursor: "pointer" } : {}} type="radio" checked={checked} disabled={disabled} onChange={(event) => onToggle(event.target.checked)} {...props} />
        </div>
    );
};

export default Radio;
