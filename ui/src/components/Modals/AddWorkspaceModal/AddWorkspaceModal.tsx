import { FC, useEffect, useMemo, useState } from 'react';
import { workerData } from 'worker_threads';
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
export const workspacesDemo = [{id:"1", name:"Worksapce1" , namespaces: [{key:"namespace1", value:"namespace1"},{key:"namespace2", value:"namespace2"}]}]; 
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
            const workspace = workspacesDemo.find(obj => obj.id = workspaceId);
            setWorkspaceName(workspace.name);
            setCheckedNamespacesKeys(workspace.namespaces);   
          }
            setSearchValue("");     
            const namespaces = [{key:"namespace1", value:"namespace1"},{key:"namespace2", value:"namespace2"},{key:"namespace3",value:"namespace3"}];
            setNamespaces(namespaces);
    } catch (e) {
            console.error(e);
        } finally {
        }
    })()
}, [isOpen])

  const onWorkspaceNameChange = (event) => {
    setWorkspaceName(event.target.value);
  }

  // const isFormValid = () : boolean => {
  //   return (Object.values(workspaceDataModal).length === 2) && Object.values(workspaceDataModal).every(val => val !== null)
  // }

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
    <h3 className='headline'>DETAILS</h3>
      <div>
        <input type="text" value={workspaceName ?? ""} className={classes.textField + " workspace__name"} placeholder={"Workspace Name"} 
               onChange={onWorkspaceNameChange}></input>
        </div>
        <h3 className='headline'>TAP SETTINGS</h3>     
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
