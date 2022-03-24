import {ReactComponent as DefaultIconDown} from './assets/default_icon_down.svg';
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
const classes = {icon: styles.icon, selectMenu: styles.list, select: styles.oasSelect, root:styles.root};

const defaultProps = {
    MenuProps: menuProps,
    IconComponent: DefaultIconDown
};

export interface SelectProps extends MUISelectProps {
    onChangeCb: (arg0: string) => void;
    value: string | string[];
    ellipsis?: boolean;
    labelOnTop?: boolean;
    className?: string;
    labelClassName?: string;
    trimItemsWhenMultiple?: boolean;
    transformDisplay?: (arg0: string) => string;
}

export const Select: React.FC<SelectProps> = ({
                                                  label,
                                                  value,
                                                  onChangeCb,
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

    const _onChange = ( item: { props: { value: any; }; }) => {
        const value = item.props.value;
        value === ALL_KEY ? onChangeCb(ALL_KEY) : onChangeCb(value);
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
                onChange: (e,data) => _onChange(data),
            })}
        >
            {multiple && <MenuItem key={ALL_KEY} value={ALL_KEY}>{transformItem(ALL_KEY)}</MenuItem>}
            {children}
        </MUISelect>
    </div>
}
