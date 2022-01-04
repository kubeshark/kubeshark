import React, {useState} from "react";
import {Modal, Backdrop, Fade, Box, Button, makeStyles} from "@material-ui/core";
import {modalStyle} from "../Filters";
import Checkbox from "../UI/Checkbox";
import variables from "../../variables.module.scss";

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

export const SettingsModal: React.FC<SettingsModalProps> = ({isOpen, onClose}) => {

    const classes = useStyles();
    const [namespaces, setNamespaces] = useState([{name: "aaa", tap: true}, {name: "bbb", tap: true}, {name: "ccc", tap: true}]);
    const [searchValue, setSearchValue] = useState("");
    const isFirstLogin = false;

    const selectAll = () => {
        setNamespaces(namespaces.map(namespace => {
            return {
                name: namespace.name,
                tap: true
            }
        }))
    };

    const clearAll = () => {
        setNamespaces(namespaces.map(namespace => {
            return {
                name: namespace.name,
                tap: false
            }
        }))
    }

    const updateTappingSettings = () => {

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
                        <div style={{textAlign: "center", fontSize: 32, color: "#205cf5", fontWeight: 600}}>Welcome to Mizu Ent.</div>
                        <div style={{fontSize: 16, color: "rgba(0,0,0,0.5)", fontWeight: 600, marginTop: 15}}>The installation has finished successfully,</div>
                        <div style={{fontSize: 16, color: "rgba(0,0,0,0.5)", fontWeight: 600, marginTop: 5}}>please review the Tapping Settings (can be done at any time)
                            press ok to continue and view traffic</div>
                    </div> : <div style={{textAlign: "center", fontSize: 32, color: "#205cf5", fontWeight: 600}}>Tapping settings</div>}
                    <div style={{marginTop: 20, backgroundColor: "#E9EBF8", borderRadius: 4, padding: 20}}>
                        <div>
                            <Button className={classes.button} size={"small"} onClick={selectAll}>select all</Button>
                            <Button style={{marginLeft: 10}} className={classes.button} size={"small"} onClick={clearAll}>clear all</Button>
                        </div>
                        <div style={{margin: "15px 0"}}><input placeholder="search" style={{outline: 0, background: "white", borderRadius: 20, width: 400, padding: "5px 10px", border: "none"}} value={searchValue} onChange={(event) => setSearchValue(event.target.value)}/></div>
                        <div>
                            <table cellPadding={5}>
                                <thead>
                                    <tr>
                                        <th style={{width: 50, textAlign: "left"}}>Tap</th>
                                        <th>Namespace</th>
                                    </tr>
                                </thead>
                                <tbody>
                                {namespaces.filter(namespace => namespace.name.includes(searchValue)).map(namespace => {
                                        return <tr>
                                            <td style={{width: 50}}><Checkbox checked={namespace.tap} onToggle={() => setNamespaces(namespaces.map(n => n.name === namespace.name ? {name: namespace.name, tap: !namespace.tap} : n))}/></td>
                                            <td>{namespace.name}</td>
                                        </tr>
                                    }
                                )}
                                </tbody>
                            </table>
                        </div>
                    </div>
                    <div style={{display: "flex", justifyContent: "flex-end", marginTop: 20}}>
                        {!isFirstLogin && <Button style={{fontSize: 14, padding: "6px 12px"}} className={classes.button} size={"small"} onClick={onClose}>Cancel</Button>}
                        <Button style={{marginLeft: 10, fontSize: 14, padding: "6px 12px"}} className={classes.button} size={"small"} onClick={updateTappingSettings}>OK</Button>
                    </div>
                </Box>
            </Fade>
        </Modal>
}
