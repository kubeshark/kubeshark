import React, {useState} from 'react';
import {HarPage} from "./components/HarPage";
import './App.sass';
import logo from './components/assets/Mizu.svg';
import {Button} from "@material-ui/core";
import Theme from './style/mui.theme';
import { ThemeProvider } from '@material-ui/core';


const App = () => {

    const [analyzeStatus, setAnalyzeStatus] = useState(null);

    return (
        <ThemeProvider theme={Theme}>
            <div className="mizuApp">
                <div className="header">
                    <div>
                        <div className="title"><img src={logo} alt="logo"/></div>
                        <div className="description">Traffic viewer for Kubernetes</div>
                    </div>
                    <div>
                        {analyzeStatus?.isAnalyzing && null}
                        <Button variant="contained" color="primary" disabled={!analyzeStatus?.isRemoteReady} onClick={() => {
                            window.open(analyzeStatus?.remoteUrl)
                        }}>
                            Go to UP9
                        </Button>

                    </div>
                </div>
                <HarPage setAnalyzeStatus={setAnalyzeStatus}/>
            </div>
        </ThemeProvider>
    );
}

export default App;
