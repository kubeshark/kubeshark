import React from "react";
import styles from './style/HarEntry.module.sass';
import StatusCode from "./StatusCode";
import {EndpointPath} from "./EndpointPath";
import incomingIcon from "./assets/incoming-traffic.svg"
import outgoingIcon from "./assets/outgoing-traffic.svg"

interface HAREntry {
    method?: string,
    path: string,
    service: string,
    id: string,
    statusCode?: number;
    url?: string;
    isCurrentRevision?: boolean;
    timestamp: Date;
	isOutgoing?: boolean;
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
            <div className={styles.directionContainer}>
                {entry.isOutgoing ?
                    <div className={styles.outgoingIcon}>
                        <img src={outgoingIcon} alt="outgoing" title="outgoing"/>
                    </div>
                    :
                    <div className={styles.incomingIcon}>
                        <img src={incomingIcon} alt="incoming" title="incoming"/>
                    </div>
                }
            </div>
            <div className={styles.timestamp}>{new Date(+entry.timestamp)?.toLocaleString()}</div>
        </div>
    </>
};
