import React, {useEffect, useState} from 'react';
import './App.sass';
import Api from "./helpers/api";


const api = new Api();

enum Page {
    Traffic,
    Setup,
    Login
}


const App = () => {
    const [isLoading, setIsLoading] = useState(true);
    const [page, setPage] = useState(Page.Traffic);

    const determinePage = async () => {
        try {
            const isInstallNeeded = await api.isInstallNeeded();
            if (isInstallNeeded) {
                setPage(Page.Setup);
            } else {
                if(await api.isAuthenticationNeeded()) {
                    setPage(Page.Login);
                }
            }
        } catch (e) {
            console.error(e);
        } finally {
            setIsLoading(false);
        }
    }

    useEffect(() => {
        determinePage();
    }, []);
}

export default App;
