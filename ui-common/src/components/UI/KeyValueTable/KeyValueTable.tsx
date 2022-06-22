import React from "react";
import { useEffect, useState } from "react";
import styles from "./KeyValueTable.module.sass"
import deleteIcon from "delete.svg"

interface KeyValueTableProps {
    data: any
    onDataChange: (data: any) => void
}

const KeyValueTable: React.FC<KeyValueTableProps> = ({ data, onDataChange }) => {

    const [keyValueData, setKeyValueData] = useState([])

    useEffect(() => {
        if (!data) return;
        const currentData = Object.entries(data)
            .map(([key, val]) => { return { "key": key, "value": val } })
        const newData = [...currentData, { key: "", value: "" }]
        setKeyValueData(newData)
    }, [data])

    const deleteHeader = (index) => {
        const newHeaders = [...keyValueData];
        newHeaders.splice(index, 1);
        setKeyValueData(newHeaders);
    }

    const addNewRow = (data) => {
        return data.filter(x => x.key === "" || x.value === "").length === 0 ? [...data, { key: '', value: '' }] : data
    }

    const onHeaderKeyChange = (index, newKey) => {
        const currentData = keyValueData.map((row, i) => i === index ? { key: newKey, value: row.value } : row)
        const newData = addNewRow(currentData)
        setKeyValueData(newData);
    }

    const onHeaderValueChange = (index, newValue) => {
        const currentData = keyValueData.map((row, i) => i === index ? { key: row.key, value: newValue } : row)
        const newData = addNewRow(currentData)
        setKeyValueData(newData);
    }
    return <div className={styles.tryNowHeadersContainer}>
        {keyValueData?.map((row, index) => {
            return <div key={index} className={styles.headerRow}>
                <div className={styles.roundInputContainer} style={{ width: "30%" }}>
                    <input
                        name="key" type="text"
                        placeholder="Add header"
                        onChange={(event) => onHeaderKeyChange(index, event.target.value)}
                        value={row.key}
                        autoComplete="off"
                        spellCheck={false} />
                </div>
                <div className={styles.roundInputContainer} style={{ width: "65%" }}>
                    <input
                        name="value" type="text"
                        placeholder="Add header content"
                        onChange={(event) => onHeaderValueChange(index, event.target.value)}
                        value={row.value?.toString()}
                        autoComplete="off"
                        spellCheck={false} />
                </div>
                {(row.key !== "" || row.value !== "") && <img alt="delete" style={{ marginLeft: "5px", cursor: "pointer" }} className="deleteIcon" src={deleteIcon}
                    onClick={() => deleteHeader(index)} />}
            </div>
        })}
    </div>
}

export default KeyValueTable
