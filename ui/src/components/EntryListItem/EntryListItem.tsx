import React from "react";
import styles from './EntryListItem.module.sass';
import StatusCode, {getClassification, StatusCodeClassification} from "../UI/StatusCode";
import Protocol, {ProtocolInterface} from "../UI/Protocol"
import {EndpointPath} from "../UI/EndpointPath";
import ingoingIconSuccess from "../assets/ingoing-traffic-success.svg"
import ingoingIconFailure from "../assets/ingoing-traffic-failure.svg"
import ingoingIconNeutral from "../assets/ingoing-traffic-neutral.svg"
import outgoingIconSuccess from "../assets/outgoing-traffic-success.svg"
import outgoingIconFailure from "../assets/outgoing-traffic-failure.svg"
import outgoingIconNeutral from "../assets/outgoing-traffic-neutral.svg"

interface Entry {
    protocol: ProtocolInterface,
    method?: string,
    summary: string,
    service: string,
    id: string,
    status_code?: number;
    url?: string;
    timestamp: Date;
    source_ip: string,
    source_port: string,
    destination_ip: string,
    destination_port: string,
	isOutgoing?: boolean;
    latency: number;
    rules: Rules;
}

interface Rules {
    status: boolean;
    latency: number;
    numberOfRules: number;
}

interface EntryProps {
    entry: Entry;
    setFocusedEntryId: (id: string) => void;
    isSelected?: boolean;
}

export const EntryItem: React.FC<EntryProps> = ({entry, setFocusedEntryId, isSelected}) => {
    const classification = getClassification(entry.status_code)
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
    // let additionalRulesProperties = "";
    // let ruleSuccess: boolean;
    let rule = 'latency' in entry.rules
    if (rule) {
        if (entry.rules.latency !== -1) {
            if (entry.rules.latency >= entry.latency) {
                // additionalRulesProperties = styles.ruleSuccessRow
                // ruleSuccess = true
            } else {
                // additionalRulesProperties = styles.ruleFailureRow
                // ruleSuccess = false
            }
            if (isSelected) {
                // additionalRulesProperties += ` ${entry.rules.latency >= entry.latency ? styles.ruleSuccessRowSelected : styles.ruleFailureRowSelected}`
            }
        } else {
            if (entry.rules.status) {
                // additionalRulesProperties = styles.ruleSuccessRow
                // ruleSuccess = true
            } else {
                // additionalRulesProperties = styles.ruleFailureRow
                // ruleSuccess = false
            }
            if (isSelected) {
                // additionalRulesProperties += ` ${entry.rules.status ? styles.ruleSuccessRowSelected : styles.ruleFailureRowSelected}`
            }
        }
    }
    let backgroundColor = "";
    if ('latency' in entry.rules) {
        if (entry.rules.latency !== -1) {
            backgroundColor = entry.rules.latency >= entry.latency ? styles.ruleSuccessRow : styles.ruleFailureRow
        } else {
            backgroundColor = entry.rules.status ? styles.ruleSuccessRow : styles.ruleFailureRow
        }
    }
    return <>
        <div
            id={entry.id}
            className={`${styles.row}
            ${isSelected ? styles.rowSelected : backgroundColor}`}
            onClick={() => setFocusedEntryId(entry.id)}
            style={{border: isSelected ? `1px ${entry.protocol.background_color} solid` : "1px transparent solid"}}
        >
            <Protocol protocol={entry.protocol} horizontal={false}/>
            {((entry.protocol.name === "http" && "status_code" in entry) || entry.status_code !== 0) && <div>
                <StatusCode statusCode={entry.status_code}/>
            </div>}
            <div className={styles.endpointServiceContainer}>
                <EndpointPath method={entry.method} path={entry.summary}/>
                <div className={styles.service}>
                    <span title="Service Name">{entry.service}</span>
                </div>
            </div>
            <div className={styles.directionContainer}>
                <span className={styles.port} title="Source Port">{entry.source_port}</span>
                {entry.isOutgoing ?
                    <img src={outgoingIcon} alt="Ingoing traffic" title="Ingoing"/>
                    :
                    <img src={ingoingIcon} alt="Outgoing traffic" title="Outgoing"/>
                }
                <span className={styles.port} title="Destination Port">{entry.destination_port}</span>
            </div>
            <div className={styles.timestamp}>
                <span title="Timestamp">
                    {new Date(+entry.timestamp)?.toLocaleString()}
                </span>
            </div>
        </div>
    </>
};
