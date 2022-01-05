import React, {useEffect, useState} from "react";
import {Modal, Backdrop, Fade, Box, Button, makeStyles} from "@material-ui/core";
import {modalStyle} from "../Filters";
import Checkbox from "../UI/Checkbox";
import './SettingsModal.sass';
import Api from "../../helpers/api";
import spinner from "../assets/spinner.svg";
import {useCommonStyles} from "../../helpers/commonStyle";

interface SettingsModalProps {
    isOpen: boolean
    onClose: () => void
}


const api = new Api();

export const SettingsModal: React.FC<SettingsModalProps> = ({isOpen, onClose}) => {

    const classes = useCommonStyles();
    const [namespaces, setNamespaces] = useState({aa: true, bb: false} as any);
    const [isLoading, setIsLoading] = useState(false);
    const [searchValue, setSearchValue] = useState("");
    const isFirstLogin = false;

    useEffect(() => {
        (async () => {
            try {
                setIsLoading(true);
                const tapConfig = await api.getTapConfig()
                setNamespaces(tapConfig?.tappedNamespaces);
            } catch (e) {
                console.error(e);
            } finally {
                setIsLoading(false);
            }
        })()
    }, [])

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
        } catch (e) {
            console.error(e);
        }
    }

    const toggleTapNamespace = (namespace) => {
        const newNamespaces = {...namespaces};
        newNamespaces[namespace] = !namespaces[namespace]
        setNamespaces(newNamespaces);
    }

    const toggleAll = () => {
        const isChecked = Object.values(namespaces).every(tap => tap === true);
        isChecked ? setAllNamespacesTappedValue(false) : setAllNamespacesTappedValue(true);
    }

    const buildNamespacesTable = () => {
        return <table cellPadding={5} style={{borderCollapse: "collapse"}}>
            <thead>
            <tr style={{borderBottomWidth: "2px"}}>
                {/*<th style={{width: 50, textAlign: "left"}}>Tap</th>*/}
                <th style={{width: 50}}><Checkbox checked={Object.values(namespaces).every(tap => tap === true)}
                                                  onToggle={toggleAll}/></th>
                <th>Namespace</th>
            </tr>
            </thead>
            <tbody>
            {Object.keys(namespaces).filter((namespace) => namespace.includes(searchValue)).map(namespace => {
                    return <tr>
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

    return <Modal
        open={isOpen}
        onClose={onClose}
        closeAfterTransition
        BackdropComponent={Backdrop}
        BackdropProps={{
            timeout: 500,
        }}
        style={{overflow: 'auto'}}
        disableBackdropClick={isFirstLogin}
    >
        <Fade in={isOpen}>
            <Box sx={modalStyle} style={{width: "40vw", maxWidth: 600, height: "50vh", padding: 0, display: "flex", justifyContent: "space-between", flexDirection: "column"}}>
                <div style={{padding: 32, paddingBottom: 0}}>
                    {isFirstLogin ? <div>
                        <div className="settingsTitle">Welcome to Mizu Ent.</div>
                        <div className="welcomeSubtitle" style={{marginTop: 15}}>The installation has finished
                            successfully,
                        </div>
                        <div className="welcomeSubtitle" style={{marginTop: 5}}>please review the Tapping Settings (can
                            be done at any time)
                            press ok to continue and view traffic
                        </div>
                    </div> : <div className="settingsTitle">Tapping Settings</div>}
                    {isLoading ? <div style={{textAlign: "center", padding: 20}}>
                        <img alt="spinner" src={spinner} style={{height: 35}}/>
                    </div> : <>
                        <div className="namespacesSettingsContainer">
                            <div style={{margin: "10px 0"}}>
                                <input className="searchNamespace" placeholder="Search" value={searchValue}
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
