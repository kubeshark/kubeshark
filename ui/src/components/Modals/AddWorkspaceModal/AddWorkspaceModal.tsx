import { FC, useEffect, useMemo, useState } from 'react';
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
  workspaceDataInput: WorkspaceData,
  onEdit: boolean
}

const api = Api.getInstance();

const AddWorkspaceModal: FC<AddWorkspaceModalProp> = ({isOpen,onCloseModal, workspaceDataInput ={}, onEdit}) => {

  const [workspaceDataModal, setWorkspaceData] = useState({} as WorkspaceData);
  const [searchValue, setSearchValue] = useState("");
  const [namespaces, setNamespaces] = useState([]);
  const [namespacesNames, setNamespaceNames] = useState([]);
  const [workspaceName, setWorkspaceName] = useState("");
  const [disable, setDisable] = useState(true);

  console.log(workspaceDataInput);
  console.log(workspaceDataModal);

  const classes = useCommonStyles();

  const title = onEdit ? "Edit Workspace" : "Add Workspace";

  useEffect(() => {
    setWorkspaceData(workspaceDataInput as WorkspaceData);
  },[workspaceDataInput])

  useEffect(() => {
    if(!isOpen) return;
    (async () => {
        try {
            setSearchValue("");     
            const namespaces = [{key:"1", value:"namespace1",isChecked:false},{key:"2", value:"namespace2",isChecked:false}];
            const list = namespaces.map(obj => {
            const isValueChecked = workspaceDataModal.namespaces.some(checkedValueKey => obj.key === checkedValueKey)
            return {...obj, isChecked: isValueChecked}
          })
            setNamespaces(namespaces);
    } catch (e) {
            console.error(e);
        } finally {
        }
    })()
}, [isOpen])

  const onWorkspaceNameChange = (event) => {
    // const data = {...workspaceData, name: event.target.value};
    //setWorkspaceData(data);
    setWorkspaceName(event.target.value);
    setGenarateDisabledState();
  }


  const onNamespaceChange = (newVal) => {
    var filteredValues = newVal.filter(obj => obj.isChecked);
    var namespaceNames = filteredValues.map(obj => obj.value);
    setNamespaceNames(namespaceNames);

    // var filteredValues = newVal.filter(obj => obj.isChecked);
    // var namespaceNames = filteredValues.map(obj => obj.value);
    // const data = {...workspaceData, namespaces: namespaceNames};
    // setWorkspaceData(data);
    setGenarateDisabledState();
  }

  const isFormValid = () : boolean => {
    return (Object.values(workspaceDataModal).length === 2) && Object.values(workspaceDataModal).every(val => val !== null)
  }

  const setGenarateDisabledState = () => {
    const isValid = isFormValid()
    setDisable(!isValid)
  }

  const onConfirm = () => {
    const data = {name: workspaceName, namespaces: namespacesNames};
    setWorkspaceData(data);
  }

  const onClose = () => {
    onCloseModal();
    setWorkspaceData({} as WorkspaceData);
  }

  return (<>
    <ConfirmationModal isOpen={isOpen} onClose={onClose} onConfirm={onConfirm} title={title}>
    <h3 className='headline'>DETAILS</h3>
      <div>
        <input type="text" value={workspaceDataModal?.name ?? ""} className={classes.textField + " workspace__name"} placeholder={"Workspace Name"} 
               onChange={onWorkspaceNameChange}></input>
        </div>
        <h3 className='headline'>TAP SETTINGS</h3>     
      <div className="namespacesSettingsContainer">
        <div>
            <input className={classes.textField + " searchNamespace"} placeholder="Search" value={searchValue}
                    onChange={(event) => setSearchValue(event.target.value)}/>
        </div>
        <SelectList valuesListInput={namespaces}
                    tableName={"Namespaces"}
                    multiSelect={true} 
                    checkedValues={workspaceDataModal.namespaces} 
                    searchValue={searchValue} 
                    setValues={onNamespaceChange} 
                    tabelClassName={undefined}>
                    </SelectList>
            </div>
    </ConfirmationModal>
    </>); 
};

export default AddWorkspaceModal;
