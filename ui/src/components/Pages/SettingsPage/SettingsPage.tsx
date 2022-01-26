import React from "react";
import { useState } from "react";
import Tabs from "../../UI/Tabs"
import { UserSettings } from "../../UserSettings/UserSettings";

const AdminSettings: React.FC<any> = ({color}) => {
    var TABS = [
        {tab:"USERS"}, {tab:"WORKSPACE"}
    ];
    
    const [currentTab, setCurrentTab] = useState(TABS[0].tab);
    return (
        <div style={{padding:" 0 24px"}}>
        <Tabs tabs={TABS} currentTab={currentTab} color={color} onChange={setCurrentTab} leftAligned/>
        {currentTab === TABS[0].tab && <React.Fragment>
                <UserSettings/>
            </React.Fragment>}
        {currentTab === TABS[1].tab && <React.Fragment>
            
        </React.Fragment>}
    </div>
    )
  }

export default AdminSettings;