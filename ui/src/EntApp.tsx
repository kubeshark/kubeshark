import React, { useEffect } from 'react';
import './App.sass';
import AppSwitchRoutes from "./components/AppSwitchRoutes";
import {BrowserRouter} from "react-router-dom";
import { useSetRecoilState } from 'recoil';
import loggedInUserStateAtom from './recoil/loggedInUserState/atom';
import Api from './helpers/api';

const api = Api.getInstance();

const EntApp = () => {
    const setUserDetails = useSetRecoilState(loggedInUserStateAtom);

    useEffect(()=>{
        (async () => {
            const userDetails = await api.whoAmI();
            setUserDetails(userDetails);
        })()
        return setUserDetails({                
            "username": "",
            "role": "",
            "workspace": {
                "id": "",
                "name": "",
                "namespaces": []
        }}); 
    },[setUserDetails])

    return (
        <div className="mizuApp">
            <BrowserRouter>
                <AppSwitchRoutes/>
            </BrowserRouter>
        </div>
    );
}

export default EntApp;
