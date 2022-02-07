import { FC, useEffect, useState } from 'react';
import Api from '../../../helpers/api';
import { useCommonStyles } from '../../../helpers/commonStyle';
import ConfirmationModal from '../../UI/Modals/ConfirmationModal';
import SelectList from '../../UI/SelectList';
import './AddWorkspaceModal.sass'
import { toast } from "react-toastify";
import spinner from "../../assets/spinner.svg";
import LoadingOverlay from "../../LoadingOverlay";

export type WorkspaceData = {
    id:string;
    name:string;
    namespaces: string[];
  }

interface AddWorkspaceModalProp {
  isOpen : boolean,
  onCloseModal: () => void,
  workspaceId: string,
  onEdit: boolean,
  onWorkspaceAdded: () => void,
}
const api = Api.getInstance();

const AddWorkspaceModal: FC<AddWorkspaceModalProp> = ({isOpen,onCloseModal, workspaceId, onEdit, onWorkspaceAdded}) => {

  const [searchValue, setSearchValue] = useState("");

  const [workspaceName, setWorkspaceName] = useState("");
  const [checkedNamespacesKeys, setCheckedNamespacesKeys] = useState([]);
  const [namespaces, setNamespaces] = useState([]);

  const classes = useCommonStyles();
  const [isLoading, setIsLoading] = useState(false);
  const [isSaveLoading, setIsSaveLoading] = useState(false);

  const title = onEdit ? "Edit Workspace" : "Add Workspace";

  useEffect(() => {
    if(!isOpen) return;
    (async () => {
        try {
          if(onEdit){
            setIsLoading(true);
            const workspace = await api.getSpecificWorkspace(workspaceId);
            setWorkspaceName(workspace.name);
            setCheckedNamespacesKeys(workspace.namespaces);   
          }
            setSearchValue("");     
            const namespaces = await api.getNamespaces();
            const namespacesMapped = namespaces.map(namespace => {
              return {key: namespace, value: namespace}
            })
            setNamespaces(namespacesMapped);
    } catch (e) {
            console.error(e);
        }finally{
          setIsLoading(false);
        }
    })()
}, [isOpen,onEdit,workspaceId])

  const onWorkspaceNameChange = (event) => {
    setWorkspaceName(event.target.value);
  }

  const isFormValid = () : boolean => {
    return (workspaceName.length > 0) && (checkedNamespacesKeys.length > 0);
  }

  const onConfirm = async () => {
      try{
        setIsSaveLoading(true);
        const workspaceData = {
          name: workspaceName,
          namespaces: checkedNamespacesKeys
        }
        if(onEdit){
          await api.editWorkspace(workspaceId, workspaceData);
          toast.success("Workspace Succesesfully Updated");
        }
        else{
          await api.createWorkspace(workspaceData);
          toast.success("Workspace Succesesfully Created ");
        }
        onWorkspaceAdded();
        onClose();   
      } catch{
        toast.error("Couldn't Create The Worksapce");
      }
      finally{
        setIsSaveLoading(false);
      }
  }

  const onClose = () => {
    onCloseModal();
    resetForm();
  }

  const resetForm = () => {
    setWorkspaceName("");
    setCheckedNamespacesKeys([]);
    setNamespaces([]);
  }

  return (<>
    <ConfirmationModal isOpen={isOpen} onClose={onClose} onConfirm={onConfirm} title={title} confirmButtonText={"add"} confirmDisabled={!isFormValid()}>
    {isSaveLoading && <LoadingOverlay/>}
      <h3 className='comfirmation-modal__sub-section-header'>DETAILS</h3>
        <div className='comfirmation-modal__sub-section'>
                  <div className="form-input workspace__name">
                    <label htmlFor="inputworkspaceName">Workspace Name</label>
                    <input id="inputworkspaceName" type="text" value={workspaceName ?? ""} className={classes.textField}
                  onChange={onWorkspaceNameChange}/> 
                  </div>
            </div>
            <h3 className='comfirmation-modal__sub-section-header'>TAP SETTINGS</h3>     
          <div className="listSettingsContainer">
            <div style={{marginTop: "17px"}}>
                <input className={classes.textField + " search"} placeholder="Search" value={searchValue}
                        onChange={(event) => setSearchValue(event.target.value)}/>
            </div>
            {isLoading ? <div style={{textAlign: "center", padding: 20}}>
                        <img alt="spinner" src={spinner} style={{height: 35}}/>
                    </div> : <> 
                    <div className='select-list-container'>
            <SelectList items={namespaces}
                        tableName={"Namespaces"}
                        multiSelect={true} 
                        checkedValues={checkedNamespacesKeys} 
                        searchValue={searchValue} 
                        setCheckedValues={setCheckedNamespacesKeys}/>
            </div>
            </>}
          </div>
    </ConfirmationModal>
    </>); 
};

export default AddWorkspaceModal;
