import { useEffect, useMemo, useState } from "react";
import Checkbox from "./Checkbox"
import Radio from "./Radio";
import './style/SelectList.sass';

export interface Props {
    valuesListInput;
    tableName:string;
    checkedValues?:string[];
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

const SelectList: React.FC<Props> = ({valuesListInput ,tableName,checkedValues=[],multiSelect=true,searchValue="",setValues,tabelClassName}) => {
    const [valuesList, setValuesList] = useState(valuesListInput as ValuesListInput);

    useEffect(() => {   
            const list = valuesList.map(obj => {
                const isValueChecked = checkedValues.some(checkedValueKey => obj.key === checkedValueKey)
                return {...obj, isChecked: isValueChecked}
            })
        setValuesList(list);
    },[valuesListInput,checkedValues]);

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
        const isChecked = valuesList.every(valueTap => valueTap.isChecked === true);
        const list = valuesList.map((obj) => {
            return {...obj, isChecked: !isChecked}
        })
        setValuesList(list);
        setValues(list);
    }


    const tableHead = multiSelect ? <tr style={{borderBottomWidth: "2px"}}>
            <th style={{width: 50}}><Checkbox checked={valuesList.every(valueTap => valueTap.isChecked === true)}
                onToggle={toggleAll}/></th>
            <th>{tableName}</th>
        </tr> : 
        <tr style={{borderBottomWidth: "2px"}}>
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