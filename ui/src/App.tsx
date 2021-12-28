import React, {useState} from 'react';
import './App.sass';
import {TrafficPage} from "./components/TrafficPage";
import {TLSWarning} from "./components/TLSWarning/TLSWarning";
import {Header} from "./components/Header/Header";

const App = () => {

    const [analyzeStatus, setAnalyzeStatus] = useState(null);
    const [showTLSWarning, setShowTLSWarning] = useState(false);
    const [userDismissedTLSWarning, setUserDismissedTLSWarning] = useState(false);
    const [addressesWithTLS, setAddressesWithTLS] = useState(new Set<string>());

    const onTLSDetected = (destAddress: string) => {
        addressesWithTLS.add(destAddress);
        setAddressesWithTLS(new Set(addressesWithTLS));

        if (!userDismissedTLSWarning) {
            setShowTLSWarning(true);
        }
    };

    return (
        <div className="mizuApp">
            <Header analyzeStatus={analyzeStatus}/>
            <TrafficPage setAnalyzeStatus={setAnalyzeStatus} onTLSDetected={onTLSDetected}/>
            <TLSWarning showTLSWarning={showTLSWarning}
                        setShowTLSWarning={setShowTLSWarning}
                        addressesWithTLS={addressesWithTLS}
                        setAddressesWithTLS={setAddressesWithTLS}
                        userDismissedTLSWarning={userDismissedTLSWarning}
                        setUserDismissedTLSWarning={setUserDismissedTLSWarning}/>
        </div>
    );
}

export default App;
