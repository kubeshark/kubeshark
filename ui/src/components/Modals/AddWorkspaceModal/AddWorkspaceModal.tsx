import React, { FC, useEffect, useState } from 'react';
import { useCommonStyles } from '../../../helpers/commonStyle';
import ConfirmationModal from '../../UI/Modals/ConfirmationModal';
import SelectList from '../../UI/SelectList';
import './AddWorkspaceModal.sass'
// import './AddUserModal.sass';

export type WorkspaceData = {
    name:string;
    namespaces: string[];
  }

interface AddWorkspaceModal {
  isOpen : boolean,
  onCloseModal: () => void,
  workspaceData: WorkspaceData
}

const AddWorkspaceModal: FC<AddWorkspaceModal> = ({isOpen,onCloseModal, workspaceData ={}}) => {

  const [isOpenModal,setIsOpen] = useState(isOpen);
  const [workspaceDataModel, setUserData] = useState(workspaceData as WorkspaceData);
  const [searchValue, setSearchValue] = useState("");
  const classes = useCommonStyles()
  const [namespaces, SetNamespaces] = useState({})

  useEffect(() => {
    setIsOpen(isOpen)
  },[isOpen])

  useEffect(() => {
    (async () => {
        try {
            const namespacesList = {"default": false, "blabla": false, "test":true};
            SetNamespaces(namespacesList)    
                          
        } catch (e) {
            console.error(e);
        }
    })();
},[])

  const onConfirm = () => {}

  return (<>
    <ConfirmationModal isOpen={isOpenModal} onClose={onCloseModal} onConfirm={onConfirm} title='Add Workspace'>
    <h3 className='headline'>DETAILS</h3>
      <div>
        <input type="text" value={workspaceDataModel?.name ?? ""} className={classes.textField + " workspace__name"} placeholder={"Workspace Name"} 
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
