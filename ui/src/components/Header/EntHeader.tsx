import React, {useContext, useEffect, useState} from "react";
import logo from '../assets/MizuEntLogo.svg';
import './Header.sass';
import userImg from '../assets/user-circle.svg';
import settingImg from '../assets/settings.svg';
import {Menu, MenuItem} from "@material-ui/core";
import PopupState, {bindMenu, bindTrigger} from "material-ui-popup-state";
import logoutIcon from '../assets/logout.png';
import {SettingsModal} from "../SettingsModal/SettingModal";
import Api from "../../helpers/api";
import {toast} from "react-toastify";
import {MizuContext, Page} from "../../EntApp";

const api = Api.getInstance();

interface EntHeaderProps {
    isFirstLogin: boolean;
    setIsFirstLogin: (flag: boolean) => void
}

export const EntHeader: React.FC<EntHeaderProps> = ({isFirstLogin, setIsFirstLogin}) => {

    const [isSettingsModalOpen, setIsSettingsModalOpen] = useState(false);

    useEffect(() => {
        if(isFirstLogin) {
            setIsSettingsModalOpen(true)
        }
    }, [isFirstLogin])

    const onSettingsModalClose = () => {
        setIsSettingsModalOpen(false);
        setIsFirstLogin(false);
    }

    return <div className="header">
        <div>
            <div className="title">
                <img style={{height: 55}} src={logo} alt="logo"/>
            </div>
        </div>
        <div style={{display: "flex", alignItems: "center"}}>
            <img className="headerIcon" alt="settings" src={settingImg} style={{marginRight: 25}} onClick={() => setIsSettingsModalOpen(true)}/>
            <ProfileButton/>
        </div>
        <SettingsModal isOpen={isSettingsModalOpen} onClose={onSettingsModalClose} isFirstLogin={isFirstLogin}/>
    </div>;
}

const ProfileButton = () => {

    const {setPage} = useContext(MizuContext);

    const logout = async (popupState) => {
        try {
            await api.logout();
            setPage(Page.Login);
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
