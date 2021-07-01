import React, {useState} from 'react';
import './App.sass';
import logo from './components/assets/Mizu.svg';
import {Button, ThemeProvider} from "@material-ui/core";
import Theme from './style/mui.theme';
import {HarPage} from "./components/HarPage";


const App = () => {

    const [analyzeStatus, setAnalyzeStatus] = useState(null);

    return (
        <ThemeProvider theme={Theme}>
            <div className="mizuApp">
                <div className="header">
                    <div style={{display: "flex", alignItems: "center"}}>
                        <div className="title"><img src={logo} alt="logo"/></div>
                        <div className="description">Traffic viewer for Kubernetes</div>
                    </div>
                    <div>
                        {analyzeStatus?.isAnalyzing &&
                        <div title={!analyzeStatus?.isRemoteReady ? "Analysis is not ready yet" : "Go To see further analysis"}>
                            <Button
                                variant="contained"
                                color="primary"
                                disabled={!analyzeStatus?.isRemoteReady}
                                onClick={() => {
                                    window.open(analyzeStatus?.remoteUrl)
                                }}>
                                Analysis Results
                            </Button>
                        </div>
                        }
                    </div>
                </div>
                <HarPage setAnalyzeStatus={setAnalyzeStatus}/>
            </div>
        </ThemeProvider>
    );
}

export default App;
