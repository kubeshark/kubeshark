import { useState } from "react";
import Tabs from "../UI/Tabs"

const AdminSettings: React.FC<any> = ({color}) => {
    var TABS = [
        {tab:"USERS"}, {tab:"WORKSPACE"}
    ];
    
    const [currentTab, setCurrentTab] = useState(TABS[0].tab);
    return (
        <div style={{padding:" 0 1rem"}}>
        <Tabs tabs={TABS} currentTab={currentTab} color={color} onChange={setCurrentTab} leftAligned/>
    </div>
    )
  }

export default AdminSettings;