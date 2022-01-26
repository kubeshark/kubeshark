import "../UserSettings/UserSettings.sass"
import {ColsType, FilterableTableAction} from "../UI/FilterableTableAction"
// import Api from "../../helpers/api"
import { useEffect, useState } from "react";
import AddWorkspaceModal from "../Modals/AddWorkspaceModal/AddWorkspaceModal";
import SelectList from "../UI/SelectList";

interface Props {}

// const api = Api.getInstance();

export const WorkspaceSettings : React.FC<Props> = ({}) => {

    const [workspacesRows, setWorkspaces] = useState([]);
    const [isOpenModal,setIsOpen] = useState(false);
    const cols : ColsType[] = [{field : "id",header:"Id"},{field : "name",header:"Name"}];

    const namespaces = {"default": false, "blabla": false, "test":true};


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
        <AddWorkspaceModal isOpen={isOpenModal} onCloseModal={() => { setIsOpen(false); } }>
            <SelectList valuesListInput={namespaces} tableName={"Namespaces"} multiSelect={false} setValues={function (newValues: any): void {
                throw new Error("Function not implemented.");
            } } tabelClassName={undefined}></SelectList>
            
        </AddWorkspaceModal>
    </>);
}

