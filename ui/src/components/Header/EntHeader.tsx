import React from "react";
import logo from '../assets/MizuEntLogo.svg';
import './Header.sass';
import userImg from '../assets/user-circle.svg';
import settingImg from '../assets/settings.svg';

export const EntHeader = () => {

    return <div className="header">
        <div>
            <div className="title">
                <img style={{height: 55}} src={logo} alt="logo"/>
            </div>
        </div>
        <div style={{display: "flex", alignItems: "center"}}>
            <img className="headerIcon" alt="settings" src={settingImg} style={{marginRight: 25}}/>
            <img className="headerIcon" alt="user" src={userImg}/>
        </div>
    </div>;
}
