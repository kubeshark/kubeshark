import React, {useState} from 'react';
import './App.sass';
import {TrafficPage} from "./components/TrafficPage";
import {TLSWarning} from "./components/TLSWarning/TLSWarning";
import {EntHeader} from "./components/Header/EntHeader";

const EntApp = () => {

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
            <EntHeader/>
            <TrafficPage onTLSDetected={onTLSDetected}/>
            <TLSWarning showTLSWarning={showTLSWarning}
                        setShowTLSWarning={setShowTLSWarning}
                        addressesWithTLS={addressesWithTLS}
                        setAddressesWithTLS={setAddressesWithTLS}
                        userDismissedTLSWarning={userDismissedTLSWarning}
                        setUserDismissedTLSWarning={setUserDismissedTLSWarning}/>
        </div>
    );
}

export default EntApp;
