import React, {useCallback, useEffect, useState} from 'react';
import './App.sass';
import {TrafficPage} from "./components/TrafficPage";
import {TLSWarning} from "./components/TLSWarning/TLSWarning";
import {EntHeader} from "./components/Header/EntHeader";
import Api from "./helpers/api";
import {toast} from "react-toastify";
import InstallPage from "./components/InstallPage";
import LoginPage from "./components/LoginPage";
import LoadingOverlay from "./components/LoadingOverlay";
import AuthPageBase from './components/AuthPageBase';
import entPageAtom, {Page} from "./recoil/entPage";
import {useRecoilState} from "recoil";

const api = Api.getInstance();

const EntApp = () => {

    const [isLoading, setIsLoading] = useState(true);
    const [showTLSWarning, setShowTLSWarning] = useState(false);
    const [userDismissedTLSWarning, setUserDismissedTLSWarning] = useState(false);
    const [addressesWithTLS, setAddressesWithTLS] = useState(new Set<string>());
    const [entPage, setEntPage] = useRecoilState(entPageAtom);
    const [isFirstLogin, setIsFirstLogin] = useState(false);

    const determinePage =  useCallback(async () => { // TODO: move to state management
        try {
            const isInstallNeeded = await api.isInstallNeeded();
            if (isInstallNeeded) {
                setEntPage(Page.Setup);
            } else {
                const isAuthNeeded = await api.isAuthenticationNeeded();
                if(isAuthNeeded) {
                    setEntPage(Page.Login);
                }
            }
        } catch (e) {
            toast.error("Error occured while checking Mizu API status, see console for mode details");
            console.error(e);
        } finally {
            setIsLoading(false);
        }
    },[setEntPage]);

    useEffect(() => {
        determinePage();
    }, [determinePage]);

    const onTLSDetected = (destAddress: string) => {
        addressesWithTLS.add(destAddress);
        setAddressesWithTLS(new Set(addressesWithTLS));

        if (!userDismissedTLSWarning) {
            setShowTLSWarning(true);
        }
    };

    let pageComponent: any;

    switch (entPage) { // TODO: move to state management / proper routing
        case Page.Traffic:
            pageComponent = <TrafficPage onTLSDetected={onTLSDetected}/>;
            break;
        case Page.Setup:
            pageComponent = <AuthPageBase><InstallPage onFirstLogin={() => setIsFirstLogin(true)}/></AuthPageBase>;
            break;
        case Page.Login:
            pageComponent = <AuthPageBase><LoginPage/></AuthPageBase>;
            break;
        default:
            pageComponent = <div>Unknown Error</div>;
    }

    if (isLoading) {
        return <LoadingOverlay/>;
    }

    return (
        <div className="mizuApp">
            {entPage === Page.Traffic && <EntHeader isFirstLogin={isFirstLogin} setIsFirstLogin={setIsFirstLogin}/>}
            {pageComponent}
            {entPage === Page.Traffic && <TLSWarning showTLSWarning={showTLSWarning}
                        setShowTLSWarning={setShowTLSWarning}
                        addressesWithTLS={addressesWithTLS}
                        setAddressesWithTLS={setAddressesWithTLS}
                        userDismissedTLSWarning={userDismissedTLSWarning}
                        setUserDismissedTLSWarning={setUserDismissedTLSWarning}/>}
        </div>
    );
}

export default EntApp;
