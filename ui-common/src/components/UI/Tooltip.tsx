import {Tooltip as MUITooltip, Fade, TooltipProps as MUITooltipProps, makeStyles} from "@material-ui/core";
import React from "react";

export interface TooltipProps extends MUITooltipProps {
    variant?: 'default' | 'wide' | 'fit';
    isSimple?: boolean;
    classes?: any;
}

export type TooltipPlacement = 'bottom-end' | 'bottom-start' | 'bottom' | 'left-end' | 'left-start' | 'left' | 'right-end' | 'right-start' | 'right' | 'top-end' | 'top-start' | 'top';

const styles = {
    default: {
        maxWidth: 300
    },
    wide: {
        maxWidth: 700
    },
    fit: {
        maxWidth: '100%'
    }
};

const useStyles = makeStyles((theme) => styles);

const Tooltip: React.FC<TooltipProps> = (props) => {

    const {isSimple, ..._props} = props;

    const classes = useStyles(props.variant);

    const variant = props.variant || 'default';

    const backgroundClass = isSimple ? "" : "noBackground"

    return (
        <MUITooltip
            classes={{tooltip: `${backgroundClass} ` + classes[variant]}}
            interactive={true}
            enterDelay={200}
            TransitionComponent={Fade}
            {..._props}
        >
            {props.children || <div/>}
        </MUITooltip>
    );
};

export default Tooltip;