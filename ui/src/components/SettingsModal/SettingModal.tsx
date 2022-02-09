import React, {useEffect, useMemo, useState} from "react";
import {Modal, Backdrop, Fade, Box, Button} from "@material-ui/core";
import {modalStyle} from "../Filters";
import Checkbox from "../UI/Checkbox";
import './SettingsModal.sass';
import Api from "../../helpers/api";
import spinner from "../assets/spinner.svg";
import {useCommonStyles} from "../../helpers/commonStyle";
import {toast} from "react-toastify";

interface SettingsModalProps {
    isOpen: boolean
    onClose: () => void
    isFirstLogin: boolean
}

const api = Api.getInstance();

export const SettingsModal: React.FC<SettingsModalProps> = ({isOpen, onClose, isFirstLogin}) => {

    const classes = useCommonStyles();
    const [namespaces, setNamespaces] = useState({});
    const [isLoading, setIsLoading] = useState(false);
    const [searchValue, setSearchValue] = useState("");

    useEffect(() => {
        if(!isOpen) return;
        (async () => {
            try {
                setSearchValue("");
                setIsLoading(true);
                const tapConfig = await api.getTapConfig()
                if(isFirstLogin) {
                    const namespacesObj = {...tapConfig?.tappedNamespaces}
                    Object.keys(tapConfig?.tappedNamespaces ?? {}).forEach(namespace => {
                        namespacesObj[namespace] = true;
                    })
                    setNamespaces(namespacesObj);
                } else {
                    setNamespaces(tapConfig?.tappedNamespaces);
                }
            } catch (e) {
                console.error(e);
            } finally {
                setIsLoading(false);
            }
        })()
    }, [isFirstLogin, isOpen])

    const setAllNamespacesTappedValue = (isTap: boolean) => {
        const newNamespaces = {};
        Object.keys(namespaces).forEach(key => {
            newNamespaces[key] = isTap;
        })
        setNamespaces(newNamespaces);
    }

    const updateTappingSettings = async () => {
        try {
            await api.setTapConfig(namespaces);
            onClose();
            toast.success("Saved successfully");
        } catch (e) {
            console.error(e);
            toast.error("Something went wrong, changes may not have been saved.")
        }
    }

    const toggleTapNamespace = (namespace) => {
        const newNamespaces = {...namespaces};
        newNamespaces[namespace] = !namespaces[namespace]
        setNamespaces(newNamespaces);
    }

    const toggleAll = () => {
        const isChecked = Object.values(namespaces).every(tap => tap === true);
        setAllNamespacesTappedValue(!isChecked);
    }

    const filteredNamespaces = useMemo(() => {
        return Object.keys(namespaces).filter((namespace) => namespace.includes(searchValue));
    },[namespaces, searchValue])

    const buildNamespacesTable = () => {
        return <table cellPadding={5} style={{borderCollapse: "collapse"}}>
            <thead>
                <tr style={{borderBottomWidth: "2px"}}>
                    <th style={{width: 50}}><Checkbox checked={Object.values(namespaces).every(tap => tap === true)}
                        onToggle={toggleAll}/></th>
                    <th>Namespace</th>
                </tr>
            </thead>
            <tbody>
                {filteredNamespaces?.map(namespace => {
                    return <tr key={namespace}>
                        <td style={{width: 50}}>
                            <Checkbox checked={namespaces[namespace]} onToggle={() => toggleTapNamespace(namespace)}/>
                        </td>
                        <td>{namespace}</td>
                    </tr>
                }
                )}
            </tbody>
        </table>
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
                        Please choose from below the namespaces for tapping, traffic for namespaces selected will be displayed
                    </div>
                    {isLoading ? <div style={{textAlign: "center", padding: 20}}>
                        <img alt="spinner" src={spinner} style={{height: 35}}/>
                    </div> : <>
                        <div className="namespacesSettingsContainer">
                            <div style={{margin: "10px 0"}}>
                                <input className={classes.textField + " searchNamespace"} placeholder="Search" value={searchValue}
                                    onChange={(event) => setSearchValue(event.target.value)}/></div>
                            <div className="namespacesTable">
                                {buildNamespacesTable()}
                            </div>
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
