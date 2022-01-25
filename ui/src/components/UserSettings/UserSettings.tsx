import "./UserSettings"
import {useCommonStyles} from "../../helpers/commonStyle";
import {ColsType, FilterableTableAction} from "../UI/FilterableTableAction"
import Api from "../../helpers/api"
import { useEffect, useState } from "react";

interface Props {

}

const api = Api.getInstance();

export const UserSettings : React.FC<Props> = ({}) => {

    const [usersRows, setUserRows] = useState([]);
    const cols : ColsType[] = [{field : "userName",header:"User"},{field : "role",header:"Role"}, {field : "role",header:"Role"}]


    useEffect(() => {
        (async () => {
            try {
                const users = await api.getUsers() 
                setUserRows(usersRows)                  
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

    const searchConfig = { searchPlaceholder: "Search User",filterRows: filterFuncFactory}
    
    const onRowDelete = (row) => {
        const filterFunc = filterFuncFactory(row.userName)
        const newUserList = usersRows.filter(filterFunc)
        setUserRows(newUserList)
    }

    const onRowEdit = (row) => {

    }

    const buttonConfig = {onClick: () => {}, text:"Add User"}
    return (<>
        <FilterableTableAction onRowEdit={onRowEdit} onRowDelete={onRowDelete} searchConfig={searchConfig} buttonConfig={buttonConfig} rows={usersRows} cols={cols}>
        </FilterableTableAction>
    </>);
}
