import React, {useEffect, useState} from 'react';
import './App.sass';
import {TrafficPage} from "./components/TrafficPage";
import {TLSWarning} from "./components/TLSWarning/TLSWarning";
import {Header} from "./components/Header/Header";
import { ToastContainer, toast } from 'react-toastify';
import 'react-toastify/dist/ReactToastify.css';
import Api from "./helpers/api";
import LoadingOverlay from './components/LoadOverlay';
import LoginPage from './components/LoginPage';
import InstallPage from './components/InstallPage';

const api = new Api();

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

const App = () => {
    const [isLoading, setIsLoading] = useState(true);
    const [analyzeStatus, setAnalyzeStatus] = useState(null);
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
                const isAuthed = await api.isAuthenticationNeeded();
                if(isAuthed) {
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

    switch (page) {
        case Page.Traffic:
            pageComponent = <TrafficPage setAnalyzeStatus={setAnalyzeStatus} onTLSDetected={onTLSDetected}/>;
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
                <Header analyzeStatus={analyzeStatus}/>
                <TrafficPage setAnalyzeStatus={setAnalyzeStatus} onTLSDetected={onTLSDetected}/>
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

export default App;
