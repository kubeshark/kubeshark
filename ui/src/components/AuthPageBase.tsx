import React from "react";
import background from "./assets/authBackground.png";
import logo from './assets/MizuEntLogoNoPowBy.svg';
import poweredBy from './assets/powered-by.svg'
import "./style/AuthBasePage.sass";


export const AuthPageBase: React.FC = ({children}) => {
    return <div className="authContainer" style={{background: `url(${background})`, backgroundSize: "cover"}}>
            <div className="authHeader">
                <img alt="logo" src={logo}/>
            </div>
            {children}
            <div className="authFooter">
                <img alt="logo" src={poweredBy}/>
            </div>
    </div>;
};

export default AuthPageBase;
