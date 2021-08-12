import React from "react";
import styles from '../style/EntryListItem.module.sass';
import restIcon from '../assets/restIcon.svg';
import kafkaIcon from '../assets/kafkaIcon.svg';
import {RestEntry, RestEntryContent} from "./RestEntryContent";
import {KafkaEntry} from "./KafkaEntryContent";

export interface Entry {
    type: string;
    timestamp: Date;
    id: string;
    rules: Rules;
    latency: number;
}

interface Rules {
    status: boolean;
    latency: number
}

interface EntryProps {
    entry: RestEntry | KafkaEntry;
    setFocusedEntryId: (id: string) => void;
    isSelected?: boolean;
}

enum EntryType {
    Rest = "rest",
    Kafka = "kafka"
}

export const EntryItem: React.FC<EntryProps> = ({entry, setFocusedEntryId, isSelected}) => {

    let backgroundColor = "";
    if ('latency' in entry.rules) {
        if (entry.rules.latency !== -1) {
            backgroundColor = entry.rules.latency >= entry.latency ? styles.ruleSuccessRow : styles.ruleFailureRow
        } else {
            backgroundColor = entry.rules.status ? styles.ruleSuccessRow : styles.ruleFailureRow
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
        <div id={entry.id} className={`${styles.row} ${isSelected ? styles.rowSelected : backgroundColor}`}
             onClick={() => setFocusedEntryId(entry.id)}>
            {entryIcon(entry) && <div style={{width: 80}}>{<img className={styles.icon} alt="icon" src={entryIcon(entry)}/>}</div>}
            {entryContent(entry)}
            <div className={styles.timestamp}>{new Date(+entry.timestamp)?.toLocaleString()}</div>
        </div>
    </>
};

