import React, {useState} from 'react';
import './App.sass';
import logo from './components/assets/Mizu.svg';
import {Button, makeStyles} from "@material-ui/core";
import {HarPage} from "./components/HarPage";
import Tooltip from "./components/Tooltip";


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
            Analysis is available <br />
            Uploaded {analyzeStatus?.sentCount} messages
        </span> :
        analyzeStatus?.sentCount > 0 ?
            <span>
                    Uploaded {analyzeStatus?.sentCount} message <br />
                    It is normally take few minutes to get first analysis results
            </span> :
            <span>
                    0 messages sent <br />
                    Analysis will start once messages will be uploaded
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
