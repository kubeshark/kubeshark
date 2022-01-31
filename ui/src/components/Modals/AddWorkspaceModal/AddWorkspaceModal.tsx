import { FC, useEffect, useState } from 'react';
import Api from '../../../helpers/api';
import { useCommonStyles } from '../../../helpers/commonStyle';
import ConfirmationModal from '../../UI/Modals/ConfirmationModal';
import SelectList from '../../UI/SelectList';
import './AddWorkspaceModal.sass'
import { toast } from "react-toastify";

export type WorkspaceData = {
    id:string;
    name:string;
    namespaces: string[];
  }

interface AddWorkspaceModalProp {
  isOpen : boolean,
  onCloseModal: () => void,
  workspaceId: string,
  onEdit: boolean
}
const api = Api.getInstance();

const AddWorkspaceModal: FC<AddWorkspaceModalProp> = ({isOpen,onCloseModal, workspaceId, onEdit}) => {

  const [searchValue, setSearchValue] = useState("");

  const [workspaceName, setWorkspaceName] = useState("");
  const [checkedNamespacesKeys, setCheckedNamespacesKeys] = useState([]);
  const [namespaces, setNamespaces] = useState([]);

  const classes = useCommonStyles();

  const title = onEdit ? "Edit Workspace" : "Add Workspace";

  useEffect(() => {
    if(!isOpen) return;
    (async () => {
        try {
          if(onEdit){
            const workspace = await api.getSpecificWorkspace(workspaceId);
            setWorkspaceName(workspace.name);
            setCheckedNamespacesKeys(workspace.namespaces);   
          }
            setSearchValue("");     
            const namespaces = await api.getTapConfig();
            const namespacesMapped = namespaces.map(namespace => {
              return {key: namespace, value: namespace}
            })
            setNamespaces(namespacesMapped);
    } catch (e) {
            console.error(e);
        } finally {
        }
    })()
}, [isOpen])

  const onWorkspaceNameChange = (event) => {
    setWorkspaceName(event.target.value);
  }

  const isFormValid = () : boolean => {
    return (workspaceName.length > 0) && (checkedNamespacesKeys.length > 0);
  }

  const onConfirm = () => {
    try{
      const workspaceData = {
        name: workspaceName,
        namespaces: checkedNamespacesKeys
      }
      console.log(workspaceData);
      onCloseModal();
      toast.success("Workspace Succesesfully Created ");
    } catch{
      toast.error("Couldn't Creat The Worksapce");
    }
  }

  const onClose = () => {
    onCloseModal();
    setWorkspaceName("");
    setCheckedNamespacesKeys([]);
    setNamespaces([]);
  }

  return (<>
    <ConfirmationModal isOpen={isOpen} onClose={onClose} onConfirm={onConfirm} title={title}>
      <h3 className='comfirmation-modal__sub-section-header'>DETAILS</h3>
        <div className='comfirmation-modal__sub-section'>
          <div>
            <input type="text" value={workspaceName ?? ""} className={classes.textField + " workspace__name"} placeholder={"Workspace Name"} 
                  onChange={onWorkspaceNameChange}></input>
            </div>
            </div>
            <h3 className='comfirmation-modal__sub-section-header'>TAP SETTINGS</h3>     
          <div className="namespacesSettingsContainer">
            <div>
                <input className={classes.textField + " searchNamespace"} placeholder="Search" value={searchValue}
                        onChange={(event) => setSearchValue(event.target.value)}/>
            </div>
            <SelectList items={namespaces}
                        tableName={"Namespaces"}
                        multiSelect={true} 
                        checkedValues={checkedNamespacesKeys} 
                        searchValue={searchValue} 
                        setCheckedValues={setCheckedNamespacesKeys} 
                        tabelClassName={undefined}>
                        </SelectList>
          </div>
    </ConfirmationModal>
    </>); 
};

export default AddWorkspaceModal;
