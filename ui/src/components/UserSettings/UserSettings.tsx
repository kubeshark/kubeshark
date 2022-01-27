import "./UserSettings.sass"
import {ColsType, FilterableTableAction} from "../UI/FilterableTableAction"
// import Api from "../../helpers/api"
import { useEffect, useState } from "react";
import { UserData,AddUserModal } from "../Modals/AddUserModal/AddUserModal";
import {Snackbar} from "@material-ui/core";
import MuiAlert from "@material-ui/lab/Alert";
import { Select } from "../UI/Select";
import { MenuItem } from "@material-ui/core";
import { settings } from "cluster";
import { SettingsModal } from "../SettingsModal/SettingModal";
import OasModal from "../Modals/OasModal/OasModal";

interface Props {

}

// const api = Api.getInstance();

export const UserSettings : React.FC<Props> = ({}) => {

    const [usersRows, setUserRows] = useState([]);
    const [userData,SetUsetData] = useState({} as UserData)
    const cols : ColsType[] = [{field : "userName",header:"User"},
                               {field : "role",header:"Role"},
                               {field : "status",header:"Status",getCellClassName : (field, val) =>{
                                   return val === "Active" ? "status--active" : "status--pending"
                               }}]
    const [isOpenModal,setIsOpen] = useState(false)
    const [alert,setAlert] = useState({open:false,sevirity:"success"})
    

    useEffect(() => {
        (async () => {
            try {
                const users = [{userName:"asd",role:"Admin",status:"Active"}]//await api.getUsers() 
                setUserRows(users)                                
            } catch (e) {
                console.error(e);
            }
        })();
    },[])

    const filterFuncFactory = (searchQuery: string) => {
        return (row) => {
            return row.userName.toLowerCase().includes(searchQuery.toLowerCase())
        }
    }

    const searchConfig = { searchPlaceholder: "Search User",filterRows: filterFuncFactory}
    
    const onRowDelete = (row) => {
        const filterFunc = filterFuncFactory(row.userName)
        const newUserList = usersRows.filter(filterFunc)
        setUserRows(newUserList)
    }

    const onRowEdit = (row) => {
        SetUsetData(row)
        setIsOpen(true)
    }



    const buttonConfig = {onClick: () => {setIsOpen(true)}, text:"Add User"}

    return (<>
        <FilterableTableAction onRowEdit={onRowEdit} onRowDelete={onRowDelete} searchConfig={searchConfig} 
                               buttonConfig={buttonConfig} rows={usersRows} cols={cols}>
        </FilterableTableAction>
        <AddUserModal isOpen={isOpenModal} onCloseModal={() => { setIsOpen(false); } } userData={userData} setShowAlert={setAlert}>
        </AddUserModal>
        <Snackbar open={alert.open} classes={{root: "alert--right"}}>
        <MuiAlert classes={{filledWarning: 'customWarningStyle'}} elevation={6} variant="filled"
                  onClose={() => setAlert({...alert,open:false})} severity={"success"}>
                    User has been added
        </MuiAlert>
    </Snackbar>
        {/* <SettingsModal isOpen={false} onClose={function (): void {
            throw new Error("Function not implemented.");
        } } isFirstLogin={false}></SettingsModal> */}
    </>);
}
