import {ReactComponent as DefaultIconDown} from '../assets/default_icon_down.svg';
import {MenuItem, Select as MUISelect} from '@material-ui/core';
import React from 'react';
import {SelectProps as MUISelectProps} from '@material-ui/core/Select/Select';
import styles from './style/Select.module.sass';

export const ALL_KEY= 'All';

const menuProps: any = {
    anchorOrigin: {
        vertical: "bottom",
        horizontal: "left"
    },
    transformOrigin: {
        vertical: "top",
        horizontal: "left"
    },
    getContentAnchorEl: null
};

// icons styles are not overwritten from the Props, only as a separate object
const classes = {icon: styles.icon, selectMenu: styles.list};

const defaultProps = {
    MenuProps: menuProps,
    IconComponent: DefaultIconDown
};

export interface SelectProps extends MUISelectProps {
    onChange: (string) => void;
    value: string | string[];
    ellipsis?: boolean;
    labelOnTop?: boolean;
    className?: string;
    labelClassName?: string;
    trimItemsWhenMultiple?: boolean;
    transformDisplay?: (string) => string;
}

export const Select: React.FC<SelectProps> = ({
                                                  label,
                                                  value,
                                                  onChange,
                                                  transformDisplay,
                                                  ellipsis = true,
                                                  multiple,
                                                  labelOnTop = false,
                                                  children,
                                                  className,
                                                  labelClassName,
                                                  trimItemsWhenMultiple,
                                                  ...props
                                              }) => {
    let _value = value;

    const _onChange = (_, item) => {
        const value = item.props.value;
        value === ALL_KEY ? onChange(ALL_KEY) : onChange(value);
    }

    if (multiple && (!_value || _value.length === 0)) _value = [ALL_KEY];

    const transformItem: (i: string) => string = transformDisplay ? transformDisplay : i => i;

    const renderValue = multiple
        ? (item: any[]) => <span className={ellipsis ? 'ellipsis' : ''}>{
            trimItemsWhenMultiple && item.length > 1 ?
                transformItem(`${item[item.length-1]} (+${item.length - 1})`):
                item?.map(transformItem).join(",")
        }</span>
        : null;

    return <div className={`select ${labelOnTop ? 'labelOnTop' : ''} ${className ? className : ''}`}>
        {label && <div className={`selectLabel ${labelClassName ? labelClassName : ''}`}>{label}</div>}
        <MUISelect
            {...Object.assign({}, defaultProps, props, {
                classes,
                value: _value,
                renderValue,
                multiple,
                onChange: _onChange,
            })}
        >
            {multiple && <MenuItem key={ALL_KEY} value={ALL_KEY}>{transformItem(ALL_KEY)}</MenuItem>}
            {children}
        </MUISelect>
    </div>
}
