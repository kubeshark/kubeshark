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

    return <Routes>
        <Route path={"/"} element={<SystemViewer/>}>
            <Route path={RouterRoutes.SETTINGS} element={<AuthenticatedRoute><SettingsPage/></AuthenticatedRoute>} />
            <Route path={"/"} element={<AuthenticatedRoute><TrafficPage/></AuthenticatedRoute>}/>
            
        </Route>
        <Route path={RouterRoutes.SETUP} element={<AuthPageBase><InstallPage/></AuthPageBase>}/>
        <Route path={RouterRoutes.SETUP + "/:inviteToken"} element={<AuthPageBase><InstallPage/></AuthPageBase>}/>
        <Route path={RouterRoutes.LOGIN} element={<AuthPageBase><LoginPage/></AuthPageBase>}/>
        
    </Routes>
}

export default AppSwitchRoutes;
