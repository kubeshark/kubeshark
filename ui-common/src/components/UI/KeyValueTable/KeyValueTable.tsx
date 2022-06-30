import React from "react";
import { useEffect, useState } from "react";
import styles from "./KeyValueTable.module.sass"
import deleteIcon from "delete.svg"
import deleteIconActive from "delete-active.svg"
import HoverImage from "../HoverImage/HoverImage";

interface KeyValueTableProps {
    data: any
    onDataChange: (data: any) => void
    keyPlaceholder?: string
    valuePlaceholder?: string
}

type Row = { key: string, value: string }

const KeyValueTable: React.FC<KeyValueTableProps> = ({ data, onDataChange, keyPlaceholder, valuePlaceholder }) => {

    const [keyValueData, setKeyValueData] = useState([] as Row[])

    useEffect(() => {
        if (!data) return;
        const currentState = [...data, { key: "", value: "" }]
        setKeyValueData(currentState)
    }, [data])

    const deleteRow = (index) => {
        const newRows = [...keyValueData];
        newRows.splice(index, 1);
        setKeyValueData(newRows);
        onDataChange(newRows.filter(row => row.key))
    }

    const addNewRow = (data: Row[]) => {
        return data.filter(x => x.key === "").length === 0 ? [...data, { key: '', value: '' }] : data
    }

    const setNewVal = (mapFunc, index) => {
        let currentData = keyValueData.map((row, i) => i === index ? mapFunc(row) : row)
        if (currentData.every(row => row.key)) {
            onDataChange(currentData)
            currentData = addNewRow(currentData)
        }
        else {
            onDataChange(currentData.filter(row => row.key))
        }

        setKeyValueData(currentData);
    }

    return <div className={styles.keyValueTableContainer}>
        {keyValueData?.map((row, index) => {
            return <div key={index} className={styles.headerRow}>
                <div className={styles.roundInputContainer} style={{ width: "30%" }}>
                    <input
                        name="key" type="text"
                        placeholder={keyPlaceholder ? keyPlaceholder : "New key"}
                        onChange={(event) => setNewVal((row) => { return { key: event.target.value, value: row.value } }, index)}
                        value={row.key}
                        autoComplete="off"
                        spellCheck={false} />
                </div>
                <div className={styles.roundInputContainer} style={{ width: "65%" }}>
                    <input
                        name="value" type="text"
                        placeholder={valuePlaceholder ? valuePlaceholder : "New Value"}
                        onChange={(event) => setNewVal((row) => { return { key: row.key, value: event.target.value } }, index)}
                        value={row.value?.toString()}
                        autoComplete="off"
                        spellCheck={false} />
                </div>
                {(row.key !== "" || row.value !== "") && <HoverImage alt="delete" style={{ marginLeft: "5px", cursor: "pointer" }} className="deleteIcon" src={deleteIcon}
                    onClick={() => deleteRow(index)} hoverSrc={deleteIconActive} />}
            </div>
        })}
    </div>
}

export default KeyValueTable
