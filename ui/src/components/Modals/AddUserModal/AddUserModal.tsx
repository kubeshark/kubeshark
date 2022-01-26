import { Button, FormControl, InputLabel, MenuItem, Select } from '@material-ui/core';
import { FC, useEffect, useState } from 'react';
import Api from '../../../helpers/api';
import { useCommonStyles } from '../../../helpers/commonStyle';
import ConfirmationModal from '../../UI/Modals/ConfirmationModal';
import SelectList from '../../UI/SelectList';
import './AddUserModal.sass';
import spinner from "../../assets/spinner.svg";

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
  const [workspaces, setWorkspaces] = useState({})
  const roles = [{key:"1",value:"Admin"}]
  const classes = useCommonStyles()

  const [userDataModel, setUserData] = useState(userData as UserData)
  const isLoading = false;

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

  function isFormValid(): boolean {
    return true;
  }

  const generateLink = () => {
    try {
      api.genareteInviteLink(userDataModel)                
  } catch (e) {
      console.error(e);
  }
    
  }

  const modalCustomActions = <>

                            </>;

  return (<>

    <ConfirmationModal isOpen={isOpen} onClose={onCloseModal} onConfirm={onConfirm} title='Add User'>
      <Button 
                                            className={classes.button + " generate-link-button"} size={"small"} onClick={generateLink}
                                            //disabled={isFormValid()}  
                                            endIcon={isLoading && <img src={spinner} alt="spinner"/>}>
                                              <span className='generate-link-button__icon'></span>
                                              
                                              {"Generate Invite Link"}
      </Button>
      <h3 className='comfirmation-modal__sub-section-header'>DETAILS</h3>
      <div className='comfirmation-modal__sub-section'>
      <div className='user__details'>
        <input type="text" value={userDataModel?.email ?? ""} className={classes.textField + " user__email"} placeholder={"User Email"} 
               onChange={(e) => {}}></input>
        <FormControl size='small' variant="outlined" className='user__role'>
        <InputLabel>User Role</InputLabel>
        <Select value={userDataModel.role} onChange={userRoleChange} >
          <MenuItem value="">
            <em>None</em>
          </MenuItem>
          {roles.map((role) => (
                <MenuItem key={role.value} value={role.value}>
                  {role.value}
                </MenuItem>
              ))}
        </Select>
      </FormControl>
      </div>
      </div>

      <h3 className='comfirmation-modal__sub-section-header'>WORKSPACE ACCESS </h3>     
      <div className="namespacesSettingsContainer">
        <div style={{margin: "10px 0"}}>
            <input className={classes.textField + " searchNamespace"} placeholder="Search" value={searchValue}
                    onChange={(event) => setSearchValue(event.target.value)}/>
        </div>
        <SelectList valuesListInput={workspaces} tableName={''} multiSelect={false} searchValue={searchValue} setValues={workspaceChange} tabelClassName={''} ></SelectList>
      </div>
    </ConfirmationModal>
    </>); 
};

