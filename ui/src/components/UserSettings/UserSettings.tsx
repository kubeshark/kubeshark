import "./UserSettings.sass"
import {ColsType, FilterableTableAction} from "../UI/FilterableTableAction"
// import Api from "../../helpers/api"
import { useEffect, useState } from "react";
import { UserData,AddUserModal } from "../Modals/AddUserModal/AddUserModal";
import Api from '../../helpers/api';
import { toast } from "react-toastify";
import ConfirmationModal from "../UI/Modals/ConfirmationModal";
import {Utils} from "../../helpers/Utils"


interface Props {

}

const api = Api.getInstance();

enum InviteStatus{
    active = "Active",
    pending = "Pending"
}

export const UserSettings : React.FC<Props> = () => {

    const [usersRows, setUserRows] = useState([]);
    const [userData,userUserData] = useState({} as UserData)
    const cols : ColsType[] = [{field : "username",header:"User", width: '30%'},
                               {field : "workspace",header:"Worksapce", width: '20%'},
                               {field : "role",header:"Role",mapValue : (val) => Utils.capitalizeFirstLetter(val)},
                               {field : "status",header:"Status",getCellClassName : (field, val) =>{
                                   return val === InviteStatus.active ? "status--active" : "status--pending"
                               }}]
    const [isOpenModal,setIsOpen] = useState(false)
    const [editMode, setEditMode] = useState(false);
    const [confirmModalOpen,setConfirmModalOpen] = useState(false)

    const getUserList =         (async () => {
        try {
            let users  = await api.getUsers()
            const mappedUsers = users
            .map((user) => {
                return {...user,status: Utils.capitalizeFirstLetter(user.status), role: user.role, workspace : user?.workspace?.name}
            })
            setUserRows(mappedUsers)                                
        } catch (e) {
            console.error(e);
        }
    })
    

    useEffect(() => {
        getUserList();
        return ()=> setUserRows([]);
    },[])

    const filterFuncFactory = (searchQuery: string) => {
        return (row) => {
            return row.username.toLowerCase().includes(searchQuery.toLowerCase()) ||
                   row.userId.toLowerCase().includes(searchQuery.toLowerCase())
        }
    }

    const searchConfig = { searchPlaceholder: "Search User",filterRows: filterFuncFactory}
    
    const findUser = (userId) => {
        const findFunc = filterFuncFactory(userId);
        return usersRows.find(findFunc);
    }
    
    const onRowDelete = (user) => {
        setConfirmModalOpen(true);
        const userForDel = findUser(user.userId);
        userUserData(userForDel)
    }

    const onConfirmDelete = () => {
        (async() => {
            try {
                await api.deleteUser(userData)
                const findFunc = filterFuncFactory(userData.userId);
                const usersLeft = usersRows.filter(e => !findFunc(e))
                setUserRows(usersLeft)
                toast.success("User Deleted succesesfully")
            } catch (error) {
                toast.error("User want not deleted")
            }
            setConfirmModalOpen(false);
        })()   
    }

    const onRowEdit = (row) => {
        userUserData(row)
        setEditMode(true)
        setIsOpen(true)
    }

    const onUserChange = (user) =>{
        getUserList()
    }

    const buttonConfig = {onClick: () => {setIsOpen(true);setEditMode(false);}, text:"Add User"}

    return (<>
        <FilterableTableAction onRowEdit={onRowEdit} onRowDelete={onRowDelete} searchConfig={searchConfig} 
                               buttonConfig={buttonConfig} rows={usersRows} cols={cols} bodyClass="table-body-style">
        </FilterableTableAction>
        <AddUserModal isOpen={isOpenModal} onCloseModal={() => { 
                     setIsOpen(false);userUserData({} as UserData) } }
                      userData={userData} isEditMode={editMode} onUserChange={onUserChange}>
        </AddUserModal>
        <ConfirmationModal isOpen={confirmModalOpen} onClose={() => setConfirmModalOpen(false)} 
                           onConfirm={onConfirmDelete} confirmButtonText="Delete user" title="Delete User"
                           confirmButtonColor="#DB2156">
            <p>Are you sure you want to delete this user?</p>
        </ConfirmationModal>
    </>);
}
