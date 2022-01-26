import "../UserSettings/UserSettings.sass"
import {ColsType, FilterableTableAction} from "../UI/FilterableTableAction"
// import Api from "../../helpers/api"
import { useEffect, useState } from "react";
import AddWorkspaceModal, { WorkspaceData } from "../Modals/AddWorkspaceModal/AddWorkspaceModal";
import SelectList from "../UI/SelectList";

interface Props {}

// const api = Api.getInstance();

export const WorkspaceSettings : React.FC<Props> = ({}) => {

    const [workspacesRows, setWorkspaces] = useState([]);
    const [workspaceData,SetWorkspaceData] = useState({} as WorkspaceData);
    const [isOpenModal,setIsOpen] = useState(false);
    const cols : ColsType[] = [{field : "id",header:"Id"},{field : "name",header:"Name"}];


    useEffect(() => {
        (async () => {
            try {
                const workspacesDemo = [{id:"1", name:"Worksapce1"}] 
                setWorkspaces(workspacesDemo)                  
            } catch (e) {
                console.error(e);
            }
        })();
    },[])

    const filterFuncFactory = (searchQuery: string) => {
            return (row) => row.name.toLowerCase().includes(searchQuery.toLowerCase())
        }

    const searchConfig = { searchPlaceholder: "Search Workspace",filterRows: filterFuncFactory}
    
    const onRowDelete = (row) => {
        const filterFunc = filterFuncFactory(row.name)
        const newWorkspaceList = workspacesRows.filter(filterFunc)
        setWorkspaces(newWorkspaceList)
    }

    const onRowEdit = (row) => {

    }

    const buttonConfig = {onClick: () => {setIsOpen(true)}, text:"Add Workspace"}
    return (<>
        <FilterableTableAction onRowEdit={onRowEdit} onRowDelete={onRowDelete} searchConfig={searchConfig} 
                               buttonConfig={buttonConfig} rows={workspacesRows} cols={cols}>
        </FilterableTableAction>
        <AddWorkspaceModal isOpen={isOpenModal} workspaceData={workspaceData} onCloseModal={() => { setIsOpen(false); } } >

            
        </AddWorkspaceModal>
    </>);
}

