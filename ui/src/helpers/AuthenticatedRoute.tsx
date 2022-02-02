import React, { ReactNode, useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import LoadingOverlay from "../components/LoadingOverlay";
import Api from "./api";
import { RouterRoutes } from "./routes";

const api = Api.getInstance();

export const AuthenticatedRoute: React.FC  = ({children}) => {

    const navigate = useNavigate();
    const [isLoading, setIsLoading] = useState(false);

   useEffect(() => {
    (async () => {
        setIsLoading(true);
        try {
            const isInstallNeeded = await api.isInstallNeeded();
            if (isInstallNeeded) {
                navigate(RouterRoutes.SETUP);
            } else {
                const isAuthNeeded = await api.isAuthenticationNeeded();
                if(isAuthNeeded) {
                    navigate(RouterRoutes.LOGIN);
                }
            }
        } catch (e) {
            console.error(e);
        } finally {
                setIsLoading(false);
        }
    })();
   }, [])

    if (isLoading) {
        return <LoadingOverlay/>;
    }

    return <>{children}</>;
}
