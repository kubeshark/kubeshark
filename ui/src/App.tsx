import React, {useState} from 'react';
import './App.sass';
import logo from './components/assets/Mizu-logo.svg';
import {Button} from "@material-ui/core";
import {HarPage} from "./components/HarPage";


const App = () => {

    const [analyzeStatus, setAnalyzeStatus] = useState(null);

    return (
        <div className="mizuApp">
            <div className="header">
                <div style={{display: "flex", alignItems: "center"}}>
                    <div className="title"><img src={logo} alt="logo"/></div>
                    <div className="description">Traffic viewer for Kubernetes</div>
                </div>
                <div>
                    {analyzeStatus?.isAnalyzing &&
                    <div
                        title={!analyzeStatus?.isRemoteReady ? "Analysis is not ready yet" : "Go To see further analysis"}>
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
                    }
                </div>
            </div>
            <HarPage setAnalyzeStatus={setAnalyzeStatus}/>
        </div>
    );
}

export default App;
