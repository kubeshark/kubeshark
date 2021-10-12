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
    statusCode?: number;
    url?: string;
    timestamp: Date;
    sourceIp: string,
    sourcePort: string,
    destinationIp: string,
    destinationPort: string,
    isOutgoing?: boolean;
    latency: number;
    rules: Rules;
    contractStatus: number,
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
    style: object;
}

export const EntryItem: React.FC<EntryProps> = ({entry, setFocusedEntryId, isSelected, style}) => {
    const classification = getClassification(entry.statusCode)
    const numberOfRules = entry.rules.numberOfRules
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
    let additionalRulesProperties = "";
    let ruleSuccess: boolean;
    let rule = 'latency' in entry.rules
    if (rule) {
        if (entry.rules.latency !== -1) {
            if (entry.rules.latency >= entry.latency || !('latency' in entry)) {
                additionalRulesProperties = styles.ruleSuccessRow
                ruleSuccess = true
            } else {
                additionalRulesProperties = styles.ruleFailureRow
                ruleSuccess = false
            }
            if (isSelected) {
                additionalRulesProperties += ` ${entry.rules.latency >= entry.latency ? styles.ruleSuccessRowSelected : styles.ruleFailureRowSelected}`
            }
        } else {
            if (entry.rules.status) {
                additionalRulesProperties = styles.ruleSuccessRow
                ruleSuccess = true
            } else {
                additionalRulesProperties = styles.ruleFailureRow
                ruleSuccess = false
            }
            if (isSelected) {
                additionalRulesProperties += ` ${entry.rules.status ? styles.ruleSuccessRowSelected : styles.ruleFailureRowSelected}`
            }
        }
    }

    var contractEnabled = true;
    var contractText = "";
    switch (entry.contractStatus) {
        case 0:
            contractEnabled = false;
            break;
        case 1:
            additionalRulesProperties = styles.ruleSuccessRow
            ruleSuccess = true
            contractText = "No Breaches"
            break;
        case 2:
            additionalRulesProperties = styles.ruleFailureRow
            ruleSuccess = false
            contractText = "Breach"
            break;
        default:
            break;
    }

    return <>
        <div
            id={entry.id}
            className={`${styles.row}
            ${isSelected && !rule && !contractEnabled ? styles.rowSelected : additionalRulesProperties}`}
            onClick={() => setFocusedEntryId(entry.id)}
            style={{
                border: isSelected ? `1px ${entry.protocol.backgroundColor} solid` : "1px transparent solid",
                position: "absolute",
                top: style['top'],
                marginTop: style['marginTop'],
                width: "calc(100% - 25px)",
            }}
        >
            <Protocol protocol={entry.protocol} horizontal={false}/>
            {((entry.protocol.name === "http" && "statusCode" in entry) || entry.statusCode !== 0) && <div>
                <StatusCode statusCode={entry.statusCode}/>
            </div>}
            <div className={styles.endpointServiceContainer}>
                <EndpointPath method={entry.method} path={entry.summary}/>
                <div className={styles.service}>
                    <span title="Service Name">{entry.service}</span>
                </div>
            </div>
            {
                rule ?
                    <div className={`${styles.ruleNumberText} ${ruleSuccess ? styles.ruleNumberTextSuccess : styles.ruleNumberTextFailure}`}>
                        {`Rules (${numberOfRules})`}
                    </div>
                : ""
            }
            {
                contractEnabled ?
                    <div className={`${styles.ruleNumberText} ${ruleSuccess ? styles.ruleNumberTextSuccess : styles.ruleNumberTextFailure}`}>
                        {contractText}
                    </div>
                : ""
            }
            <div className={styles.directionContainer}>
                <span className={styles.port} title="Source Port">{entry.sourcePort}</span>
                {entry.isOutgoing ?
                    <img src={outgoingIcon} alt="Ingoing traffic" title="Ingoing"/>
                    :
                    <img src={ingoingIcon} alt="Outgoing traffic" title="Outgoing"/>
                }
                <span className={styles.port} title="Destination Port">{entry.destinationPort}</span>
            </div>
            <div className={styles.timestamp}>
                <span title="Timestamp">
                    {new Date(+entry.timestamp)?.toLocaleString()}
                </span>
            </div>
        </div>
    </>

}
