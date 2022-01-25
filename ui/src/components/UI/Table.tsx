import React, {useEffect, useState} from "react";
import './style/Table.sass';
import Delete from '@material-ui/icons/Delete';
import Edit from '@material-ui/icons/Edit';

export interface ColsType {
    field:string,
    cellClassName?: string,
    header:string,
    width?:number,
    getCellClassName?:(field:string,value : any) => string
};

interface TableProps {
    rows : any[];
    cols : ColsType[]
    onRowEdit : (any) => void;
    onRowDelete : (any) => void;
    noDataMeesage? : string;
}

export const Table: React.FC<TableProps> = ({rows, cols, onRowDelete, onRowEdit, noDataMeesage = "No Data Found"}) => {

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

    // const filteredValues = useMemo(() => {
    //     return tableRows.filter((listValue) => listValue.find(row));
    // },[tableRows, searchValue])

    return <table cellPadding={5} style={{borderCollapse: "collapse"}} className="mui-table">
    <thead className="mui-table__thead">
    <tr style={{borderBottomWidth: "2px"}} className="mui-table__tr">
        {cols?.map((col)=> {
            return <th key={col.header} className="mui-table__th">{col.header}</th>
        })}
        <th></th>
    </tr>
    </thead>
    <tbody className="mui-table__tbody">
        {
            ((tableRows == null) || (tableRows.length === 0)) ?
            <tr className="mui-table__no-data">
                {/* <img src={circleImg} alt="No data Found"></img>
                <div className="mui-table__no-data-message">{noDataMeesage}</div> */}
            </tr>

            :
        
            tableRows?.map((rowData,index) => {
                return <tr key={rowData?.id ?? index} className="mui-table__tr">
                    {cols.map((col,index) => {                        
                        return <td key={`${rowData?.id} + ${index}`} className="mui-table__td">
                                 <span key={Math.random().toString()} className={`${col.getCellClassName ? col.getCellClassName(col.field, rowData[col.field]) : ""}${col?.cellClassName ?? ""}`}>
                                     {rowData[col.field]}
                                </span>
                            </td>
                    })}
                    <td className="mui-table__td mui-table__row-actions">        
                        <span onClick={() => _onRowEdit(rowData)}>
                            <Edit className="mui-table__row-actions--edit"></Edit>
                        </span>
                        <span  onClick={() => _onRowDelete(rowData)}>
                            <Delete className="mui-table__row-actions--delete"></Delete>
                        </span>
                    </td>
                </tr>
            })
        }
    </tbody>
</table>
}