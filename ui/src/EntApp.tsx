import React, {useEffect, useState} from 'react';
import './App.sass';
import {TrafficPage} from "./components/TrafficPage";
import {TLSWarning} from "./components/TLSWarning/TLSWarning";
import {EntHeader} from "./components/Header/EntHeader";
import Api from "./helpers/api";
import {toast} from "react-toastify";
import InstallPage from "./components/InstallPage";
import LoginPage from "./components/LoginPage";
import LoadingOverlay from "./components/LoadingOverlay";

const api = Api.getInstance();

// TODO: move to state management
export enum Page {
    Traffic,
    Setup,
    Login
}

// TODO: move to state management
export interface MizuContextModel {
    page: Page;
    setPage: (page: Page) => void;
}

// TODO: move to state management
export const MizuContext = React.createContext<MizuContextModel>(null);

const EntApp = () => {

    const [isLoading, setIsLoading] = useState(true);
    const [showTLSWarning, setShowTLSWarning] = useState(false);
    const [userDismissedTLSWarning, setUserDismissedTLSWarning] = useState(false);
    const [addressesWithTLS, setAddressesWithTLS] = useState(new Set<string>());
    const [page, setPage] = useState(Page.Traffic); // TODO: move to state management

    const determinePage = async () => { // TODO: move to state management
        try {
            const isInstallNeeded = await api.isInstallNeeded();
            if (isInstallNeeded) {
                setPage(Page.Setup);
            } else {
                const isAuthNeeded = await api.isAuthenticationNeeded();
                if(isAuthNeeded) {
                    setPage(Page.Login);
                }
            }
        } catch (e) {
            toast.error("Error occured while checking Mizu API status, see console for mode details");
            console.error(e);
        } finally {
            setIsLoading(false);
        }
    }

    useEffect(() => {
        determinePage();
    }, []);

    const onTLSDetected = (destAddress: string) => {
        addressesWithTLS.add(destAddress);
        setAddressesWithTLS(new Set(addressesWithTLS));

        if (!userDismissedTLSWarning) {
            setShowTLSWarning(true);
        }
    };

    let pageComponent: any;

    switch (page) { // TODO: move to state management / proper routing
        case Page.Traffic:
            pageComponent = <TrafficPage onTLSDetected={onTLSDetected}/>;
            break;
        case Page.Setup:
            pageComponent = <InstallPage/>;
            break;
        case Page.Login:
            pageComponent = <LoginPage/>;
            break;
        default:
            pageComponent = <div>Unknown Error</div>;
    }

    if (isLoading) {
        return <LoadingOverlay/>;
    }

    return (
        <div className="mizuApp">
            <MizuContext.Provider value={{page, setPage}}>
                <EntHeader/>
                {pageComponent}
                <TLSWarning showTLSWarning={showTLSWarning}
                            setShowTLSWarning={setShowTLSWarning}
                            addressesWithTLS={addressesWithTLS}
                            setAddressesWithTLS={setAddressesWithTLS}
                            userDismissedTLSWarning={userDismissedTLSWarning}
                            setUserDismissedTLSWarning={setUserDismissedTLSWarning}/>
            </MizuContext.Provider>
        </div>
    );
}

export default EntApp;
