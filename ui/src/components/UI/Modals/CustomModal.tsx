import React from 'react';
import { makeStyles, withStyles, Modal, Backdrop } from '@material-ui/core';

const useStyles = makeStyles({
    modal: {
        display: "flex",
        alignItems: "center",
        justifyContent: "center"
    },
    modalContents: {
        borderRadius: "8px",
        position: "relative",
        outline: "none",
        minWidth: "300px",
        backgroundColor: "#171C30"
    },
    paddingModal: {
        padding: "20px"
    },
    modalControl: {
        width: "250px",
        margin: "20px",
    },
    wideModal: {
        width: "50vw",
        maxWidth: 700,
    },

    modalButton: {
        margin: "0 auto"
    },
});

const MyBackdrop = withStyles({
    root: {
        backgroundColor: '#090b14e6'
    },
})(Backdrop);

export interface CustomModalProps {
    open: boolean;
    children: React.ReactElement | React.ReactElement[];
    disableBackdropClick?: boolean;
    onClose?: () => void;
    className?: string;
    isPadding?: boolean;
    isWide? :boolean;
}

const CustomModal: React.FunctionComponent<CustomModalProps> = ({ open = false, onClose, disableBackdropClick = false, children, className, isPadding, isWide }) => {
    const classes = useStyles({});

    return <Modal disableEnforceFocus open={open} onClose={onClose} disableBackdropClick={disableBackdropClick} className={`${classes.modal} ${className ?  className : ''}`} BackdropComponent={MyBackdrop}>
        <div className={`${classes.modalContents} ${isPadding ?  classes.paddingModal : ''} ${isWide ?  classes.wideModal : ''}`} >
            {children}
        </div>
    </Modal>
}

export default CustomModal;