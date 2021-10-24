import React, {useEffect, useState} from 'react';
import './App.sass';
import logo from './components/assets/Mizu-logo.svg';
import logo_up9 from './components/assets/logo_up9.svg';
import {Button, Snackbar} from "@material-ui/core";
import {TrafficPage} from "./components/TrafficPage";
import Tooltip from "./components/UI/Tooltip";
import {makeStyles} from "@material-ui/core/styles";
import MuiAlert from '@material-ui/lab/Alert';
import Api from "./helpers/api";


const useStyles = makeStyles(() => ({
    tooltip: {
        backgroundColor: "#3868dc",
        color: "white",
        fontSize: 13,
    },
}));

const api = new Api();

const App = () => {

    const classes = useStyles();

    const [analyzeStatus, setAnalyzeStatus] = useState(null);
    const [showTLSWarning, setShowTLSWarning] = useState(false);
    const [userDismissedTLSWarning, setUserDismissedTLSWarning] = useState(false);
    const [addressesWithTLS, setAddressesWithTLS] = useState(new Set());
    const [statusAuth, setStatusAuth] = useState(null);

    useEffect(() => {
        (async () => {
            try {
                const recentTLSLinks = await api.getRecentTLSLinks();
                if (recentTLSLinks?.length > 0) {
                    setAddressesWithTLS(new Set(recentTLSLinks));
                    setShowTLSWarning(true);
                }
                const auth = await api.getAuthStatus();
                setStatusAuth(auth);
            } catch (e) {
                console.error(e);
            }

        })();
    }, []);

    const onTLSDetected = (destAddress: string) => {
        addressesWithTLS.add(destAddress);
        setAddressesWithTLS(new Set(addressesWithTLS));

        if (!userDismissedTLSWarning) {
            setShowTLSWarning(true);
        }
    };

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

    return (
        <div className="mizuApp">
            <div className="header">
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
            <TrafficPage setAnalyzeStatus={setAnalyzeStatus} onTLSDetected={onTLSDetected}/>
            <Snackbar open={showTLSWarning && !userDismissedTLSWarning}>
                <MuiAlert classes={{ filledWarning: 'customWarningStyle' }} elevation={6} variant="filled" onClose={() => setUserDismissedTLSWarning(true)} severity="warning">
                    Mizu is detecting TLS traffic, this type of traffic will not be displayed.
                    {addressesWithTLS.size > 0 && <ul className="httpsDomains"> {Array.from(addressesWithTLS, address => <li>{address}</li>)} </ul>}
                </MuiAlert>
            </Snackbar>
        </div>
    );
}

export default App;
