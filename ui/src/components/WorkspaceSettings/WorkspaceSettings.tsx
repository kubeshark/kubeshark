import "./UserSettings"
import {useCommonStyles} from "../../helpers/commonStyle";
import {ColsType, FilterableTableAction} from "../UI/FilterableTableAction"
import Api from "../../helpers/api"
import { useEffect, useState } from "react";

interface Props {}

const api = Api.getInstance();

export const WorkspaceSettings : React.FC<Props> = ({}) => {

    const [workspaces, setWorkspaces] = useState([]);
    const cols : ColsType[] = [{field : "userName",header:"User"},{field : "role",header:"Role"}, {field : "role",header:"Role"}]


    useEffect(() => {
        (async () => {
            try {
                const workspaces = await api.getUsers() 
                setWorkspaces(workspaces)                  
            } catch (e) {
                console.error(e);
            }
        })();
    },[])

    const filterFuncFactory = (searchQuery: string) => {
        return (row) => {
            return row.userName.toLowercase().includes(searchQuery.toLowerCase()) > -1
        }
    }

    const searchConfig = { searchPlaceholder: "Search Workspace",filterRows: filterFuncFactory}
    
    const onRowDelete = (row) => {
        const filterFunc = filterFuncFactory(row.userName)
        const newUserList = workspaces.filter(filterFunc)
        setWorkspaces(newUserList)
    }

    const onRowEdit = (row) => {

    }

    const buttonConfig = {onClick: () => {}, text:"Add User"}
    return (<>
        <FilterableTableAction onRowEdit={onRowEdit} onRowDelete={onRowDelete} searchConfig={searchConfig} 
                               buttonConfig={buttonConfig} rows={workspaces} cols={cols}>
        </FilterableTableAction>
    </>);
}
