import React from "react";
import styles from './EntryListItem.module.sass';
import restIcon from '../assets/restIcon.svg';
import kafkaIcon from '../assets/kafkaIcon.svg';
import {RestEntry, RestEntryContent} from "./RestEntryContent";
import {KafkaEntry, KafkaEntryContent} from "./KafkaEntryContent";

export interface BaseEntry {
    type: string;
    timestamp: Date;
    id: string;
    rules: Rules;
    latency: number;
}

interface Rules {
    status: boolean;
    latency: number;
    numberOfRules: number;
}

interface EntryProps {
    entry: RestEntry | KafkaEntry | any;
    setFocusedEntry: (entry: RestEntry | KafkaEntry) => void;
    isSelected?: boolean;
}

export enum EntryType {
    Rest = "rest",
    Kafka = "kafka"
}

export const EntryItem: React.FC<EntryProps> = ({entry, setFocusedEntry, isSelected}) => {

    let additionalRulesProperties = "";
    let rule = 'latency' in entry.rules
    if (rule) {
        if (entry.rules.latency !== -1) {
            if (entry.rules.latency >= entry.latency) {
                additionalRulesProperties = styles.ruleSuccessRow
            } else {
                additionalRulesProperties = styles.ruleFailureRow
            }
            if (isSelected) {
                additionalRulesProperties += ` ${entry.rules.latency >= entry.latency ? styles.ruleSuccessRowSelected : styles.ruleFailureRowSelected}`
            }
        } else {
            if (entry.rules.status) {
                additionalRulesProperties = styles.ruleSuccessRow
            } else {
                additionalRulesProperties = styles.ruleFailureRow
            }
            if (isSelected) {
                additionalRulesProperties += ` ${entry.rules.status ? styles.ruleSuccessRowSelected : styles.ruleFailureRowSelected}`
            }
        }
    }

    let icon, content;

    switch (entry.type) {
        case EntryType.Rest:
            content = <RestEntryContent entry={entry}/>;
            icon = restIcon;
            break;
        case EntryType.Kafka:
            content = <KafkaEntryContent entry={entry}/>;
            icon = kafkaIcon;
            break;
        default:
            content = <RestEntryContent entry={entry}/>;
            icon = restIcon;
            break;
    }

    return <>
        <div id={entry.id} className={`${styles.row} ${isSelected && !rule ? styles.rowSelected : additionalRulesProperties}`}
             onClick={() => setFocusedEntry(entry)}>
            {icon && <div style={{width: 80}}>{<img className={styles.icon} alt="icon" src={icon}/>}</div>}
            {content}
            <div className={styles.timestamp}>{new Date(+entry.timestamp)?.toLocaleString()}</div>
        </div>
    </>
};

