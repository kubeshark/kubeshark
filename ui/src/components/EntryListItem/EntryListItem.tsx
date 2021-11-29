import React, {useState} from "react";
import styles from './EntryListItem.module.sass';
import StatusCode, {getClassification, StatusCodeClassification} from "../UI/StatusCode";
import Protocol, {ProtocolInterface} from "../UI/Protocol"
import {Summary} from "../UI/Summary";
import Queryable from "../UI/Queryable";
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
    forceSelect: boolean;
    headingMode: boolean;
}

export const EntryItem: React.FC<EntryProps> = ({entry, setFocusedEntryId, style, updateQuery, forceSelect, headingMode}) => {

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
                setIsSelected(!isSelected);
                setFocusedEntryId(entry.id.toString());
            }}
            style={{
                border: isSelected ? `1px ${entry.protocol.backgroundColor} solid` : "1px transparent solid",
                position: !headingMode ? "absolute" : "unset",
                top: style['top'],
                marginTop: style['marginTop'],
                width: !headingMode ? "calc(100% - 10px)" : "calc(100%)",
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
                    <Queryable
                        query={`service == "${entry.service}"`}
                        updateQuery={updateQuery}
                        displayIconOnMouseOver={true}
                    >
                        <span
                            title="Service Name"
                        >
                            {entry.service}
                        </span>
                    </Queryable>
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
                <Queryable
                        query={`src.ip == "${entry.sourceIp}"`}
                        updateQuery={updateQuery}
                        displayIconOnMouseOver={true}
                        flipped={true}
                        iconStyle={{marginRight: "16px"}}
                >
                    <span
                        className={`${styles.tcpInfo} ${styles.ip}`}
                        title="Source IP"
                    >
                        {entry.sourceIp}
                    </span>
                </Queryable>
                <span className={`${styles.tcpInfo}`} style={{marginTop: "18px"}}>:</span>
                <Queryable
                        query={`src.port == "${entry.sourcePort}"`}
                        updateQuery={updateQuery}
                        displayIconOnMouseOver={true}
                        flipped={true}
                        iconStyle={{marginTop: "28px"}}
                >
                    <span
                        className={`${styles.tcpInfo} ${styles.port}`}
                        title="Source Port"
                    >
                        {entry.sourcePort}
                    </span>
                </Queryable>
                {entry.isOutgoing ?
                    <Queryable
                            query={`outgoing == true`}
                            updateQuery={updateQuery}
                            displayIconOnMouseOver={true}
                            flipped={true}
                            iconStyle={{marginTop: "28px"}}
                    >
                        <img
                            src={outgoingIcon}
                            alt="Ingoing traffic"
                            title="Ingoing"
                        />
                    </Queryable>
                    :
                    <Queryable
                            query={`outgoing == true`}
                            updateQuery={updateQuery}
                            displayIconOnMouseOver={true}
                            flipped={true}
                            iconStyle={{marginTop: "28px"}}
                    >
                        <img
                            src={ingoingIcon}
                            alt="Outgoing traffic"
                            title="Outgoing"
                            onClick={() => {
                                updateQuery(`outgoing == false`)
                            }}
                        />
                    </Queryable>
                }
                <Queryable
                        query={`dst.ip == "${entry.destinationIp}"`}
                        updateQuery={updateQuery}
                        displayIconOnMouseOver={true}
                        flipped={false}
                        iconStyle={{marginTop: "28px"}}
                >
                    <span
                        className={`${styles.tcpInfo} ${styles.ip}`}
                        title="Destination IP"
                    >
                        {entry.destinationIp}
                    </span>
                </Queryable>
                <span className={`${styles.tcpInfo}`} style={{marginTop: "18px"}}>:</span>
                <Queryable
                        query={`dst.port == "${entry.destinationPort}"`}
                        updateQuery={updateQuery}
                        displayIconOnMouseOver={true}
                        flipped={false}
                >
                    <span
                        className={`${styles.tcpInfo} ${styles.port}`}
                        title="Destination Port"
                    >
                        {entry.destinationPort}
                    </span>
                </Queryable>
            </div>
            <div className={styles.timestamp}>
                <Queryable
                        query={`timestamp >= datetime("${new Date(+entry.timestamp)?.toLocaleString("en-US", {timeZone: 'UTC' })}")`}
                        updateQuery={updateQuery}
                        displayIconOnMouseOver={true}
                        flipped={false}
                >
                    <span
                        title="Timestamp"
                    >
                        {new Date(+entry.timestamp)?.toLocaleString("en-US")}
                    </span>
                </Queryable>
            </div>
        </div>
    </>

}
