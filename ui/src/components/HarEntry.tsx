import React from "react";
import styles from './style/HarEntry.module.sass';
import StatusCode from "./StatusCode";
import {EndpointPath} from "./EndpointPath";

interface HAREntry {
    method?: string,
    path: string,
    service: string,
    id: string,
    statusCode?: number;
    url?: string;
    isCurrentRevision?: boolean;
    timestamp: Date;
}

interface HAREntryProps {
    entry: HAREntry;
    setFocusedEntryId: (id: string) => void;
    isSelected?: boolean;
}

export const HarEntry: React.FC<HAREntryProps> = ({entry, setFocusedEntryId, isSelected}) => {

    return <>
        <div id={entry.id} className={`${styles.row} ${isSelected ? styles.rowSelected : ''}`} onClick={() => setFocusedEntryId(entry.id)}>
            {entry.statusCode && <div>
                <StatusCode statusCode={entry.statusCode}/>
            </div>}
            <div className={styles.endpointServiceContainer}>
                <EndpointPath method={entry.method} path={entry.path}/>
                <div className={styles.service}>
                    {entry.service}
                </div>
            </div>
            <div className={styles.timestamp}>{new Date(+entry.timestamp)?.toLocaleString()}</div>
        </div>
    </>
};