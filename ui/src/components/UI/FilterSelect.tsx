import React from "react";
import { MenuItem } from '@material-ui/core';
import style from './style/FilterSelect.module.sass';
import { Select, SelectProps } from "./Select";

interface FilterSelectProps extends SelectProps {
    items: string[];
    value: string | string[];
    onChange: (string) => void;
    label?: string;
    allowMultiple?: boolean;
    transformDisplay?: (string) => string;
}

export const FilterSelect: React.FC<FilterSelectProps> = ({items, value, onChange, label, allowMultiple= false, transformDisplay}) => {
    return <Select
        value={value}
        multiple={allowMultiple}
        label={label}
        onChange={onChange}
        transformDisplay={transformDisplay}
        labelOnTop={true}
        labelClassName={style.SelectLabel}
        trimItemsWhenMultiple={true}
    >
        {items?.map(item => <MenuItem key={item} value={item}><span className='uppercase'>{item}</span></MenuItem>)}
    </Select>
};
