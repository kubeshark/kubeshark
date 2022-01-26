import { FormControl, InputLabel, MenuItem, Select, TextField } from '@material-ui/core';
import React, { FC, useEffect, useState } from 'react';
import Api from '../../../helpers/api';
import { useCommonStyles } from '../../../helpers/commonStyle';
import ConfirmationModal from '../../UI/Modals/ConfirmationModal';
import SelectList from '../../UI/SelectList';
import './AddUserModal.sass';

export type UserData = {
  role:string;
  email : string;
  workspace : string;
}

interface AddUserModalProps {
  isOpen : boolean,
  onCloseModal : () => void
  userData : UserData;
}

const api = Api.getInstance();

export const AddUserModal: FC<AddUserModalProps> = ({isOpen, onCloseModal, userData = {}}) => {

  const [isOpenModal,setIsOpen] = useState(isOpen)
  //const [editUserData, setEditUserData] = useState(userData)
  const [searchValue, setSearchValue] = useState("");
  //const [userRole,setUserRole] = useState("")
  const [workspaces, setWorkspaces] = useState({})
  const roles = [{key:"1",value:"Admin"}]
  const classes = useCommonStyles()

  const [userDataModel, setUserData] = useState(userData as UserData)

  // useEffect(() => {
  //   setIsOpen(isOpen)
  // },[isOpen])

  useEffect(() => {
    (async () => {
        try {
            const workspacesList = {"default":true} //await api.getWorkspaces() 
            setWorkspaces(workspacesList)    
                          
        } catch (e) {
            console.error(e);
        }
    })();
},[])

  useEffect(()=> {
    setUserData(userData as UserData)
  },[userData])

  // const onClose = () => {
  //   setIsOpen(false)
  // }

  const onConfirm = () => {}

  const workspaceChange = (newVal) => {
    setWorkspaces(newVal);
    const  data = {...userDataModel, workspace : newVal}
    setUserData(data)
  }

  const userRoleChange = (e) => {
    const  data = {...userDataModel, role : e.currentTarget.value}
    setUserData(data)
  }

  return (<>
    <ConfirmationModal isOpen={isOpen} onClose={onCloseModal} onConfirm={onConfirm} title='Add User'>
      <h3>DETAILS</h3>
      <div>
        <input type="text" value={userDataModel?.email ?? ""} className={classes.textField + " user__email"} placeholder={"User Email"} 
               onChange={(e) => {}}></input>
        <TextField select size='small' onChange={userRoleChange} value={userDataModel.role}>
          {roles.map((role) => (
                <MenuItem key={role.value} value={role.value}>
                  {role.value}
                </MenuItem>
              ))}
        </TextField>
        {/* <FormControl fullWidth size="small">
          <InputLabel>Select Role</InputLabel>
          <Select className="user__role" label="Select Role" placeholder='Select Role' onChange={userRoleChange} value={userDataModel.role}>

            {roles.map((role) => (
                <MenuItem key={role.value} value={role.value}>
                  {role.value}
                </MenuItem>
              ))}
          </Select>
        </FormControl> */}
      </div>
      <h3>WORKSPACE ACCESS </h3>     
      <div className="namespacesSettingsContainer">
        <div style={{margin: "10px 0"}}>
            <input className={classes.textField + " searchNamespace"} placeholder="Search" value={searchValue}
                    onChange={(event) => setSearchValue(event.target.value)}/>
        </div>
        <SelectList valuesListInput={workspaces} tableName={'Workspace'} multiSelect={false} searchValue={searchValue} setValues={workspaceChange} tabelClassName={''} ></SelectList>
      </div>
    </ConfirmationModal>
    </>); 
};

export default AddUserModal;
