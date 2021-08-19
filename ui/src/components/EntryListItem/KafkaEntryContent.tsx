import {BaseEntry} from "./EntryListItem";
import React from "react";

export interface KafkaEntry extends BaseEntry{
}

interface KafkaEntryContentProps {
    entry: KafkaEntry;
}

export const KafkaEntryContent: React.FC<KafkaEntryContentProps> = ({entry}) => {

    return <>
    </>
}