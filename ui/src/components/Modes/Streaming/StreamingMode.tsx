import React from "react";
import styles from '../../style/EntriesList.module.sass';
import {Filters} from "../../Filters";
import {StreamingEntriesList} from "./StreamingEntriesList";

interface StreamingModeProps {
    query: string
    setQuery: any
    queryBackgroundColor: string
    ws: any
    openWebSocket: (query: string, resetEntriesBuffer: boolean) => void;
    entries: any[];
    listEntryREF: any;
    onSnapBrokenEvent: () => void;
    isSnappedToBottom: boolean;
    setIsSnappedToBottom: any;
    queriedCurrent: number;
    queriedTotal: number;
    startTime: number;
}

export const StreamingMode: React.FC<StreamingModeProps> = ({query, setQuery, queryBackgroundColor, ws, openWebSocket, entries, listEntryREF, onSnapBrokenEvent, isSnappedToBottom, setIsSnappedToBottom, queriedCurrent, queriedTotal, startTime}) => {

    return <>
        <Filters
            query={query}
            setQuery={setQuery}
            backgroundColor={queryBackgroundColor}
            ws={ws}
            openWebSocket={openWebSocket}
        />
        <div className={styles.container}>
            <StreamingEntriesList
                entries={entries}
                listEntryREF={listEntryREF}
                onSnapBrokenEvent={onSnapBrokenEvent}
                isSnappedToBottom={isSnappedToBottom}
                setIsSnappedToBottom={setIsSnappedToBottom}
                queriedCurrent={queriedCurrent}
                queriedTotal={queriedTotal}
                startTime={startTime}
            />
        </div>
</>;
};
