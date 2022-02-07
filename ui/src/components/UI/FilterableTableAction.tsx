import { Button } from "@material-ui/core";
import React, { useEffect, useMemo, useState } from "react";
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
    bodyClass?: string;
}

export const FilterableTableAction: React.FC<Props> = ({onRowDelete,onRowEdit, searchConfig, buttonConfig, rows, cols, bodyClass}) => {

    const classes = useCommonStyles()

    const [tableRows,setRows] = useState(rows as any[])
    const [inputSearch, setInputSearch] = useState("")

    useEffect(() => {
        setRows(rows);
    },[rows])

    const onChange = (e) => {
        setInputSearch(e.target.value)
    }

    const filteredValues = useMemo(() => {
        const searchFunc = searchConfig.filterRows(inputSearch)
        return tableRows.filter(searchFunc)
    },[tableRows, inputSearch,searchConfig])

    return (<>
        <div className="filterable-table">
            <div className="actions-header">
                <input type="text" className={classes.textField + " actions-header__search-box"} placeholder={searchConfig.searchPlaceholder} onChange={onChange}></input>
                <Button style={{height: '100%'}} className={classes.button + " actions-header__action-button"} size={"small"} onClick={buttonConfig.onClick}>
                            {buttonConfig.text} 
                </Button>
            </div>
            <Table rows={filteredValues} cols={cols} onRowEdit={onRowEdit} onRowDelete={onRowDelete} bodyClass={bodyClass}></Table>
        </div>
    </>);
};