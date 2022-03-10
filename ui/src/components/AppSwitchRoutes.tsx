import React, {useCallback, useEffect, useState} from "react";
import {Route, Routes, useNavigate} from "react-router-dom";
import {RouterRoutes} from "../helpers/routes";
import {useRecoilState} from "recoil";
import entPageAtom, {Page} from "../recoil/entPage";
import {toast} from "react-toastify";
import AuthPageBase from "./Pages/AuthPage/AuthPageBase";
import InstallPage from "./Pages/AuthPage/InstallPage";
import LoginPage from "./Pages/AuthPage/LoginPage";
import {UI} from "@up9/mizu-common"
import SystemViewer from "./Pages/SystemViewer/SystemViewer";
import Api from "../helpers/api";
import { TrafficPage } from "./Pages/TrafficPage/TrafficPage";


const api = Api.getInstance();

const AppSwitchRoutes = () => {

    const navigate = useNavigate();

    const [isLoading, setIsLoading] = useState(true);
    const [entPage, setEntPage] = useRecoilState(entPageAtom);
    const [isFirstLogin, setIsFirstLogin] = useState(false);

    const determinePage =  useCallback(async () => {
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

    useEffect(() => {
        switch (entPage) {
            case Page.Traffic:
                navigate("/");
                break;
            case Page.Setup:
                navigate(RouterRoutes.SETUP);
                break;
            case Page.Login:
                navigate(RouterRoutes.LOGIN);
                break;
            default:
                navigate(RouterRoutes.LOGIN);
        }
        // eslint-disable-next-line
    },[entPage])


    if (isLoading) {
        return <UI.LoadingOverlay/>;
    }

    return <Routes>
        <Route path={"/"} element={<SystemViewer isFirstLogin={isFirstLogin} setIsFirstLogin={setIsFirstLogin}/>}>
            <Route path={"/"} element={<TrafficPage/>} />
            <Route path={RouterRoutes.SETTINGS} element={<></>} /> {/*todo: set settings component*/}
        </Route>
        <Route path={RouterRoutes.LOGIN} element={<AuthPageBase><LoginPage/></AuthPageBase>}/>
        <Route path={RouterRoutes.SETUP} element={<AuthPageBase><InstallPage onFirstLogin={() => setIsFirstLogin(true)}/></AuthPageBase>}/>
    </Routes>
}

export default AppSwitchRoutes;
