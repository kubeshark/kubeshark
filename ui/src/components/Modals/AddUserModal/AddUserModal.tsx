import { Button, FormControl, IconButton, InputAdornment, InputLabel, OutlinedInput } from '@material-ui/core';

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
  inviteToken: string;
  workspace : {name:string, id:string}
}

interface AddUserModalProps {
  isOpen : boolean,
  onCloseModal : () => void,
  userData : UserData,
  isEditMode : boolean,
  onUserChange: (UserData) => void,
}

const api = Api.getInstance();
const formService = new FormService()

export const AddUserModal: FC<AddUserModalProps> = ({isOpen, onCloseModal, userData, isEditMode,onUserChange}) => {

  const [searchValue, setSearchValue] = useState("");
  const [workspaces, setWorkspaces] = useState([])
  const [editMode, setEditMode] = useState(isEditMode);
  const [invite, setInvite] = useState({sent:false,isSuceeded:false,link : null});
  const roles = [{key:"1",value:"admin"},{key:"2",value:"user"}]
  const [isDisplayErrorMessage, setIsDisplayErrorMessage] = useState(false)
  const classes = useCommonStyles()
  const [userDataModel, setUserData] = useState(userData as UserData)
  const [isLoadingGenarate, setIsLoadingGenarate] = useState(false);
  const [isLoadingSave, setIsLoadingSave] = useState(false);

  useEffect(() => {
    (async () => {
        try {          
          const list = await api.getWorkspaces() 
          const workspacesList = list.map((obj) => {return {key:obj.id, value:obj.name,isChecked:false}})
          setWorkspaces(workspacesList)                         
        } catch (e) {
            toast.error("Error finding workspaces")
            console.error(e)
        }
    })();
    return () => setWorkspaces([]);
},[])

  useEffect(()=> {
    if (isEditMode && userData?.inviteToken){
      setInvite({...invite,link : mapTokenToLink(userData?.inviteToken), isSuceeded : true,sent:true})
      userData.workspaceId = userData?.workspace?.id
    }
    
    setEditMode(isEditMode)
    setUserData({...userData} as UserData)
  // eslint-disable-next-line react-hooks/exhaustive-deps
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
        console.error(error)
      }
  }

  const handleFormChange = (prop) => (event) => {
    setUserData({ ...userDataModel, [prop]: event?.target?.value || event[0] });
  };

  const isFormDisabled = () : boolean => {
    return !(userDataModel?.role && userDataModel?.username && userDataModel?.workspaceId && formService.isValidEmail(userDataModel?.username))
  }

  const mapTokenToLink = (token) => {
    return`${window.location.origin}${RouterRoutes.SETUP}/${token}`
  }

  const applyOverlayLoading = (asyncFunc : ()=> Promise<any>, setStateLoadingFunc) => {
    return async () => {
      setStateLoadingFunc(true)
      await asyncFunc()
      setStateLoadingFunc(false)
    }
  }

  const generateLink =  async() => {
      try {
        const res = await api.genareteInviteLink(userDataModel) 
        setInvite({...invite,isSuceeded:true,sent:true, link: mapTokenToLink(res.inviteToken)})
        toast.success("User has been added") 
        onUserChange(userDataModel)   
    } catch (e) {
      toast.error("Error accrued generating link") 
      console.error(e)
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
    console.error(e)
  }
  }

  const isShowInviteLink = () => {
    return ((invite.isSuceeded && invite.link));
  }

  const showGenerateButton = () => {
    return (!invite.isSuceeded || !(invite.link && invite.sent)) 
  }

  const onBlurEmail = (e) => {
    const isValid = formService.isValidEmail(e.target.value)
    const isErrorDisplay = (!isValid && !!userDataModel?.username)
    setIsDisplayErrorMessage(isErrorDisplay)
  }

  const handleCopyInviteLink = (e) => {navigator.clipboard.writeText(invite.link)}

  const addUsermodalCustomActions = <>
      
      <div className={isShowInviteLink() ? "invite-link-row" : ""}>
      {isShowInviteLink() && <FormControl variant="outlined" size={"small"} className='invite-link-field'> 
          <InputLabel htmlFor="outlined-adornment-password">Invite link</InputLabel>
          <OutlinedInput type={'text'} value={invite.link}  classes={{input: "u-input-padding"}}
            endAdornment={
              <InputAdornment position="end">
                <IconButton aria-label="copy invite link" onClick={handleCopyInviteLink} edge="end">
                  {<span className='generate-link-button__icon'></span>}
                </IconButton>
              </InputAdornment>
            } label="Invite link"/>
        </FormControl>}
        {showGenerateButton() && <Button 
                                    className={classes.button + " generate-link-button"} size={"small"} 
                                    onClick={!isEditMode ? applyOverlayLoading(generateLink,setIsLoadingGenarate) : applyOverlayLoading(inviteExistingUser, setIsLoadingGenarate)}
                                    disabled={isFormDisabled()} endIcon={isLoadingGenarate && <img src={spinner} alt="spinner"/>}
                                    startIcon={<span className='generate-link-button__icon'></span>}>{"Generate Invite Link"}
                              </Button>}
           {!isEditMode &&  isShowInviteLink() &&  <Button style={{height: '100%',marginLeft:'20px'}} className={classes.button} size={"small"} onClick={onClose}>
                        Done
            </Button>}                            
              {isEditMode && 
              <Button style={{height: '100%', marginLeft:'20px'}} disabled={isFormDisabled()} className={classes.button} size={"small"} onClick={applyOverlayLoading(updateUser, setIsLoadingSave)} 
                      endIcon={isLoadingSave && <img src={spinner} alt="spinner"/>}>
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
      <div className="form-input user__email">
        <label htmlFor="inputUserEmail">User email</label>
        <input id="inputUserEmail" disabled={editMode} onChange={handleFormChange("username")} onBlur={onBlurEmail}
               value={userDataModel?.username ?? ""} className={classes.textField} placeholder='name@company.com'/>  
        <label className='error-message'>{(isDisplayErrorMessage) ? "*Email is not valid" : ""}</label>  
      </div>
      <div className="form-input user__role">
        <label htmlFor="inputUserRole">User role</label>
        <select value={userDataModel?.role || ""} onChange={handleFormChange("role")} id='inputUserRole' className={classes.textField} >
            <option></option>
        {roles.map((role) => (            
                <option key={role.value} value={role.value}>
                  {Utils.capitalizeFirstLetter(role.value)}
                </option>
              ))}
        </select>
      </div>                  
      </div>
      </div>
      <h3 className='comfirmation-modal__sub-section-header'>WORKSPACE ACCESS </h3>     
      <div className="listSettingsContainer">
        <div style={{marginTop: "17px"}}>
            <input className={classes.textField + " search"} placeholder="Search" value={searchValue}
                    onChange={(event) => setSearchValue(event.target.value)}/>
        </div>
        <div className='select-list-container'>
        <SelectList items={workspaces} tableName={''} multiSelect={false} searchValue={searchValue}
          setCheckedValues={handleFormChange("workspaceId")} checkedValues={[userDataModel?.workspaceId]}/>
        </div>
      </div>
    </ConfirmationModal>
    </>); 
};