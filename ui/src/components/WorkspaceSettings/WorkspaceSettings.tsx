import "../UserSettings/UserSettings.sass"
import {ColsType, FilterableTableAction} from "../UI/FilterableTableAction"
import Api from "../../helpers/api"
import { useEffect, useState } from "react";
import AddWorkspaceModal, { WorkspaceData } from "../Modals/AddWorkspaceModal/AddWorkspaceModal";
import { toast } from "react-toastify";
import ConfirmationModal from "../UI/Modals/ConfirmationModal";
import spinner from "../assets/spinner.svg";


const api = Api.getInstance();

export const WorkspaceSettings : React.FC = () => {

    const [workspacesRows, setWorkspacesRows] = useState([]);
    const cols : ColsType[] = [{field : "name",header:"Name"}];

    const [workspaceData,setWorkspaceData] = useState({} as WorkspaceData);
    const [isOpenModal,setIsOpen] = useState(false);
    const [isEditMode,setIsEditMode] = useState(false);
    const [isOpenDeleteModal, setIsOpenDeleteModal] = useState(false);  
    const [isLoading, setIsLoading] = useState(false);  

    const buttonConfig = {onClick: () => {setIsOpen(true); setIsEditMode(false);setWorkspaceData({} as WorkspaceData)}, text:"Add Workspace"}

    const getWorkspaces = async() => {
        try {
            setIsLoading(true);
            const workspaces = await api.getWorkspaces();
            setWorkspacesRows(workspaces)                 
        } catch (e) {
            console.error(e);
        } finally {
            setIsLoading(false);
        }
    }

    useEffect(() => {
        getWorkspaces();
    },[])

    const onWorkspaceAdded = ()=> getWorkspaces();

    const filterFuncFactory = (searchQuery: string) => {
        return (row) => {
            return row.name.toLowerCase().includes(searchQuery.toLowerCase());
        }
    }

    const searchConfig = { searchPlaceholder: "Search Workspace",filterRows: filterFuncFactory};
    
    const onRowDelete = async (workspace) => {
        setIsOpenDeleteModal(true); 
        setWorkspaceData(workspace);  
    }
    
    const onDeleteConfirmation = async() => {
            try{
                const workspaceLeft = workspacesRows.filter(ws => ws.id !== workspaceData.id);
                setWorkspacesRows(workspaceLeft);
                await api.deleteWorkspace(workspaceData.id);
                setIsOpenDeleteModal(false);
                setWorkspaceData({} as WorkspaceData);
                toast.success("Workspace Succesesfully Deleted ");
            } catch(e) {
                console.error(e);
                toast.error("Workspace hasn't deleted");
            }
    }

    const onRowEdit = (row) => {
       setIsOpen(true);
       setIsEditMode(true);
       setWorkspaceData(row);
    }
   
    return (<>
         {isLoading? <div style={{textAlign: "center", padding: 20}}>
                        <img alt="spinner" src={spinner} style={{height: 35}}/>
                    </div> : <>
                    <FilterableTableAction onRowEdit={onRowEdit} onRowDelete={onRowDelete} searchConfig={searchConfig} 
                               buttonConfig={buttonConfig} rows={workspacesRows} cols={cols} bodyClass="table-body-style"/></>}
        <AddWorkspaceModal isOpen={isOpenModal} workspaceId={workspaceData.id} onEdit={isEditMode} onWorkspaceAdded={onWorkspaceAdded} onCloseModal={() => { setIsOpen(false);} } >            
        </AddWorkspaceModal>
        <ConfirmationModal isOpen={isOpenDeleteModal} onClose={() => setIsOpenDeleteModal(false)} 
                           onConfirm={onDeleteConfirmation} confirmButtonText="Delete Workspace" title="Delete Workspace"
                           confirmButtonColor="#DB2156" className={"delete-comfirmation-modal"}>
            <p>Are you sure you want to delete this workspace?</p>
        </ConfirmationModal>
    </>);
}

