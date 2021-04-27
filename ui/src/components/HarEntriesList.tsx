import {HarEntry} from "./HarEntry";
import React, {useEffect, useRef} from "react";
import styles from './style/HarEntriesList.module.sass';

interface HarEntriesListProps {
    entries: any[];
    focusedEntryId: string;
    setFocusedEntryId: (id: string) => void
}

export const HarEntriesList: React.FC<HarEntriesListProps> = ({entries, focusedEntryId, setFocusedEntryId}) => {
    const entriesDiv = useRef(null);
    const totalCount = null; //todo

    // Scroll to bottom in case results do not fit in screen
    useEffect(() => {
        if (entriesDiv.current && totalCount > 0) {
            entriesDiv.current.scrollTop = entriesDiv.current.scrollHeight;
        }
    }, [entriesDiv, totalCount])

    return <>
        <div ref={entriesDiv} className={styles.list}>
            {entries?.map(entry => <HarEntry key={entry.id}
                                             entry={entry}
                                             setFocusedEntryId={setFocusedEntryId}
                                             isSelected={focusedEntryId === entry.id}
            />)}

        </div>
    </>;
};
