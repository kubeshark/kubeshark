import "../UserSettings/UserSettings.sass"
import {ColsType, FilterableTableAction} from "../UI/FilterableTableAction"
import Api from "../../helpers/api"
import { useEffect, useState } from "react";
import AddWorkspaceModal, { WorkspaceData } from "../Modals/AddWorkspaceModal/AddWorkspaceModal";
import { toast } from "react-toastify";
import ConfirmationModal from "../UI/Modals/ConfirmationModal";

interface Props {}

const api = Api.getInstance();

export const WorkspaceSettings : React.FC<Props> = ({}) => {

    const [workspacesRows, setWorkspacesRows] = useState([]);
    const cols : ColsType[] = [{field : "name",header:"Name"}];

    const [workspaceData,SetWorkspaceData] = useState({} as WorkspaceData);
    const [isOpenModal,setIsOpen] = useState(false);
    const [isEditMode,setIsEditMode] = useState(false);
    const [isOpenDeleteModal, setIsOpenDeleteModal] = useState(false);    

    const buttonConfig = {onClick: () => {setIsOpen(true); setIsEditMode(false);SetWorkspaceData({} as WorkspaceData)}, text:"Add Workspace"}

    useEffect(() => {
        (async () => {
            try {
                const workspaces = await api.getWorkspaces();
                setWorkspacesRows(workspaces)                 
            } catch (e) {
                console.error(e);
            }
        })();
    },[isOpenModal])

    const filterFuncFactory = (searchQuery: string) => {
        return (row) => {
            return row.name.toLowerCase().includes(searchQuery.toLowerCase());
        }
    }

    const searchConfig = { searchPlaceholder: "Search Workspace",filterRows: filterFuncFactory};
    
    const onRowDelete = async (workspace) => {
        setIsOpenDeleteModal(true); 
        SetWorkspaceData(workspace);  
    }
    
    const onDeleteConfirmation = () => {
        (async() => {
            try{
                const workspaceLeft = workspacesRows.filter(ws => ws.id !== workspaceData.id);
                setWorkspacesRows(workspaceLeft);
                await api.deleteWorkspace(workspaceData.id);
                setIsOpenDeleteModal(false);
                SetWorkspaceData({} as WorkspaceData);
                toast.success("Workspace Succesesfully Deleted ");
            } catch {
                toast.error("Workspace hasn't deleted");
            }
        })();
    }

    const onRowEdit = (row) => {
       setIsOpen(true);
       setIsEditMode(true);
       SetWorkspaceData(row);
    }
   
    return (<>
        <FilterableTableAction onRowEdit={onRowEdit} onRowDelete={onRowDelete} searchConfig={searchConfig} 
                               buttonConfig={buttonConfig} rows={workspacesRows} cols={cols} bodyClass="table-body-style">
        </FilterableTableAction>
        <AddWorkspaceModal isOpen={isOpenModal} workspaceId={workspaceData.id} onEdit={isEditMode} onCloseModal={() => { setIsOpen(false);} } >            
        </AddWorkspaceModal>
        <ConfirmationModal isOpen={isOpenDeleteModal} onClose={() => setIsOpenDeleteModal(false)} 
                           onConfirm={onDeleteConfirmation} confirmButtonText="Delete Workspace" title="Delete Workspace"
                           confirmButtonColor="#DB2156">
            <p>Are you sure you want to delete this workspace?</p>
        </ConfirmationModal>
    </>);
}

