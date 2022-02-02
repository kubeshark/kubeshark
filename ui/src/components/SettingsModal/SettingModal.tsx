import React, {useEffect, useState} from "react";
import {Modal, Backdrop, Fade, Box, Button} from "@material-ui/core";
import {modalStyle} from "../Filters";
import './SettingsModal.sass';
import Api from "../../helpers/api";
import spinner from "../assets/spinner.svg";
import {useCommonStyles} from "../../helpers/commonStyle";
import {toast} from "react-toastify";
import SelectList from "../UI/SelectList";
import { adminUsername } from "../../consts";

interface SettingsModalProps {
    isOpen: boolean
    onClose: () => void
    isFirstLogin: boolean
}

const api = Api.getInstance();

export const SettingsModal: React.FC<SettingsModalProps> = ({isOpen, onClose, isFirstLogin}) => {

    const classes = useCommonStyles();
    const [namespaces, setNamespaces] = useState([]);
    const [checkedNamespacesKeys, setCheckedNamespacesKeys] = useState([]);
    const [isLoading, setIsLoading] = useState(false);
    const [searchValue, setSearchValue] = useState("");

    useEffect(() => {
        if(!isOpen) return;
        (async () => {
            try {
                setSearchValue("");
                setIsLoading(true);
                // const tapConfig = await api.getTapConfig()
                const namespaces = await api.getNamespaces();
                const namespacesMapped = namespaces.map(namespace => {
                    return {key: namespace, value: namespace}
                  })
                setNamespaces(namespacesMapped);
                // if(isFirstLogin) {
                //     const namespacesObj = {...tapConfig?.tappedNamespaces}
                //     Object.keys(tapConfig?.tappedNamespaces ?? {}).forEach(namespace => {
                //         namespacesObj[namespace] = true;
                //     })
                //     setNamespaces(namespacesObj);
                // } else {
                //     setNamespaces(tapConfig?.tappedNamespaces);
                // }
            } catch (e) {
                console.error(e);
            } finally {
                setIsLoading(false);
            }
        })()
    }, [isFirstLogin, isOpen])

    const updateTappingSettings = async () => {
        try {
            const defaultWorkspace = {
                name: "default",
                namespaces: checkedNamespacesKeys
            }
            await api.createWorkspace(defaultWorkspace,adminUsername);
            onClose();
            toast.success("Saved successfully");
        } catch (e) {
            console.error(e);
            toast.error("Something went wrong, changes may not have been saved.")
        }
    }

    const onModalClose = (reason) => {
        if(reason === 'backdropClick' && isFirstLogin) return;
        onClose();
    }

    return <Modal
        open={isOpen}
        onClose={(event, reason) => onModalClose(reason)}
        closeAfterTransition
        BackdropComponent={Backdrop}
        BackdropProps={{
            timeout: 500,
        }}
        style={{overflow: 'auto'}}
    >
        <Fade in={isOpen}>
            <Box sx={modalStyle} style={{width: "40vw", maxWidth: 600, height: "70vh", padding: 0, display: "flex", justifyContent: "space-between", flexDirection: "column"}}>
                <div style={{padding: 32, paddingBottom: 0}}>
                    <div className="settingsTitle">Tapping Settings</div>
                    <div className="settingsSubtitle" style={{marginTop: 20}}>
                        Please choose from below the namespaces for tapping, traffic for namespaces selected will be displayed as default workspace.
                    </div>
                    {isLoading ? <div style={{textAlign: "center", padding: 20}}>
                        <img alt="spinner" src={spinner} style={{height: 35}}/>
                    </div> : <>
                        <div className="namespacesSettingsContainer">
                            <div style={{margin: "10px 0"}}>
                                <input className={classes.textField + " searchNamespace"} placeholder="Search" value={searchValue}
                                       onChange={(event) => setSearchValue(event.target.value)}/></div>
                                <SelectList items={namespaces} tableName={'Namespace'} multiSelect={true} searchValue={searchValue} setCheckedValues={setCheckedNamespacesKeys} tabelClassName={'namespacesTable'} checkedValues={checkedNamespacesKeys}/>
                        </div>
                    </>}
                </div>
                <div className="settingsActionsContainer">
                    {!isFirstLogin &&
                    <Button style={{width: 100}} className={classes.outlinedButton} size={"small"}
                            onClick={onClose} variant='outlined'>Cancel</Button>}
                    <Button style={{width: 100, marginLeft: 20}} className={classes.button} size={"small"}
                            onClick={updateTappingSettings}>OK</Button>
                </div>
            </Box>
        </Fade>
    </Modal>
}
