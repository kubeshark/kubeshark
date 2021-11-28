import React, {useState} from "react";
import styles from './EntryListItem.module.sass';
import StatusCode, {getClassification, StatusCodeClassification} from "../UI/StatusCode";
import Protocol, {ProtocolInterface} from "../UI/Protocol"
import {Summary} from "../UI/Summary";
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
    id: number,
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
    style: object;
    updateQuery: any;
    addSelectedEntries: (id: number) => void;
    removeSelectedEntries: (id: number) => void;
    forceSelect: boolean;
    headingMode: boolean;
}

export const EntryItem: React.FC<EntryProps> = ({entry, setFocusedEntryId, style, updateQuery, addSelectedEntries, removeSelectedEntries, forceSelect, headingMode}) => {

    const [isSelected, setIsSelected] = useState(!forceSelect ? false : true);

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
    let ruleSuccess = true;
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
            id={entry.id.toString()}
            className={`${styles.row}
            ${isSelected && !rule && !contractEnabled ? styles.rowSelected : additionalRulesProperties}`}
            onClick={() => {
                if (!setFocusedEntryId) return;
                if (!headingMode) {
                    if (isSelected) {
                        removeSelectedEntries(entry.id);
                    } else {
                        addSelectedEntries(entry.id);
                    }
                }
                setIsSelected(!isSelected);
                setFocusedEntryId(entry.id.toString());
            }}
            style={{
                border: isSelected ? `1px ${entry.protocol.backgroundColor} solid` : "1px transparent solid",
                position: !headingMode ? "absolute" : "unset",
                top: style['top'],
                marginTop: style['marginTop'],
                width: !headingMode ? "calc(100% - 25px)" : "calc(100% - 18px)",
            }}
        >
            {!headingMode ? <Protocol
                protocol={entry.protocol}
                horizontal={false}
                updateQuery={updateQuery}
            /> : null}
            {((entry.protocol.name === "http" && "statusCode" in entry) || entry.statusCode !== 0) && <div>
                <StatusCode statusCode={entry.statusCode} updateQuery={updateQuery}/>
            </div>}
            <div className={styles.endpointServiceContainer}>
                <Summary method={entry.method} summary={entry.summary} updateQuery={updateQuery}/>
                <div className={styles.service}>
                    <span
                        title="Service Name"
                        className="queryable"
                        onClick={() => {
                            updateQuery(`service == "${entry.service}"`)
                        }}
                    >
                        {entry.service}
                    </span>
                </div>
            </div>
            {
                rule ?
                    <div className={`${styles.ruleNumberText} ${ruleSuccess ? styles.ruleNumberTextSuccess : styles.ruleNumberTextFailure} ${rule && contractEnabled ? styles.separatorRight : ""}`}>
                        {`Rules (${numberOfRules})`}
                    </div>
                : ""
            }
            {
                contractEnabled ?
                    <div className={`${styles.ruleNumberText} ${ruleSuccess ? styles.ruleNumberTextSuccess : styles.ruleNumberTextFailure} ${rule && contractEnabled ? styles.separatorLeft : ""}`}>
                        {contractText}
                    </div>
                : ""
            }
            <div className={styles.separatorRight}>
                <span
                    className={`queryable ${styles.tcpInfo} ${styles.ip}`}
                    title="Source IP"
                    onClick={() => {
                        updateQuery(`src.ip == "${entry.sourceIp}"`)
                    }}
                >
                    {entry.sourceIp}
                </span>
                <span className={`${styles.tcpInfo}`}>:</span>
                <span
                    className={`queryable ${styles.tcpInfo} ${styles.port}`}
                    title="Source Port"
                    onClick={() => {
                        updateQuery(`src.port == "${entry.sourcePort}"`)
                    }}
                >
                    {entry.sourcePort}
                </span>
                {entry.isOutgoing ?
                    <img
                        src={outgoingIcon}
                        alt="Ingoing traffic"
                        title="Ingoing"
                        onClick={() => {
                            updateQuery(`outgoing == true`)
                        }}
                    />
                    :
                    <img
                        src={ingoingIcon}
                        alt="Outgoing traffic"
                        title="Outgoing"
                        onClick={() => {
                            updateQuery(`outgoing == false`)
                        }}
                    />
                }
                <span
                    className={`queryable ${styles.tcpInfo} ${styles.ip}`}
                    title="Destination IP"
                    onClick={() => {
                        updateQuery(`dst.ip == "${entry.destinationIp}"`)
                    }}
                >
                    {entry.destinationIp}
                </span>
                <span className={`${styles.tcpInfo}`}>:</span>
                <span
                    className={`queryable ${styles.tcpInfo} ${styles.port}`}
                    title="Destination Port"
                    onClick={() => {
                        updateQuery(`dst.port == "${entry.destinationPort}"`)
                    }}
                >
                    {entry.destinationPort}
                </span>
            </div>
            <div className={styles.timestamp}>
                <span
                    title="Timestamp"
                    className="queryable"
                    onClick={() => {
                        updateQuery(`timestamp >= datetime("${new Date(+entry.timestamp)?.toLocaleString("en-US", {timeZone: 'UTC' })}")`)
                    }}
                >
                    {new Date(+entry.timestamp)?.toLocaleString("en-US")}
                </span>
            </div>
        </div>
    </>

}
