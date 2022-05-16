import React, { useCallback, useEffect, useMemo, useState } from "react";
import Radio from "./Radio";
import styles from './style/SelectList.module.sass'
import NoDataMessage from "./NoDataMessage";
import Checkbox from "./Checkbox";
import { useCommonStyles } from "../../helpers/commonStyle";


export interface Props {
    items;
    tableName: string;
    checkedValues?: string[];
    multiSelect: boolean;
    setCheckedValues: (newValues) => void;
    tableClassName?;
    checkBoxWidth?: string;
    inputSearchClass? : string
    isFilterable? : boolean
}

const SelectList: React.FC<Props> = ({ items, tableName, checkedValues = [], multiSelect = true, setCheckedValues, tableClassName,
    checkBoxWidth = 50 ,inputSearchClass,isFilterable = true}) => {
    const commonClasses = useCommonStyles();
    const [searchValue, setSearchValue] = useState("")  
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
        <th>
            All
        </th>
    </tr> :
        <tr>
        </tr>

    const tableBody = filteredValues.length === 0 ?
        <tr>
            <td colSpan={2} className={styles.displayBlock}>
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

    return <React.Fragment>
        <h3 className={styles.subSectionHeader}>
            {tableName}
            <span className={styles.totalSelected}>&nbsp;({checkedValues.length})</span>
        </h3>
        {isFilterable && <input className={commonClasses.textField + ` ${inputSearchClass}`} placeholder="Search" value={searchValue}
                                onChange={(event) => setSearchValue(event.target.value)} data-cy="searchInput" />}
        <div className={tableClassName ? tableClassName + ` ${styles.selectListTable}` : ` ${styles.selectListTable}`} style={{marginTop: !multiSelect ? "20px":  ""}}>
        <table cellPadding={5} style={{ borderCollapse: "collapse" }}>
            <thead>
                {tableHead}
            </thead>
            <tbody>
                {tableBody}
            </tbody>
        </table>
    </div>
    </React.Fragment>
}

export default SelectList;
