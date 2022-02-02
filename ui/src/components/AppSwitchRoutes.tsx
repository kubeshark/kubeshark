import React, { useState} from "react";
import {Route, Routes} from "react-router-dom";
import {RouterRoutes} from "../helpers/routes";
import AuthPageBase from "./Pages/AuthPage/AuthPageBase";
import InstallPage from "./Pages/AuthPage/InstallPage";
import LoginPage from "./Pages/AuthPage/LoginPage";
import SystemViewer from "./Pages/SystemViewer/SystemViewer";

import {TrafficPage} from "./Pages/TrafficPage/TrafficPage";
import SettingsPage from "./Pages/SettingsPage/SettingsPage";
import { AuthenticatedRoute } from "../helpers/AuthenticatedRoute";


const AppSwitchRoutes = () => {

    const [isFirstLogin, setIsFirstLogin] = useState(false);
    



    return <Routes>
        <Route path={"/"} element={<SystemViewer isFirstLogin={isFirstLogin} setIsFirstLogin={setIsFirstLogin}/>}>
            <Route path={RouterRoutes.SETTINGS} element={<AuthenticatedRoute><SettingsPage/></AuthenticatedRoute>} /> {/*todo: set settings component*/}
            <Route path={"/"} element={<AuthenticatedRoute><TrafficPage/></AuthenticatedRoute>}/>
            
        </Route>
        <Route path={RouterRoutes.SETUP+ "/:inviteToken"} element={<AuthPageBase><InstallPage onFirstLogin={() => setIsFirstLogin(true)}/></AuthPageBase>}/>
        <Route path={RouterRoutes.LOGIN} element={<AuthPageBase><LoginPage/></AuthPageBase>}/>
        
    </Routes>
}

export default AppSwitchRoutes;
