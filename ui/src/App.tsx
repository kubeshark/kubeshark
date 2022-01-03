import React, {useEffect, useState} from 'react';
import './App.sass';
import Header from './components/Header';
import { TrafficPage } from './components/TrafficPage';
import Api from "./helpers/api";
import LoginPage from './components/LoginPage';
import InstallPage from './components/InstallPage';
import { ToastContainer, toast } from 'react-toastify';
import 'react-toastify/dist/ReactToastify.css';
import LoadingOverlay from './components/LoadOverlay';

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

    const [page, setPage] = useState(Page.Traffic); // TODO: move to state management
    const [analyzeStatus, setAnalyzeStatus] = useState(null); // TODO: move to state management

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

    let pageComponent: any;

    switch (page) {
        case Page.Traffic:
            pageComponent = <TrafficPage setAnalyzeStatus={setAnalyzeStatus}/>;
            break;
        case Page.Setup:
            pageComponent = <InstallPage/>;
            break;
        case Page.Login:
            pageComponent = <LoginPage/>;
            break;
        default:
            pageComponent = <div>Unknown page</div>;
    }

    return <div className="mizuApp">
        <MizuContext.Provider value={{page, setPage}}>
            <Header analyzeStatus={analyzeStatus}/>
            {isLoading ? <LoadingOverlay/> : pageComponent}
        </MizuContext.Provider>
        <ToastContainer
                    position="bottom-right"
                    autoClose={5000}
                    hideProgressBar={false}
                    newestOnTop={false}
                    closeOnClick
                    rtl={false}
                    pauseOnFocusLoss
                    draggable
                    pauseOnHover
        />
    </div>
}

export default App;
