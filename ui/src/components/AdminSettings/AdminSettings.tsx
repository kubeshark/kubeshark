import BasicTabs from "../UI/BasicTab/BasicTabs";
import Tab from "../UI/BasicTab/Tab";

const AdminSettings: React.FC = () => {
  return (
      <div style={{overflowY:"auto", height:"100%", backgroundColor:"white",color:"$494677",borderRadius:"4px",padding:"10px",position:"relative",display: 'inline-block',}}>
        <BasicTabs>
            <Tab title="USERS">USERS</Tab>
            <Tab title="WORKSPACE">WORKSPACE</Tab>
        </BasicTabs>
    </div>
  )
}

export default AdminSettings;