import React, {useEffect, useState} from 'react';
import './App.sass';
import logo from './components/assets/Mizu-logo.svg';
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

    useEffect(() => {
        (async () => {
            const recentTLSLinks = await api.getRecentTLSLinks();

            if (recentTLSLinks?.length > 0) {
                setAddressesWithTLS(new Set([...addressesWithTLS, ...recentTLSLinks]));
                setShowTLSWarning(true);
            }

        })();
        // eslint-disable-next-line
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
                {analyzeStatus?.isAnalyzing &&
                <Tooltip title={analysisMessage} isSimple classes={classes}>
                    <div>

                        <Button
                            variant="contained"
                            color="primary"
                            disabled={!analyzeStatus?.isRemoteReady}
                            onClick={() => {
                                window.open(analyzeStatus?.remoteUrl)
                            }}>
                            Analysis
                        </Button>
                    </div>
                </Tooltip>
                }
            </div>
            <TrafficPage setAnalyzeStatus={setAnalyzeStatus} onTLSDetected={onTLSDetected}/>
            <Snackbar open={showTLSWarning && !userDismissedTLSWarning}>
                <MuiAlert elevation={6} variant="filled" onClose={() => setUserDismissedTLSWarning(true)} severity="warning">
                    Mizu is detecting TLS traffic{addressesWithTLS.size ? ` (directed to ${Array.from(addressesWithTLS).join(", ")})` : ''}, this type of traffic will not be displayed.
                </MuiAlert>
            </Snackbar>
        </div>
    );
}

export default App;
