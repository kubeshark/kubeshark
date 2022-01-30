import { useMemo, useState } from "react";
import Checkbox from "./Checkbox"
import Radio from "./Radio";
import './style/SelectList.sass';

export interface Props {
    valuesListInput;
    tableName:string;
    multiSelect:boolean;
    searchValue?:string;
    setValues: (newValues)=> void;
    tabelClassName
}

export type ValuesListInput = {
    key: string;
    value: string;
    isChecked: boolean;
}[]

const SelectList: React.FC<Props> = ({valuesListInput ,tableName,multiSelect=true,searchValue="",setValues,tabelClassName}) => {
    const [valuesList, setValuesList] = useState(valuesListInput as ValuesListInput);

    const toggleValues = (checkedKey) => {
        if (!multiSelect){
            unToggleAll(checkedKey);
        }
        else {
            const newValues: ValuesListInput = [...valuesList];
            newValues.map(item => item.key === checkedKey ? item.isChecked = !item.isChecked : item.isChecked);
            setValuesList(newValues);
            setValues(newValues);
        }
    }

    const unToggleAll = (checkedKey) => {
        const list = valuesList.map((obj) => {
            return {...obj, isChecked:checkedKey === obj.key}
        })
        setValuesList(list);
        setValues(list);
    }


    const toggleAll = () => {
        const list = valuesList.map((obj) => {
            return {...obj, isChecked: true}
        })
        setValuesList(list);
        setValues(list);
    }


    const tableHead = multiSelect ? 
        <tr style={{borderBottomWidth: "2px"}}>
            <th style={{width: 50}}><Checkbox checked={valuesList.every(valueTap => valueTap.isChecked === false)}
                onToggle={toggleAll}/></th>
            <th>{tableName}</th>
        </tr> : 
         <tr style={{borderBottomWidth: "2px",display: !tableName ? "none" : "table"}}>
            <th>{tableName}</th>    
        </tr>

    const filteredValues = useMemo(() => {
        return valuesList.filter((listValue) => listValue?.value?.includes(searchValue));
    },[valuesList, searchValue])

        return <div className={tabelClassName + " select-list-table"}>
                <table cellPadding={5} style={{borderCollapse: "collapse"}}>
                    <thead>
                    {tableHead}
                    </thead>
                    <tbody>
                    {filteredValues?.map(listValue => {
                            return <tr key={listValue.key}>
                                <td style={{width: 50}}>
                                    {multiSelect && <Checkbox checked={valuesList.find(item => item.key === listValue.key)?.isChecked} onToggle={() => toggleValues(listValue.key)}/>}
                                    {!multiSelect && <Radio checked={valuesList.find(item => item.key === listValue.key)?.isChecked} onToggle={() => toggleValues(listValue.key)}/>}
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