import React from "react";
import styles from '../../style/EntriesList.module.sass';
import {Filters} from "../../Filters";
import {PagingEntriesList} from "./PagingEntriesList";

interface PagingModeProps {
    query: string
    setQuery: any
    queryBackgroundColor: string
    ws: any
    openWebSocket: (query: string, resetEntriesBuffer: boolean) => void;
    entries: any[];
    listEntryREF: any;
}

export const PagingMode: React.FC<PagingModeProps> = ({query, setQuery, queryBackgroundColor, ws, openWebSocket, entries, listEntryREF}) => {

    return <>
        <Filters
            query={query}
            setQuery={setQuery}
            backgroundColor={queryBackgroundColor}
            ws={ws}
            openWebSocket={openWebSocket}
        />
        <div className={styles.container}>
            <PagingEntriesList
                entries={entries}
                listEntryREF={listEntryREF}
            />
        </div>
</>;
};
