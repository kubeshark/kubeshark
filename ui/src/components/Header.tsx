import { Button, makeStyles } from "@material-ui/core";
import React, { useEffect, useState } from "react";
import Tooltip from "./UI/Tooltip";
import logo from './assets/Mizu-logo.svg';
import logo_up9 from './assets/logo_up9.svg';
import Api from "../helpers/api";

const api = new Api();

const useStyles = makeStyles(() => ({
    tooltip: {
        backgroundColor: "#3868dc",
        color: "white",
        fontSize: 13,
    },
}));


interface HeaderProps {
    analyzeStatus: any;  // TODO: move to state management
}

const Header: React.FC<HeaderProps> = ({analyzeStatus}) => {

    const [statusAuth, setStatusAuth] = useState(null);
    const classes = useStyles();

    useEffect(() => {
        (async () => {
            try {
                const auth = await api.getAuthStatus();
                setStatusAuth(auth);
            } catch (e) {
                console.error(e);
            }

        })();
    }, []);


    const analysisMessage = analyzeStatus?.isRemoteReady ?
        <span>
            <table>
                <tr>
                    <td>Status</td>
                    <td><b>Available</b></td>
                </tr>
                <tr>
                    <td>Messages</td>
                    <td><b>{analyzeStatus?.sentCount}</b></td>
                </tr>
            </table>
        </span> :
        analyzeStatus?.sentCount > 0 ?
            <span>
                <table>
                    <tr>
                        <td>Status</td>
                        <td><b>Processing</b></td>
                    </tr>
                    <tr>
                        <td>Messages</td>
                        <td><b>{analyzeStatus?.sentCount}</b></td>
                    </tr>
                    <tr>
                        <td colSpan={2}> Please allow a few minutes for the analysis to complete</td>
                    </tr>
                </table>
            </span> :
            <span>
                <table>
                    <tr>
                        <td>Status</td>
                        <td><b>Waiting for traffic</b></td>
                    </tr>
                    <tr>
                        <td>Messages</td>
                        <td><b>{analyzeStatus?.sentCount}</b></td>
                    </tr>
                </table>

            </span>

    return <div className="header">
        <div style={{display: "flex", alignItems: "center"}}>
            <div className="title"><img src={logo} alt="logo"/></div>
            <div className="description">Traffic viewer for Kubernetes</div>
        </div>
        <div style={{display: "flex", alignItems: "center"}}>

            {analyzeStatus?.isAnalyzing &&
                <div>
                    <Tooltip title={analysisMessage} isSimple classes={classes}>
                        <div>
                            <Button
                                style={{fontFamily: "system-ui",
                                    fontWeight: 600,
                                    fontSize: 12,
                                    padding: 8}}
                                size={"small"}
                                variant="contained"
                                color="primary"
                                startIcon={<img style={{height: 24, maxHeight: "none", maxWidth: "none"}} src={logo_up9} alt={"up9"}/>}
                                disabled={!analyzeStatus?.isRemoteReady}
                                onClick={() => {
                                    window.open(analyzeStatus?.remoteUrl)
                                }}>
                                Analysis
                            </Button>
                        </div>
                    </Tooltip>
                </div>
            }
            {statusAuth?.email && <div style={{display: "flex", borderLeft: "2px #87878759 solid", paddingLeft: 10, marginLeft: 10}}>
                <div style={{color: "rgba(0,0,0,0.75)"}}>
                    <div style={{fontWeight: 600, fontSize: 13}}>{statusAuth.email}</div>
                    <div style={{fontSize:11}}>{statusAuth.model}</div>
                </div>
            </div>}
        </div>
    </div>
};

export default Header;