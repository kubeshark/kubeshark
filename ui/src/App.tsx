import React, {useState} from 'react';
import './App.sass';
import logo from './components/assets/Mizu-logo.svg';
import {Button} from "@material-ui/core";
import {HarPage} from "./components/HarPage";
import Tooltip from "./components/Tooltip";
import {makeStyles} from "@material-ui/core/styles";


const useStyles = makeStyles(() => ({
    tooltip: {
        backgroundColor: "#3868dc",
        color: "white",
        fontSize: 13,
    },
}));

const App = () => {

    const classes = useStyles();

    const [analyzeStatus, setAnalyzeStatus] = useState(null);

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
            <HarPage setAnalyzeStatus={setAnalyzeStatus}/>
        </div>
    );
}

export default App;
