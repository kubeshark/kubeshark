import React from "react";
import { useState } from "react";
import Tabs from "../../UI/Tabs"
import { UserSettings } from "../../UserSettings/UserSettings";
import { WorkspaceSettings } from "../../WorkspaceSettings/WorkspaceSettings";
import "./SettingsPage.sass"

const AdminSettings: React.FC<any> = ({color}) => {
    var TABS = [
        {tab:"USERS"}, {tab:"WORKSPACE"}
    ];
    
    const [currentTab, setCurrentTab] = useState(TABS[0].tab);
    return (<>
        <div className="settings-page">
            <div className="header-section">
                <div className="header-section__container">
                    <div className="header-section__title">Settings</div>
                    <Tabs tabs={TABS} currentTab={currentTab} color={color} onChange={setCurrentTab} leftAligned classes={{root:"tabs-nav"}}/>
                </div>
            </div>
            <div className="tab-content">
                <div className="tab-content__container">
                {currentTab === TABS[0].tab && <React.Fragment>
                        <UserSettings/>
                    </React.Fragment>}
                {currentTab === TABS[1].tab && <React.Fragment>
                    <WorkspaceSettings/>
                </React.Fragment>}
                </div>
            </div>
    </div>
    </>
    )
  }

export default AdminSettings;