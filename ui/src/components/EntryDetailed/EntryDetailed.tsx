import React from "react";
import styles from './EntryDetailed.module.sass';
import {makeStyles} from "@material-ui/core";
import {EntryType} from "../EntryListItem/EntryListItem";
import {RestEntryDetailsTitle} from "./Rest/RestEntryDetailsTitle";
import {KafkaEntryDetailsTitle} from "./Kafka/KafkaEntryDetailsTitle";
import {RestEntryDetailsContent} from "./Rest/RestEntryDetailsContent";
import {KafkaEntryDetailsContent} from "./Kafka/KafkaEntryDetailsContent";

const useStyles = makeStyles(() => ({
    entryTitle: {
        display: 'flex',
        minHeight: 46,
        maxHeight: 46,
        alignItems: 'center',
        marginBottom: 8,
        padding: 5,
        paddingBottom: 0
    }
}));

interface EntryDetailedProps {
    entryData: any;
    classes?: any;
    entryType: string;
}

export const EntryDetailed: React.FC<EntryDetailedProps> = ({classes, entryData, entryType}) => {
    const classesTitle = useStyles();

    let title, content;

    switch (entryType) {
        case EntryType.Rest:
            title = <RestEntryDetailsTitle entryData={entryData}/>;
            content = <RestEntryDetailsContent entryData={entryData}/>;
            break;
        case EntryType.Kafka:
            title = <KafkaEntryDetailsTitle entryData={entryData}/>;
            content = <KafkaEntryDetailsContent entryData={entryData}/>;
            break;
        default:
            title = <RestEntryDetailsTitle entryData={entryData}/>;
            content = <RestEntryDetailsContent entryData={entryData}/>;
            break;
    }

    return <>
        <div className={classesTitle.entryTitle}>{title}</div>
        <div className={styles.content}>
            <div className={styles.body}>
                {content}
            </div>
        </div>
    </>
};