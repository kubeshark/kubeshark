import React from "react";
import {AuthPresentation} from "../AuthPresentation/AuthPresentation";
import logo from '../assets/Mizu-logo.svg';
import './Header.sass';
import {UI} from "@up9/mizu-common"


export const Header: React.FC = () => {
    return <div className="header">
        <div style={{display: "flex", alignItems: "center"}}>
            <div className="title"><img src={logo} alt="logo"/></div>
            <div className="description">Traffic viewer for Kubernetes</div>
        </div>
        <div style={{display: "flex", alignItems: "center"}}>
            <UI.InformationIcon/>
            <AuthPresentation/>
        </div>
    </div>;
}
