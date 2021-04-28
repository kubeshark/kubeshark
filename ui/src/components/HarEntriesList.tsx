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
            </div>
    </>;
};
