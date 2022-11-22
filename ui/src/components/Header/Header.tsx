import React from "react";
import {AuthPresentation} from "../AuthPresentation/AuthPresentation";
import logo from '../assets/Kubeshark-logo.svg';
import './Header.sass';
import {UI} from "@up9/kubeshark-common"


export const Header: React.FC = () => {
    return <div className="header">
        <div style={{display: "flex", alignItems: "center"}}>
          <img className="logo" src={logo} alt="logo"/>
            <div className="title">Kubeshark</div>
            <div className="description">Traffic viewer for Kubernetes</div>
        </div>
        <div style={{display: "flex", alignItems: "center"}}>
            <UI.InformationIcon/>
            <AuthPresentation/>
        </div>
    </div>;
}
