import React, { useCallback, useEffect, useMemo, useState } from "react";
import Radio from "./Radio";
import styles from './style/SelectList.module.sass'
import NoDataMessage from "./NoDataMessage";
import Checkbox from "./Checkbox";


export interface Props {
    items;
    tableName: string;
    checkedValues?: string[];
    multiSelect: boolean;
    searchValue?: string;
    setCheckedValues: (newValues) => void;
    tableClassName?
    checkBoxWidth?: string
}

const SelectList: React.FC<Props> = ({ items, tableName, checkedValues = [], multiSelect = true, searchValue = "", setCheckedValues, tableClassName,
    checkBoxWidth = 50 }) => {
    const noItemsMessage = "No items to show";
    const [headerChecked, setHeaderChecked] = useState(false)

    const filteredValues = useMemo(() => {
        return items.filter((listValue) => listValue?.value?.includes(searchValue));
    }, [items, searchValue])

    const filteredValuesKeys = useMemo(() => {
        return filteredValues.map(x => x.key)
    }, [filteredValues])

    const toggleValue = (checkedKey) => {
        if (!multiSelect) {
            const newCheckedValues = [];
            newCheckedValues.push(checkedKey);
            setCheckedValues(newCheckedValues);
        }
        else {
            const newCheckedValues = [...checkedValues];
            let index = newCheckedValues.indexOf(checkedKey);

            if (index > -1)
                newCheckedValues.splice(index, 1);
            else
                newCheckedValues.push(checkedKey);

            setCheckedValues(newCheckedValues);
        }
    }

    useEffect(() => {
        const setAllChecked = filteredValuesKeys.every(val => checkedValues.includes(val))
        setHeaderChecked(setAllChecked)
    }, [filteredValuesKeys, checkedValues])

    const toggleAll = useCallback((shouldCheckAll) => {
        let newChecked = checkedValues.filter(x => !filteredValuesKeys.includes(x))

        if (shouldCheckAll) {
            const disabledItems = items.filter(i => i.disabled).map(x => x.key)
            newChecked = [...filteredValuesKeys, ...newChecked].filter(x => !disabledItems.includes(x))
        }

        setCheckedValues(newChecked)
    }, [searchValue, checkedValues, filteredValuesKeys])

    const dataFieldFunc = (listValue) => listValue.component ? listValue.component :
        <span className={styles.nowrap} title={listValue.value}>
            {listValue.value}
        </span>

    const tableHead = multiSelect ? <tr style={{ borderBottomWidth: "2px" }}>
        <th style={{ width: checkBoxWidth }}><Checkbox data-cy="checkbox-all" checked={headerChecked}
            onToggle={(isChecked) => toggleAll(isChecked)} /></th>
        <th>{tableName}</th>
    </tr> :
        <tr style={{ borderBottomWidth: "2px" }}>
            <th>{tableName}</th>
        </tr>

    const tableBody = filteredValues.length === 0 ?
        <tr>
            <td colSpan={2}>
                <NoDataMessage messageText={noItemsMessage} />
            </td>
        </tr>
        :
        filteredValues?.map(listValue => {
            return <tr key={listValue.key}>
                <td style={{ width: checkBoxWidth }}>
                    {multiSelect && <Checkbox data-cy={"checkbox-" + listValue.value} disabled={listValue.disabled} checked={checkedValues.includes(listValue.key)} onToggle={() => toggleValue(listValue.key)} />}
                    {!multiSelect && <Radio data-cy={"radio-" + listValue.value} disabled={listValue.disabled} checked={checkedValues.includes(listValue.key)} onToggle={() => toggleValue(listValue.key)} />}
                </td>
                <td>
                    {dataFieldFunc(listValue)}
                </td>
            </tr>
        }
        )

    return <div className={tableClassName ? tableClassName + ` ${styles.selectListTable}` : ` ${styles.selectListTable}`}>
        <table cellPadding={5} style={{ borderCollapse: "collapse" }}>
            <thead>
                {tableHead}
            </thead>
            <tbody>
                {tableBody}
            </tbody>
        </table>
    </div>
}

export default SelectList;