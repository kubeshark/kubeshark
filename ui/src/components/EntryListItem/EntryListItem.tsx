import React from "react";
import styles from './EntryListItem.module.sass';
import restIcon from '../assets/restIcon.svg';
import kafkaIcon from '../assets/kafkaIcon.svg';
import {RestEntry, RestEntryContent} from "./RestEntryContent";
import {KafkaEntry} from "./KafkaEntryContent";

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
    entry: RestEntry | KafkaEntry;
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

    const entryContent = (entry) => {
        let entryComponent;
        switch (entry.type) {
            case EntryType.Rest:
                entryComponent = <RestEntryContent entry={entry}/>;
                break;
            default:
                entryComponent = <RestEntryContent entry={entry}/>;
                break;
        }
        return entryComponent;
    }

    const entryIcon = (entry) => {
        // if(entry.path.indexOf("items") > -1) return kafkaIcon;
        // return restIcon;
        let icon;
        switch (entry.type) {
            case EntryType.Rest:
                icon = restIcon;
                break;
            case EntryType.Kafka:
                icon = kafkaIcon;
                break;
            default:
                icon = null;
                break;
        }
        return icon;
    }

    return <>
        <div id={entry.id} className={`${styles.row} ${isSelected ? styles.rowSelected : additionalRulesProperties}`}
             onClick={() => setFocusedEntry(entry)}>
            {entryIcon(entry) && <div style={{width: 80}}>{<img className={styles.icon} alt="icon" src={entryIcon(entry)}/>}</div>}
            {entryContent(entry)}
            <div className={styles.timestamp}>{new Date(+entry.timestamp)?.toLocaleString()}</div>
        </div>
    </>
};

