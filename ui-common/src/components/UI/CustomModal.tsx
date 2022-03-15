import React from 'react';
import { makeStyles, Modal, Backdrop, Fade, Box } from '@material-ui/core';
import {useCommonStyles} from "../../helpers/commonStyle";

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
        backgroundColor: "rgb(255, 255, 255)",       
    },
    modalBackdrop :{
        background : "rgba(24, 51, 121, 0.8)"
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

    return <Modal disableEnforceFocus open={open} onClose={(event, reason) => onModalClose(reason)}
                  className={`${classes.modal}`}     
                  closeAfterTransition
                  BackdropComponent={Backdrop}
                  BackdropProps={{
                      timeout: 500,
                      className:`${classes.modalBackdrop}`
                  }}>
                <div className={`${classes.modalContents} ${globals} ${className ?  className : ''}`} >
                    <Fade in={open}>
                        <Box>
                            {children}
                        </Box>
                    </Fade>
                </div>
            </Modal>
}

export default CustomModal;