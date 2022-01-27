import "./UserSettings.sass"
import {ColsType, FilterableTableAction} from "../UI/FilterableTableAction"
// import Api from "../../helpers/api"
import { useEffect, useState } from "react";
import { UserData,AddUserModal } from "../Modals/AddUserModal/AddUserModal";
import Api from '../../helpers/api';

import {Snackbar} from "@material-ui/core";
import MuiAlert from "@material-ui/lab/Alert";
import { Select } from "../UI/Select";
import { MenuItem } from "@material-ui/core";
import { settings } from "cluster";
import { SettingsModal } from "../SettingsModal/SettingModal";
import OasModal from "../Modals/OasModal/OasModal";
import { apiDefineProperty } from "mobx/dist/internal";
import { toast } from "react-toastify";

interface Props {

}

const api = Api.getInstance();

export const UserSettings : React.FC<Props> = ({}) => {

    const [usersRows, setUserRows] = useState([]);
    const [userData,userUserData] = useState({} as UserData)
    const cols : ColsType[] = [{field : "username",header:"User"},
                               {field : "role",header:"Role"},
                               {field : "status",header:"Status",getCellClassName : (field, val) =>{
                                   return val === "Active" ? "status--active" : "status--pending"
                               }}]
    const [isOpenModal,setIsOpen] = useState(false)
    const [alert,setAlert] = useState({open:false,sevirity:"success"})
    

    useEffect(() => {
        (async () => {
            try {
                const users = [{username:"asd",role:"Admin",status:"Active",userId : "1"},
                               {username:"aaaaaaa",role:"User",status:"Active",userId : "2"}]//await api.getUsers() 
                setUserRows(users)                                
            } catch (e) {
                console.error(e);
            }
        })();
    },[])

    const filterFuncFactory = (searchQuery: string) => {
        return (row) => {
            return row.username.toLowerCase().includes(searchQuery.toLowerCase()) ||
                   row.userId.toLowerCase().includes(searchQuery.toLowerCase())
        }
    }

    const searchConfig = { searchPlaceholder: "Search User",filterRows: filterFuncFactory}
    
    const onRowDelete = (user) => {
        const findFunc = filterFuncFactory(user.userId);
        const userToDelete = usersRows.find(findFunc);
        (async() => {
            try {
                //await api.deleteUser(user)
                const usersLeft = usersRows.filter(e => !findFunc(e))
                setUserRows(usersLeft)
                toast.success("User Deleted succesesfully")
            } catch (error) {
                toast.error("User want not deleted")
            }
        })()
        
    }

    const onRowEdit = (row) => {
        userUserData(row)
        setIsOpen(true)
    }



    const buttonConfig = {onClick: () => {setIsOpen(true)}, text:"Add User"}

    return (<>
        <FilterableTableAction onRowEdit={onRowEdit} onRowDelete={onRowDelete} searchConfig={searchConfig} 
                               buttonConfig={buttonConfig} rows={usersRows} cols={cols}>
        </FilterableTableAction>
        <AddUserModal isOpen={isOpenModal} onCloseModal={() => { setIsOpen(false);userUserData({} as UserData) } } userData={userData} setShowAlert={setAlert}>
        </AddUserModal>
        {/* <Snackbar open={alert.open} classes={{root: "alert--right"}}>
        <MuiAlert classes={{filledWarning: 'customWarningStyle'}} elevation={6} variant="filled"
                  onClose={() => setAlert({...alert,open:false})} severity={"success"}>
                    User has been added
        </MuiAlert>
    </Snackbar> */}
        {/* <SettingsModal isOpen={false} onClose={function (): void {
            throw new Error("Function not implemented.");
        } } isFirstLogin={false}></SettingsModal> */}
    </>);
}
