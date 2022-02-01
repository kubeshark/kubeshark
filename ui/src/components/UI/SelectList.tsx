import { useEffect, useMemo } from "react";
import Checkbox from "./Checkbox"
import Radio from "./Radio";
import './style/SelectList.sass';

export type ValuesListInput = {
    key: string;
    value: string;
}[]
export interface Props {
    items;
    tableName:string;
    checkedValues?:string[];
    multiSelect:boolean;
    searchValue?:string;
    setCheckedValues: (newValues)=> void;
    tabelClassName
}

const SelectList: React.FC<Props> = ({items ,tableName,checkedValues=[],multiSelect=true,searchValue="",setCheckedValues,tabelClassName}) => {
 
    const filteredValues = useMemo(() => {
        return items.filter((listValue) => listValue?.value?.includes(searchValue));
    },[items, searchValue])

    const toggleValue = (checkedKey) => {
        if (!multiSelect){
            // unToggleAll();
            const newCheckedValues = [];
            newCheckedValues.push(checkedKey);
            setCheckedValues(newCheckedValues);
        }
        else {
            const newCheckedValues = [...checkedValues];
            let index = newCheckedValues.indexOf(checkedKey);
            if(index > -1) 
                newCheckedValues.splice(index,1);
            else 
                newCheckedValues.push(checkedKey);   
            setCheckedValues(newCheckedValues);
        }
    }

    const unToggleAll = () => {
        setCheckedValues([]);
    }

    const toggleAll = () => {
        const newCheckedValues = [...checkedValues];
        if(newCheckedValues.length === items.length) setCheckedValues([]);
        else {
            items.forEach((obj) => {
            if(!newCheckedValues.includes(obj.key))
                newCheckedValues.push(obj.key);
            })
            setCheckedValues(newCheckedValues);
        }
    }

    const tableHead = multiSelect ? <tr style={{borderBottomWidth: "2px"}}>
            <th style={{width: 50}}><Checkbox checked={items.length === checkedValues.length}
                onToggle={toggleAll}/></th>
            <th>{tableName}</th>
        </tr> : 
        <tr style={{borderBottomWidth: "2px"}}>
            <th>{tableName}</th>    
        </tr>

        return <div className={tabelClassName + " select-list-table"}>
                <table cellPadding={5} style={{borderCollapse: "collapse"}}>
                    <thead>
                    {tableHead}
                    </thead>
                    <tbody>
                    {filteredValues?.map(listValue => {
                            return <tr key={listValue.key}>
                                <td style={{width: 50}}>
                                    {multiSelect && <Checkbox checked={checkedValues.includes(listValue.key)} onToggle={() => toggleValue(listValue.key)}/>}
                                    {!multiSelect && <Radio checked={checkedValues.includes(listValue.key)} onToggle={() => toggleValue(listValue.key)}/>}
                                </td>
                                <td>{listValue.value}</td>
                            </tr>   
                        }
                    )}
                    </tbody>
            </table>
        </div>
}

export default SelectList;