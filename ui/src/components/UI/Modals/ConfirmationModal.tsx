import React, {ReactElement} from 'react';
import iconClose from '../../assets/closeIcon.svg';
import CustomModal from './CustomModal';
import {observer} from 'mobx-react-lite';
import {Button} from "@material-ui/core";
import './ConfirmationModal.sass';
import spinner from "../../assets/spinner.svg";
import {useCommonStyles} from "../../../helpers/commonStyle";

interface ConfirmationModalProps {
    title?: string;
    isOpen: boolean;
    onClose: () => void;
    onConfirm: () => void;
    closeButtonText?: string;
    confirmButtonText?: any;
    subContent?: string;
    confirmDisabled?: boolean;
    isWide?: boolean;
    confirmButtonColor?: string;
    titleColor?: string;
    img?: ReactElement;
    isLoading?: boolean;
    className?: any;
}

const ConfirmationModal: React.FC<ConfirmationModalProps> = observer(({title, isOpen, onClose, onConfirm, confirmButtonText,
                                                                        closeButtonText, subContent, confirmDisabled = false, isWide,
                                                                        confirmButtonColor, titleColor, img, isLoading,children,
                                                                        className}) => {
    const classes = useCommonStyles();
    const confirmStyle = {width: 100, marginLeft: 20}                                                                  
    return (
        <CustomModal open={isOpen} onClose={onClose} disableBackdropClick={true} isWide={isWide} className={`${className}`}>
            <div className="confirmationHeader">
                <div className="confirmationTitle" style={titleColor ? {color: titleColor} : {}}>{title ?? "CONFIRMATION"}</div>
                <img src={iconClose} onClick={onClose} alt="close"/>
                
            </div>
            <div className="confirmationText" style={img ? {display: "flex", alignItems: "center"} : {}}>
                {img && <div style={{paddingRight: 20}}>
                    {img}
                </div>}
                <div>
                    {children}
                    {subContent && <div style={{marginTop: 10, color: "#FFFFFF80"}}>
                        {subContent}
                    </div>}
                </div>
            </div>

            <div className="confirmationActions">
                <Button disabled={isLoading} style={{width: 100}} className={classes.outlinedButton} size={"small"}
                            variant='outlined' onClick={onClose}>{closeButtonText ?? "CANCEL"}
                </Button>
                <Button style={confirmButtonColor ? {backgroundColor: confirmButtonColor,...confirmStyle} : {...confirmStyle}} 
                    className={classes.button} size={"small"}
                    onClick={onConfirm} 
                    disabled={confirmDisabled || isLoading} 
                    endIcon={isLoading && <img src={spinner} alt="spinner"/>}>{confirmButtonText ?? "YES"}
                </Button>
            </div>
        </CustomModal>
    )
});

export default ConfirmationModal;