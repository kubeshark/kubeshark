import React from "react";
import background from "./assets/authBackground.png";
import logo from './assets/MizuEntLogoFull.svg';
import "./style/AuthBasePage.sass";


export const AuthPageBase: React.FC = ({children}) => {
    return <div className="authContainer" style={{background: `url(${background})`, backgroundSize: "cover"}}>
            <div className="authHeader">
                <img src={logo}/>
            </div>
            {children}
    </div>;
};

export default AuthPageBase;
