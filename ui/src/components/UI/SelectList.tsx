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
            newValues[checkedKey].isChecked = !valuesList[checkedKey].isChecked;
            setValuesList(newValues);
            setValues(newValues);
        }
    }

    const unToggleAll = (checkedKey) => {
        const newValues = [];
        valuesList.forEach(valueList => {
            if (valueList.key !== checkedKey)
                newValues[checkedKey] = false;
            else
                newValues[checkedKey] = true;
        })
        setValuesList(newValues);
        setValues(newValues);
    }

    const toggleAll = () => {
        const isChecked = valuesList.every(valueTap => valueTap.isChecked === true);
        setAllCheckedValue(!isChecked);
    }

    const setAllCheckedValue = (isTap: boolean) => {
        const newValues = [];
        valuesList.forEach(valueList => {
            newValues[valueList.key] = isTap;
        })
        setValuesList(newValues);
        setValues(newValues);
    }

    const tableHead = multiSelect ? <tr style={{borderBottomWidth: "2px"}}>
            <th style={{width: 50}}><Checkbox checked={valuesList.every(valueTap => valueTap.isChecked === false)}
                onToggle={toggleAll}/></th>
            <th>{tableName}</th>
        </tr> : 
        <tr style={{borderBottomWidth: "2px"}}>
            <th>{tableName}</th>    
        </tr>

    const filteredValues = useMemo(() => {
        return valuesList.filter((listValue) => listValue.value.includes(searchValue));
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
                                    {multiSelect && <Checkbox checked={valuesList[listValue.key].isChecked} onToggle={() => toggleValues(listValue.key)}/>}
                                    {!multiSelect && <Radio checked={valuesList[listValue.key].isChecked} onToggle={() => toggleValues(listValue.key)}/>}
                                </td>
                                <td>{listValue}</td>
                            </tr>   
                        }
                    )}
                    </tbody>
            </table>
        </div>
}

export default SelectList;