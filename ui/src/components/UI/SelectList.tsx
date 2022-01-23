import { useMemo, useState } from "react";
import Checkbox from "./Checkbox"

export interface Props {
    valuesListInput;
    tableName:string;
    multiSelect:boolean;
    searchValue?:string;
    setValues: (newValues)=> void;
    tabelClassName
}

const SelectList: React.FC<Props> = ({valuesListInput,tableName,multiSelect=true,searchValue="",setValues,tabelClassName}) => {
    const [valuesList, setValuesList] = useState(valuesListInput);

    const toggleValues = (value) => {
        if (!multiSelect){
            unToggleAll(value);
        }
        else {
            const newValues = {...valuesList};
            newValues[value] = !valuesList[value];
            setValuesList(newValues);
            setValues(newValues);
        }
    }

    const unToggleAll = (value) => {
        const newValues = {};
        Object.keys(valuesList).forEach(key => {
            if (key !== value)
                newValues[key] = false;
            else
                newValues[key] = true;
        })
        setValuesList(newValues);
        setValues(newValues);
    }

    const toggleAll = () => {
        const isChecked = Object.values(valuesList).every(tap => tap === true);
        setAllCheckedValue(!isChecked);
    }

    const setAllCheckedValue = (isTap: boolean) => {
        const newValues = {};
        Object.keys(valuesList).forEach(key => {
            newValues[key] = isTap;
        })
        setValuesList(newValues);
        setValues(newValues);
    }

    const tableHead = multiSelect ? <tr style={{borderBottomWidth: "2px"}}>
            <th style={{width: 50}}><Checkbox checked={Object.values(valuesList).every(tap => tap === true)}
                onToggle={toggleAll}/></th>
            <th>{tableName}</th>
        </tr> : 
        <tr style={{borderBottomWidth: "2px"}}>
            <th>{tableName}</th>
        </tr>

    const filteredValues = useMemo(() => {
        return Object.keys(valuesList).filter((listValue) => listValue.includes(searchValue));
    },[valuesList, searchValue])

        return <div className={tabelClassName}>
                <table cellPadding={5} style={{borderCollapse: "collapse"}}>
                    <thead>
                    {tableHead}
                    </thead>
                    <tbody>
                    {filteredValues?.map(listValue => {
                            return <tr key={listValue}>
                                <td style={{width: 50}}>
                                    <Checkbox checked={valuesList[listValue]} onToggle={() => toggleValues(listValue)}/>
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