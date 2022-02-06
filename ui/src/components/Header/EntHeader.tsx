import React, { useEffect, useState } from "react";
import logo from '../assets/MizuEntLogo.svg';
import './Header.sass';
import userImg from '../assets/user-circle.svg';
import settingImg from '../assets/settings.svg';
import {Menu, MenuItem} from "@material-ui/core";
import PopupState, {bindMenu, bindTrigger} from "material-ui-popup-state";
import logoutIcon from '../assets/logout.png';
import Api from "../../helpers/api";
import {toast} from "react-toastify";
import { useRecoilValue, useSetRecoilState } from "recoil";
import {useNavigate} from "react-router-dom";
import {RouterRoutes} from "../../helpers/routes";
import { SettingsModal } from "../SettingsModal/SettingModal";
import loggedInUserStateAtom from "../../recoil/loggedInUserState/atom";
import { Roles } from "../../recoil/loggedInUserState";

const api = Api.getInstance();


export const EntHeader: React.FC = () => {
    const navigate = useNavigate();
    const userState = useRecoilValue(loggedInUserStateAtom);

    const [isSettingsModalOpen, setIsSettingsModalOpen] = useState(false);

    useEffect(() => {
        if(userState.role === Roles.admin && !userState.workspace) {
            setIsSettingsModalOpen(true);
        }
    }, [userState])

    const onSettingsModalClose = () => {
        setIsSettingsModalOpen(false);
    }

    return <div className="header">
        <div>
            <div className="title">
                <img className="entLogo" style={{height: 55}} src={logo} alt="logo" onClick={() => navigate("/")}/>
            </div>  
        </div>
        <div style={{display: "flex", alignItems: "center"}}>
        {userState.role === Roles.admin && <img className="headerIcon" alt="settings" src={settingImg} style={{marginRight: 25}} onClick={() => navigate(RouterRoutes.SETTINGS)}/>}
            <ProfileButton/>
        </div>
        <SettingsModal isOpen={isSettingsModalOpen} onClose={onSettingsModalClose}/>
    </div>;
}


const ProfileButton = () => {

    const navigate = useNavigate();
    const setUserState = useSetRecoilState(loggedInUserStateAtom);

    const logout = async (popupState) => {
        try {
            await api.logout();
            setUserState({
                "username": "",
                "role": "",
                "workspace": {
                    "id": "",
                    "name": "",
                    "namespaces": []
            }
        });
            navigate(RouterRoutes.LOGIN);
        } catch (e) {
            toast.error("Something went wrong, please check the console");
            console.error(e);
        }
        popupState.close();
    }

    return (<PopupState variant="popover" popupId="demo-popup-menu">
        {(popupState) => (
            <React.Fragment>
                <img className="headerIcon" alt="user" src={userImg} {...bindTrigger(popupState)}/>
                <Menu {...bindMenu(popupState)}>
                    <MenuItem style={{fontSize: 12, fontWeight: 600}} onClick={() => logout(popupState)}>
                        <img alt="logout" src={logoutIcon} style={{marginRight: 5, height: 16}}/>
                        Log Out
                    </MenuItem>
                </Menu>
            </React.Fragment>
        )}
    </PopupState>);
}
