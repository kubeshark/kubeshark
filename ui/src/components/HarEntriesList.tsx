import {HarEntry} from "./HarEntry";
import React, {useEffect, useRef} from "react";
import styles from './style/HarEntriesList.module.sass';
import ScrollableFeed from 'react-scrollable-feed'

interface HarEntriesListProps {
    entries: any[];
    focusedEntryId: string;
    setFocusedEntryId: (id: string) => void
}

export const HarEntriesList: React.FC<HarEntriesListProps> = ({entries, focusedEntryId, setFocusedEntryId}) => {

    return <>
            <div className={styles.list}>
                <ScrollableFeed>
                    {entries?.map(entry => <HarEntry key={entry.id}
                                                     entry={entry}
                                                     setFocusedEntryId={setFocusedEntryId}
                                                     isSelected={focusedEntryId === entry.id}
                    />)}
                </ScrollableFeed>
                {entries?.length > 0 && <div className={styles.footer}>
                    <div><b>{entries?.length}</b> requests</div>
                    <div>Started listening at <span style={{marginRight: 5, fontWeight: 600, fontSize: 13}}>{new Date(+entries[0].timestamp*1000)?.toLocaleString()}</span></div>
                </div>}
            </div>
    </>;
};
