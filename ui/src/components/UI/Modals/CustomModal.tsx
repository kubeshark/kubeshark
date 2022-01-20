import React from 'react';
import { makeStyles, withStyles, Modal, Backdrop, Fade, Box } from '@material-ui/core';
import {useCommonStyles} from "../../../helpers/commonStyle";
import { PropertiesTable } from 'redoc/typings/common-elements';

const useStyles = makeStyles({
    modal: {
        display: "flex",
        alignItems: "center",
        justifyContent: "center"
    },
    modalContents: {
        borderRadius: "5px",
        
        outline: "none",
        minWidth: "300px",
        backgroundColor: "rgb(255, 255, 255)"
    }
});





export interface CustomModalProps {
    open: boolean;
    children: React.ReactElement | React.ReactElement[];
    disableBackdropClick?: boolean;
    onClose?: () => void;
    className?: string;
    isPadding?: boolean;
    isWide? :boolean;
}

const CustomModal: React.FunctionComponent<CustomModalProps> = ({ open = false, onClose, disableBackdropClick = false, children, className}) => {
    const classes = useStyles({});
    const globals = useCommonStyles().modal
     

    const onModalClose = (reason) => {
        if(reason === 'backdropClick' && disableBackdropClick)
            return;   
        onClose();
    }

    return <Modal disableEnforceFocus open={open} onClose={(event, reason) => onModalClose(reason)}  className={`${classes.modal} ${className ?  className : ''}`}>
        <div className={`${classes.modalContents} ${globals}`} >
            <Fade in={open}>
                <Box>
                    {children}
                </Box>
            </Fade>
        </div>
    </Modal>
}

export default CustomModal;