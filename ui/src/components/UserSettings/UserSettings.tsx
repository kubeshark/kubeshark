import "./UserSettings.sass"
import {useCommonStyles} from "../../helpers/commonStyle";
import {ColsType, FilterableTableAction} from "../UI/FilterableTableAction"
import Api from "../../helpers/api"
import { useEffect, useState } from "react";
import AddUserModal, { UserData } from "../Modals/AddUserModal/AddUserModal";
import { Select } from "../UI/Select";
import { MenuItem } from "@material-ui/core";
import { settings } from "cluster";
import { SettingsModal } from "../SettingsModal/SettingModal";
import OasModal from "../Modals/OasModal/OasModal";

interface Props {

}

const api = Api.getInstance();

export const UserSettings : React.FC<Props> = ({}) => {

    const [usersRows, setUserRows] = useState([]);
    const [userData,SetUsetData] = useState({} as UserData)
    const cols : ColsType[] = [{field : "userName",header:"User"},
                               {field : "role",header:"Role"},
                               {field : "status",header:"Status",getCellClassName : (field, val) =>{
                                   return val === "Active" ? "status--active" : "status--pending"
                               }}]
    const [isOpenModal,setIsOpen] = useState(false)
    

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
        // open Edit user Modal
    }



    const buttonConfig = {onClick: () => {setIsOpen(true)}, text:"Add User"}

    return (<>
        <FilterableTableAction onRowEdit={onRowEdit} onRowDelete={onRowDelete} searchConfig={searchConfig} 
                               buttonConfig={buttonConfig} rows={usersRows} cols={cols}>
        </FilterableTableAction>
        <AddUserModal isOpen={isOpenModal} onCloseModal={() => { setIsOpen(false); } } userData={userData}></AddUserModal>
        {/* <SettingsModal isOpen={false} onClose={function (): void {
            throw new Error("Function not implemented.");
        } } isFirstLogin={false}></SettingsModal> */}
    </>);
}
