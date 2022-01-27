import { FC, useEffect, useState } from 'react';
import Api from '../../../helpers/api';
import { useCommonStyles } from '../../../helpers/commonStyle';
import ConfirmationModal from '../../UI/Modals/ConfirmationModal';
import SelectList from '../../UI/SelectList';
import './AddWorkspaceModal.sass'

export type WorkspaceData = {
    name:string;
    namespaces: string[];
  }

interface AddWorkspaceModalProp {
  isOpen : boolean,
  onCloseModal: () => void,
  workspaceData: WorkspaceData,
  onEdit: boolean
}

const api = Api.getInstance();

const AddWorkspaceModal: FC<AddWorkspaceModalProp> = ({isOpen,onCloseModal, workspaceData ={}, onEdit}) => {

  const [workspaceDataModel, setUserData] = useState(workspaceData as WorkspaceData);
  const [searchValue, setSearchValue] = useState("");
  const classes = useCommonStyles()
  const [namespaces, setNamespaces] = useState({});

  const title = onEdit ? "Edit Workspace" : "Add Workspace";


  useEffect(() => {
    if(!isOpen) return;
    (async () => {
        try {
            setSearchValue(""); 
            const tapConfig = await api.getTapConfig();
            console.log(tapConfig);
            // if(isFirstLogin) {
                const namespacesObj = {...tapConfig?.tappedNamespaces}
                Object.keys(tapConfig?.tappedNamespaces ?? {}).forEach(namespace => {
                    namespacesObj[namespace] = true;
                })
                setNamespaces(namespacesObj);
            // } else {
                setNamespaces(tapConfig?.tappedNamespaces);
            // }
        } catch (e) {
            console.error(e);
        } finally {
        }
    })()
}, [isOpen])

  const onConfirm = () => {}

  return (<>
    <ConfirmationModal isOpen={isOpen} onClose={onCloseModal} onConfirm={onConfirm} title={title}>
    <h3 className='headline'>DETAILS</h3>
      <div>
        <input type="text" value={workspaceData?.name ?? ""} className={classes.textField + " workspace__name"} placeholder={"Workspace Name"} 
               onChange={(e) => {}}></input>
        </div>
        <h3 className='headline'>TAP SETTINGS</h3>     
      <div className="namespacesSettingsContainer">
        <div>
            <input className={classes.textField + " searchNamespace"} placeholder="Search" value={searchValue}
                    onChange={(event) => setSearchValue(event.target.value)}/>
        </div>
        <SelectList valuesListInput={namespaces} tableName={"Namespaces"} multiSelect={true} searchValue={searchValue} setValues={setUserData} tabelClassName={undefined}></SelectList>
            </div>
    </ConfirmationModal>
    </>); 
};

export default AddWorkspaceModal;
