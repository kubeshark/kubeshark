import React from "react";
import styles from './style/HarEntry.module.sass';
import StatusCode, {getClassification, StatusCodeClassification} from "./StatusCode";
import {EndpointPath} from "./EndpointPath";
import ingoingIconSuccess from "./assets/ingoing-traffic-success.svg"
import ingoingIconFailure from "./assets/ingoing-traffic-failure.svg"
import ingoingIconNeutral from "./assets/ingoing-traffic-neutral.svg"
import outgoingIconSuccess from "./assets/outgoing-traffic-success.svg"
import outgoingIconFailure from "./assets/outgoing-traffic-failure.svg"
import outgoingIconNeutral from "./assets/outgoing-traffic-neutral.svg"

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
    latency: number;
    applicableRules: ApplicableRules;
}

interface ApplicableRules {
    status: boolean;
    latency: number
}

interface HAREntryProps {
    entry: HAREntry;
    setFocusedEntryId: (id: string) => void;
    isSelected?: boolean;
}

export const HarEntry: React.FC<HAREntryProps> = ({entry, setFocusedEntryId, isSelected}) => {
    const classification = getClassification(entry.statusCode)
    let ingoingIcon;
    let outgoingIcon;
    switch(classification) {
        case StatusCodeClassification.SUCCESS: {
            ingoingIcon = ingoingIconSuccess;
            outgoingIcon = outgoingIconSuccess;
            break;
        }
        case StatusCodeClassification.FAILURE: {
            ingoingIcon = ingoingIconFailure;
            outgoingIcon = outgoingIconFailure;
            break;
        }
        case StatusCodeClassification.NEUTRAL: {
            ingoingIcon = ingoingIconNeutral;
            outgoingIcon = outgoingIconNeutral;
            break;
        }
    }
    let backgroundColor = "";
    if (entry.applicableRules.latency !== -1) {
        const latency = entry.applicableRules.latency
        if (latency > entry.latency) {
            backgroundColor = styles.ruleSuccessRow
        } else {
            backgroundColor = styles.ruleFailureRow
        }
    } else if (!entry.applicableRules.status) {
        backgroundColor = styles.ruleFailureRow
    } else if (entry.applicableRules.status) {
        backgroundColor = styles.ruleSuccessRow
    }
    return <>
        <div id={entry.id} className={`${styles.row} ${isSelected ? styles.rowSelected : backgroundColor}`} onClick={() => setFocusedEntryId(entry.id)}>
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
                    <img src={outgoingIcon} alt="outgoing traffic" title="outgoing"/>
                    :
                    <img src={ingoingIcon} alt="ingoing traffic" title="ingoing"/>
                }
            </div>
            <div className={styles.timestamp}>{new Date(+entry.timestamp)?.toLocaleString()}</div>
        </div>
    </>
};
