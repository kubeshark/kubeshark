import { Button, FormControl, IconButton, InputAdornment, InputLabel, MenuItem, OutlinedInput,Select, TextField } from '@material-ui/core';

import { FC, useEffect, useState } from 'react';
import Api from '../../../helpers/api';
import { useCommonStyles } from '../../../helpers/commonStyle';
import ConfirmationModal from '../../UI/Modals/ConfirmationModal';
import {toast} from "react-toastify";
import SelectList from '../../UI/SelectList';
import './AddUserModal.sass';
import spinner from "../../assets/spinner.svg";
import {FormService} from "../../../helpers/FormService"
import {RouterRoutes} from "../../../helpers/routes";
import {Utils} from "../../../helpers/Utils"


export type UserData = {
  role:string;
  username : string;
  workspaceId : string;
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
const fromService = new FormService()

export const AddUserModal: FC<AddUserModalProps> = ({isOpen, onCloseModal, userData, isEditMode,onUserChange}) => {


  //const [editUserData, setEditUserData] = useState(userData)
  const [searchValue, setSearchValue] = useState("");

  const [workspaces, setWorkspaces] = useState([])

  //const { control, handleSubmit,register } = useForm<UserData>();
  const [editMode, setEditMode] = useState(isEditMode);
  const [invite, setInvite] = useState({sent:false,isSuceeded:false,link : null});
  const roles = [{key:"1",value:"admin"},{key:"2",value:"user"}]
  const [isDisplayErrorMessage, setIsDisplayErrorMessage] = useState(false)
  const classes = useCommonStyles()

  const [userDataModel, setUserData] = useState(userData as UserData)
  const isLoading = false;

  // useEffect(() => {
  //   setIsOpen(isOpen)
  // },[isOpen])

  useEffect(() => {
    (async () => {
        try {          
          const list = await api.getWorkspaces() 
          const workspacesList = list.map((obj) => {return {key:obj.id, value:obj.name,isChecked:false}})
          setWorkspaces(workspacesList)                         
        } catch (e) {
            toast.error("Error finding workspaces")
        }
    })();
    return () => setWorkspaces([]);
},[])

  useEffect(()=> {
    (async () => {
      try {
          let specificUser : any = {}
          if (isEditMode && userData.userId){
            specificUser= await api.getUserDetails(userData)
          }
          
          setEditMode(isEditMode)
          setUserData({...userData,workspaceId : specificUser.workspaceId} as UserData)
      } catch (e) {
          toast.error("Error getting user details")
      }
  })();
  },[isEditMode, userData])

  const onClose = () => {
    onCloseModal()
    setUserData({} as UserData)
    setInvite({sent:false,isSuceeded:false,link:""})
    setEditMode(false)
    setSearchValue("")
  }

  const updateUser = async() =>{
      try {
        await api.updateUser(userDataModel)
        toast.success("User has been modified")  
      } catch (error) {
        toast.error("Error accured modifing user")  
      }
  }

  const workspaceChange = (workspaces) => {
    const  data = {...userDataModel, workspaceId : workspaces.length ? workspaces[0] : ""}
    setUserData((prevState) => {return data});
  }

  const userRoleChange = (e) => {
    const  data = {...userDataModel, role : e.target.value}
    setUserData(data)
  }

  const userNameChange = (e) => {
    const  data = {...userDataModel, username : e.currentTarget.value}
    setUserData(data)
  }

  const handleChange = (prop) => (event) => {
    //setValues({ ...values, [prop]: event.target.value });
  };

  const isFormDisabled = () : boolean => {
    return !(userDataModel?.role && userDataModel?.username && userDataModel?.workspaceId)
  }


  const mapTokenToLink = (token) => {
    return`${window.location.origin}${RouterRoutes.SETUP}/${token}`
  }

  const generateLink =  async() => {
      try {
        const res = await api.genareteInviteLink(userDataModel) 
        setInvite({...invite,isSuceeded:true,sent:true, link: mapTokenToLink(res.inviteToken)})
        toast.success("User has been added") 
        onUserChange(userDataModel)   
    } catch (e) {
      toast.error("Error accrued generating link") 
    }
  }

  const inviteExistingUser = async() => {
    try {
      const res = await api.inviteExistingUser(userDataModel.userId) 
      setInvite({...invite,isSuceeded:true,sent:true, link: mapTokenToLink(res.inviteToken)})
      toast.success("Invite link created") 
      onUserChange(userDataModel)   
  } catch (e) {
    toast.error("Error accrued generating link") 
  }
  }

  const isShowInviteLink = () => {
    return ((invite.isSuceeded && invite.link));
  }

  const showGenerateButton = () => {
    return (!invite.isSuceeded || !(invite.link && invite.sent)) 
  }

  const onBlurEmail = (e) => {
    const isValid = fromService.isValidEmail(e.target.value)
    const isErrorDisplay = (!isValid && !!userDataModel?.username)
    setIsDisplayErrorMessage(isErrorDisplay)
  }

  const handleCopyinviteLink = (e) => {navigator.clipboard.writeText(invite.link)}

  const addUsermodalCustomActions = <>
      
      <div className={isShowInviteLink() ? "invite-link-row" : ""}>
      {isShowInviteLink() && <FormControl variant="outlined" size={"small"} className='invite-link-field'> 
          <InputLabel htmlFor="outlined-adornment-password">Invite link</InputLabel>
          <OutlinedInput type={'text'} value={invite.link} onChange={handleChange('password')}  classes={{input: "u-input-padding"}}
            endAdornment={
              <InputAdornment position="end">
                <IconButton aria-label="copy invite link" onClick={handleCopyinviteLink} edge="end">
                  {<span className='generate-link-button__icon'></span>}
                </IconButton>
              </InputAdornment>
            } label="Invite link"/>
        </FormControl>}
        {showGenerateButton() && <Button 
                                            className={classes.button + " generate-link-button"} size={"small"} 
                                            onClick={!isEditMode ? generateLink : inviteExistingUser}
                                            disabled={isFormDisabled()}  
                                            endIcon={isLoading && <img src={spinner} alt="spinner"/>}
                                            startIcon={<span className='generate-link-button__icon'></span>}>
                                              {"Generate Invite Link"}
                              </Button>}
           {!isEditMode &&  isShowInviteLink() &&  <Button style={{height: '100%',marginLeft:'20px'}} className={classes.button} size={"small"} onClick={onClose}>
                        Done
            </Button>}                            
              {isEditMode && <Button style={{height: '100%', marginLeft:'20px'}} disabled={isFormDisabled()} className={classes.button} size={"small"} onClick={updateUser}>
              Save
            </Button>
          }

      </div>

   </>;

  return (<>

    <ConfirmationModal isOpen={isOpen} onClose={onClose} onConfirm={onClose} 
                       title={`${editMode ? "Edit" : "Add"} User`} customActions={addUsermodalCustomActions}>

      <h3 className='comfirmation-modal__sub-section-header'>DETAILS</h3>
      <div className='comfirmation-modal__sub-section'>
      <div className='user__details'>
        <TextField value={userDataModel?.username ?? ""} className={"user__email"} size='small' label="User email" onChange={userNameChange} onBlur={onBlurEmail}
                   error={isDisplayErrorMessage} helperText={(isDisplayErrorMessage) ? "*Email is not valid" : ""} variant="outlined" disabled={editMode}/>
        


  {/* className='user__role u-input-padding' */}
  {/* classes={{ select : 'u-input-padding' }}  */}
      <FormControl variant="outlined" className='user__role' size={"small"}>
        <InputLabel id="user-role-outlined-label">User role</InputLabel>
        <Select labelId="user-role-outlined-label" id="demo-simple-select-outlined" value={userDataModel.role ?? ""} onChange={userRoleChange} label="User role">
           {roles.map((role) => (
                <MenuItem key={role.value} value={role.value}>
                  {Utils.capitalizeFirstLetter(role.value)}
                </MenuItem>
              ))}
        </Select>
      </FormControl>
      </div>
      </div>
      <h3 className='comfirmation-modal__sub-section-header'>WORKSPACE ACCESS </h3>     
      <div className="namespacesSettingsContainer">
        <div style={{marginTop: "17px"}}>
            <input className={classes.textField + " search-workspace"} placeholder="Search" value={searchValue}
                    onChange={(event) => setSearchValue(event.target.value)}/>
        </div>
        <SelectList items={workspaces} tableName={''} multiSelect={false} searchValue={searchValue}
        setCheckedValues={workspaceChange} tabelClassName={''} checkedValues={[userDataModel.workspaceId]} >
        </SelectList>
      </div>

    </ConfirmationModal>
    </>); 
};