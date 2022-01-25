import { Button } from "@material-ui/core";
import React, { useEffect, useState } from "react";
import { Table } from "./Table";
import {useCommonStyles} from "../../helpers/commonStyle";
import {ColsType} from "../UI/Table"
import './style/FilterableTableAction.sass';


export type {ColsType} from "../UI/Table"

type filterFuncFactory = (query:string) => (any) => boolean
export interface Props {
    onRowEdit : (any) => void;
    onRowDelete : (any) => void;
    searchConfig : {searchPlaceholder : string;filterRows: filterFuncFactory};
    buttonConfig : {onClick : () => void, text:string}
    rows: any[];
    cols: ColsType[];
}

export const FilterableTableAction: React.FC<Props> = ({onRowDelete,onRowEdit, searchConfig, buttonConfig, rows, cols}) => {

    const classes = useCommonStyles()

    const [tableRows,setRows] = useState(rows as any[])
    const [inputSearch, setInputSearch] = useState("")
    let allRows = rows;

    useEffect(() => {
        allRows = rows;
        setRows(rows);
    },[rows])

    useEffect(()=> {  
        if(inputSearch !== ""){
            const searchFunc = searchConfig.filterRows(inputSearch)
            const filtered = tableRows.filter(searchFunc)
            setRows(filtered)
        }
        else{
            setRows(allRows);
        }
    },[inputSearch])

    const onChange = (e) => {
        setInputSearch(e.target.value)
    }

    return (<>
        <div className="filterable-table">
            <div className="actions-header">
                <input type="text" className={classes.textField + " actions-header__search-box"} placeholder={searchConfig.searchPlaceholder} onChange={onChange}></input>
                <Button style={{height: '100%'}} className={classes.button + " actions-header__action-button"} size={"small"} onClick={buttonConfig.onClick}>
                            {buttonConfig.text}
                </Button>
            </div>
            <Table rows={tableRows} cols={cols} onRowEdit={onRowEdit} onRowDelete={onRowDelete}></Table>
        </div>
    </>);
};