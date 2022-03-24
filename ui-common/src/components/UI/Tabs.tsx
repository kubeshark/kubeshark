import Tooltip from "./Tooltip";
import React from "react";
import { makeStyles, Theme,createStyles } from '@material-ui/core/styles';
import variables from '../../variables.module.scss';

interface Tab {
    tab: string,
    disabled?: boolean,
    disabledMessage?: string,
    highlight?: boolean,
    badge?: any,
}

interface Props {
    classes?: any,
    tabs: Tab[],
    currentTab: string,
    color?: string,
    onChange: (string) => void,
    leftAligned?: boolean,
    dark?: boolean,
}

const useTabsStyles = makeStyles((theme : Theme) => createStyles({
    root: {
        height: 40,
        paddingTop: 15
    },

    tab: {
        display: 'inline-block',
        textTransform: 'uppercase',
        color: variables.blueColor,
        cursor: 'pointer',
    },

    tabsAlignLeft: {
        textAlign: 'left'
    },

    active: {
        fontWeight: 400,
        color: variables.fontColor,
        cursor: 'unset',
        borderBottom: "2px solid " + variables.fontColor,
        paddingBottom: 6,

        "&.dark": {
            color: theme.palette.common.black,
            borderBottom: "2px solid " + theme.palette.common.black,
        }
    },

    disabled: {
        color: theme.palette.primary.dark,
        cursor: 'unset'
    },

    highlight: {
        color: theme.palette.primary.light,
    },

    separator: {
        borderRight: "1px solid " + theme.palette.primary.dark,
        height: 20,
        verticalAlign: 'middle',
        margin: '0 20px',
        cursor: 'unset',
    }
}));


const Tabs: React.FC<Props> = ({classes={}, tabs, currentTab, color, onChange, leftAligned, dark}) => {
    const _classes = {...useTabsStyles(), ...classes};
    return <div className={`${_classes.root} ${leftAligned ? _classes.tabsAlignLeft : ''}`}>
        {tabs.map(({tab, disabled, disabledMessage, highlight, badge}, index) => {
            const active = currentTab === tab;
            const tabLink = <span
                data-cy={"tab-" + tab}
                key={tab}
                className={`${_classes.tab} ${active ? _classes.active : ''} ${disabled ? _classes.disabled : ''} ${highlight ? _classes.highlight : ''} ${dark ? 'dark' : ''}`}
                onClick={() => !disabled && onChange(tab)}
                style={{color: color}}
            >
                {tab}

                {React.isValidElement(badge) && badge}
            </span>;

            return <span key={tab}>
                {disabled && disabledMessage ? <Tooltip title={disabledMessage} isSimple>{tabLink}</Tooltip> : tabLink}
                {index < tabs.length - 1 && <span className={_classes.tab + ' ' + _classes.separator} key={tab + '_sepparator'}></span>}
            </span>;
        })}
    </div>;
}

export default Tabs;
