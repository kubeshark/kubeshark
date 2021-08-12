import {Entry} from "./EntryListItem";
import React from "react";

export interface KafkaEntry extends Entry{
}

interface KafkaEntryContentProps {
    entry: KafkaEntry;
}

export const KafkaEntryContent: React.FC<KafkaEntryContentProps> = ({entry}) => {

    return <>
    </>
}