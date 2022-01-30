import { Button, FormControl, IconButton, InputAdornment, InputLabel, MenuItem, OutlinedInput, Select } from '@material-ui/core';

import { FC, useEffect, useState } from 'react';
import Api from '../../../helpers/api';
import { useCommonStyles } from '../../../helpers/commonStyle';
import ConfirmationModal from '../../UI/Modals/ConfirmationModal';
import {toast} from "react-toastify";
import SelectList from '../../UI/SelectList';
import './AddUserModal.sass';
import spinner from "../../assets/spinner.svg";
import { values } from 'mobx';

export type UserData = {
  role:string;
  username : string;
  workspace : string;
  userId: string;
}

interface AddUserModalProps {
  isOpen : boolean,
  onCloseModal : () => void,
  userData : UserData,
  isEditMode : boolean,
  onUserChange: (UserData) => void,
}

const api = Api.getInstance();

export const AddUserModal: FC<AddUserModalProps> = ({isOpen, onCloseModal, userData, isEditMode,onUserChange}) => {


  //const [editUserData, setEditUserData] = useState(userData)
  const [searchValue, setSearchValue] = useState("");
  const [workspaces, setWorkspaces] = useState([])
  //const { control, handleSubmit,register } = useForm<UserData>();
  const [disable, setDisable] = useState(true);
  const [editMode, setEditMode] = useState(isEditMode);
  const [invite, setInvite] = useState({sent:false,isSuceeded:false,link : null});
  const roles = [{key:"1",value:"Admin"},{key:"2",value:"User"}]
  const classes = useCommonStyles()

  const [userDataModel, setUserData] = useState(userData as UserData)
  const isLoading = false;

  // useEffect(() => {
  //   setIsOpen(isOpen)
  // },[isOpen])

  useEffect(() => {
    (async () => {
        try {
            const workspacesList = [
              {
                  "id": "f54b18ec-aa15-4b2c-a4d5-8eda17e44c93",
                  "name": "sock-shop"
              },
              {
                  "id": "c7ad9158-d840-46c0-b5ce-2487c013723f",
                  "name": "test"
              }
          ].map((obj) => {return {key:obj.id, value:obj.name,isChecked:false}})
          //await api.getWorkspaces() 
          setWorkspaces(workspacesList)    
                          
        } catch (e) {
            toast.error("Error finding workspaces")
        }
    })();
},[])

  useEffect(()=> {
    (async () => {
      try { 
          setEditMode(isEditMode)
          if (isEditMode) {
            
            //const userDetails = await api.getUserDetails(userData)
            //const data = {...userData,...userDetails}
            
          }
          else{
            
          }
          setUserData(userData as UserData)
      } catch (e) {
          toast.error("Error getting user details")
      }
  })();
  },[isEditMode, userData])

  // const onClose = () => {
  //   setIsOpen(false)
  // }

  const onClose = () => {
    onCloseModal()
    setUserData({} as UserData)
    setInvite({sent:false,isSuceeded:false,link:""})
  }

  const workspaceChange = (workspaces) => {
    //setWorkspaces(newVal);
    const selectedWorksapce = workspaces.find(x=> x.isChecked)
    const  data = {...userDataModel, workspace : selectedWorksapce.key}
    setUserData(data)
    setGenarateDisabledState()
  }

  const userRoleChange = (e) => {
    const  data = {...userDataModel, role : e.target.value}
    setUserData(data)
    setGenarateDisabledState()
  }

  const userNameChange = (e) => {
    const  data = {...userDataModel, username : e.currentTarget.value}
    setUserData(data)
    setGenarateDisabledState()
  }

  const handleChange = (prop) => (event) => {
    //setValues({ ...values, [prop]: event.target.value });
  };

  const isFormValid = () : boolean => {
    return true;
    //return (Object.values(userDataModel).length === 3) && Object.values(userDataModel).every(val => val !== null)
  }

  const setGenarateDisabledState = () => {
    const isValid = isFormValid()
    setDisable(!isValid)
  }

  const generateLink =  () => {
    try {
      if (editMode) {
        //await api.updateUser(userDataModel)
        setInvite({...invite,isSuceeded:true,sent:true,link:"asdasdasdasdasdasdasdasdads"})
        toast.success("User has been modified")  
      }
      else{
        //const res = await api.genareteInviteLink(userDataModel) 
        setInvite({...invite,isSuceeded:true,sent:true, link:"asdasdasdasdasdasdasdasdads"})
        toast.success("User has been added") 
      }

      onUserChange(userDataModel)
   
  } catch (e) {
    toast.error("Error accrued generating link") 
  }
    
  }

  const handleCopyinviteLink = (e) => {navigator.clipboard.writeText(invite.link)}

  const modalCustomActions = <>
            {(!invite.isSuceeded || !(invite.link && invite.sent)) && <Button 
                                            className={classes.button + " generate-link-button"} size={"small"} onClick={generateLink}
                                            disabled={disable}  
                                            endIcon={isLoading && <img src={spinner} alt="spinner"/>}>
                                              <span className='generate-link-button__icon'></span>
                                              {"Generate Invite Link"}
                              </Button>}
                              {invite.isSuceeded && invite.link && <div className="invite-link-row">
        <FormControl variant="outlined" size={"small"} className='invite-link-field'>
          <InputLabel htmlFor="outlined-adornment-password">Invite link</InputLabel>
          <OutlinedInput 
            type={'text'}
            value={invite.link}           
            onChange={handleChange('password')}
            endAdornment={
              <InputAdornment position="end">
                <IconButton aria-label="cpoy invite link" onClick={handleCopyinviteLink} edge="end">
                  {<span className='generate-link-button__icon'></span>}
                </IconButton>
              </InputAdornment>
            }
            label="Invite link"
          />
        </FormControl>
            <Button style={{height: '100%'}} className={classes.button} size={"small"} onClick={onClose}>
                        Done
            </Button>
      </div>}
                            </>;

  return (<>

    <ConfirmationModal isOpen={isOpen} onClose={onClose} onConfirm={onClose} 
                       title={`${editMode ? "Edit" : "Add"} User`} customActions={modalCustomActions}>

      <h3 className='comfirmation-modal__sub-section-header'>DETAILS</h3>
      <div className='comfirmation-modal__sub-section'>
      <div className='user__details'>
        <input type="text" value={userDataModel?.username ?? ""} className={classes.textField + " user__email"} 
                placeholder={"User Email"} onChange={userNameChange} disabled={editMode}>
        </input>
        
              {/* <Controller name="role" control={control} rules={{ required: true }}
        render={({ field }) =>    }
      /> */}

      <FormControl size='small' variant="outlined" className='user__role'>
        <InputLabel>User Role</InputLabel>
        <Select value={userDataModel.role ?? ""} onChange={userRoleChange} classes={{ root : 'my-class-name' }} >
          <MenuItem value="0">
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
            <input className={classes.textField + " search-workspace"} placeholder="Search" value={searchValue}
                    onChange={(event) => setSearchValue(event.target.value)}/>
        </div>
        <SelectList valuesListInput={workspaces} tableName={''} multiSelect={false} searchValue={searchValue} 
                    setValues= {workspaceChange} tabelClassName={''} >
        </SelectList>
      </div>

    </ConfirmationModal>
    </>); 
};