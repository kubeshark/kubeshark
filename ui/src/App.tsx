import React, {useState} from 'react';
import './App.sass';
import {TLSWarning} from "./components/TLSWarning/TLSWarning";
import {Header} from "./components/Header/Header";
import {TrafficPage} from "./components/TrafficPage";
import { ServiceMapModal } from './components/ServiceMapModal/ServiceMapModal';

const App = () => {

    const [analyzeStatus, setAnalyzeStatus] = useState(null);
    const [showTLSWarning, setShowTLSWarning] = useState(false);
    const [userDismissedTLSWarning, setUserDismissedTLSWarning] = useState(false);
    const [addressesWithTLS, setAddressesWithTLS] = useState(new Set<string>());
    const [openServiceMapModal, setOpenServiceMapModal] = useState(false);

    const onTLSDetected = (destAddress: string) => {
        addressesWithTLS.add(destAddress);
        setAddressesWithTLS(new Set(addressesWithTLS));

        if (!userDismissedTLSWarning) {
            setShowTLSWarning(true);
        }
    };

    return (
        <div className="mizuApp">
            <Header analyzeStatus={analyzeStatus} />
            <TrafficPage setAnalyzeStatus={setAnalyzeStatus} onTLSDetected={onTLSDetected} setOpenServiceMapModal={setOpenServiceMapModal} />
            <TLSWarning showTLSWarning={showTLSWarning}
                setShowTLSWarning={setShowTLSWarning}
                addressesWithTLS={addressesWithTLS}
                setAddressesWithTLS={setAddressesWithTLS}
                userDismissedTLSWarning={userDismissedTLSWarning}
                setUserDismissedTLSWarning={setUserDismissedTLSWarning} />
            {window["isServiceMapEnabled"] && <ServiceMapModal
                isOpen={openServiceMapModal}
                onOpen={() => setOpenServiceMapModal(true)}
                onClose={() => setOpenServiceMapModal(false)}
            />}
        </div>
    );
}

export default App;
