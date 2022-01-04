import React, {useEffect, useState} from "react";
import {Modal, Backdrop, Fade, Box, Button, makeStyles} from "@material-ui/core";
import {modalStyle} from "../Filters";
import Checkbox from "../UI/Checkbox";
import './SettingsModal.sass';
import Api from "../../helpers/api";
import spinner from "../assets/spinner.svg";

interface SettingsModalProps {
    isOpen: boolean
    onClose: () => void
}

// @ts-ignore
const useStyles = makeStyles(() => ({
    button: {
        backgroundColor: "#205cf5",
        color: "white",
        fontWeight: "600 !important",
        fontSize: 12,
        padding: "4px 10px",

        "&:hover": {
            backgroundColor: "#205cf5",
        },
    }
}));

const api = new Api();

export const SettingsModal: React.FC<SettingsModalProps> = ({isOpen, onClose}) => {

    const classes = useStyles();
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

    const selectAll = () => {
        setAllNamespacesTappedValue(true);
    };

    const clearAll = () => {
        setAllNamespacesTappedValue(false);
    }

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

    const buildNamespacesTable = () => {
        return <table cellPadding={5}>
            <thead>
            <tr>
                <th style={{width: 50, textAlign: "left"}}>Tap</th>
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
                <Box sx={modalStyle} style={{width: "50vw"}}>
                    {isFirstLogin ? <div>
                        <div className="settingsTitle">Welcome to Mizu Ent.</div>
                        <div className="welcomeSubtitle" style={{marginTop: 15}}>The installation has finished successfully,</div>
                        <div className="welcomeSubtitle" style={{marginTop: 5}}>please review the Tapping Settings (can be done at any time)
                            press ok to continue and view traffic</div>
                    </div> : <div className="settingsTitle">Tapping settings</div>}
                    {isLoading ? <div style={{textAlign: "center", padding: 20}}>
                        <img alt="spinner" src={spinner} style={{height: 35}}/>
                    </div> : <>
                        <div className="namespacesSettingsContainer">
                            <div>
                                <Button className={classes.button} size={"small"} onClick={selectAll}>select all</Button>
                                <Button style={{marginLeft: 10}} className={classes.button} size={"small"} onClick={clearAll}>clear all</Button>
                            </div>
                            <div style={{margin: "15px 0"}}>
                                <input className="searchNamespace" placeholder="search" value={searchValue} onChange={(event) => setSearchValue(event.target.value)}/></div>
                            <div>
                                {buildNamespacesTable()}
                            </div>
                        </div>
                        <div className="settingsActionsContainer">
                            {!isFirstLogin && <Button style={{fontSize: 14, padding: "6px 12px"}} className={classes.button} size={"small"} onClick={onClose}>Cancel</Button>}
                            <Button style={{marginLeft: 10, fontSize: 14, padding: "6px 12px"}} className={classes.button} size={"small"} onClick={updateTappingSettings}>OK</Button>
                        </div>
                    </>}
                </Box>
            </Fade>
        </Modal>
}
