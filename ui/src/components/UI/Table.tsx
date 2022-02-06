import React, {useEffect, useState} from "react";
import './style/Table.sass';
import NoDataMessage from "./NoDataMessage";

export interface ColsType {
    field:string,
    cellClassName?: string,
    header:string,
    width?:string,
    getCellClassName?:(field:string,value : any) => string
    mapValue? : (val : any) => any
};

interface TableProps {
    rows : any[];
    cols : ColsType[]
    onRowEdit : (any) => void;
    onRowDelete : (any) => void;
    noDataMeesage? : string;
    bodyClass?: string;
}

export const Table: React.FC<TableProps> = ({rows, cols, onRowDelete, onRowEdit, noDataMeesage = "No Data Found",bodyClass}) => {

    const [tableRows, updateTableRows] = useState(rows);

    useEffect(() => {
        updateTableRows(rows)
    },[rows])

    const _onRowEdit = (row) => {
        onRowEdit(row)
    }
    
    const _onRowDelete = (row) => {
        onRowDelete(row);
    }

    return <table cellPadding={5} style={{borderCollapse: "collapse"}} className="mui-table">
    <thead className="mui-table__thead">
    <tr style={{borderBottomWidth: "2px"}} className="mui-table__tr">
        {cols?.map((col)=> {
            return <th key={col.header} className="mui-table__th" style={{width: col.width}}>{col.header}</th>
        })}
        <th></th>
    </tr>
    </thead>
    <tbody className={`mui-table__tbody ${bodyClass}`}>
        {
            ((tableRows == null) || (tableRows.length === 0)) ?
            <tr className="mui-table__tr--no-data">
                <td >
                    <NoDataMessage messageText={noDataMeesage}/>
                </td>
            </tr>
            :
            tableRows?.map((rowData,index) => {
                return <tr key={rowData?.id ?? index} className="mui-table__tr">
                    {cols.map((col,index) => {                        
                        return <td key={`${rowData?.id} + ${index}`} className="mui-table__td" style={{width: col.width}}>
                                 <span key={Math.random().toString()} className={`${col.getCellClassName ? col.getCellClassName(col.field, rowData[col.field]) : ""}${col?.cellClassName ?? ""}`}>
                                     {col.mapValue ? col.mapValue(rowData[col.field]) : rowData[col.field]}
                                </span>
                            </td>
                    })}
                    <td className="mui-table__td mui-table__row-actions">        
                        <span onClick={() => _onRowEdit(rowData)} className="mui-table__action mui-table__action--edit"></span>
                        <span  onClick={() => _onRowDelete(rowData)} className="mui-table__action mui-table__action--delete"></span>
                    </td>
                </tr>
            })
        }
    </tbody>
</table>
}