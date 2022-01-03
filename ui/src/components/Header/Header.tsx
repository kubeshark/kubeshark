import React, { useContext } from "react";
import {AuthPresentation} from "../AuthPresentation/AuthPresentation";
import {AnalyzeButton} from "../AnalyzeButton/AnalyzeButton";
import logo from '../assets/Mizu-logo.svg';
import { Button } from "@material-ui/core";
import { toast } from "react-toastify";
import Api from "../../helpers/api";
import { MizuContext, Page } from "../../App";

interface HeaderProps {
    analyzeStatus: any
}
export const Header: React.FC<HeaderProps> = ({analyzeStatus}) => {

    const {setPage} = useContext(MizuContext);

    return <div className="header">
        <div style={{display: "flex", alignItems: "center"}}>
            <div className="title"><img src={logo} alt="logo"/></div>
            <div className="description">Traffic viewer for Kubernetes</div>
        </div>
        <div style={{display: "flex", alignItems: "center"}}>
            {analyzeStatus?.isAnalyzing && <AnalyzeButton analyzeStatus={analyzeStatus}/>}
            <AuthPresentation/>
        </div>
    </div>;
}
