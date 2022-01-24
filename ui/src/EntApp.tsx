import React from 'react';
import './App.sass';
import AppSwitchRoutes from "./components/AppSwitchRoutes";
import {BrowserRouter} from "react-router-dom";

const EntApp = () => {

    return (
        <div className="mizuApp">
            <BrowserRouter>
                <AppSwitchRoutes/>
            </BrowserRouter>
        </div>
    );
}

export default EntApp;
