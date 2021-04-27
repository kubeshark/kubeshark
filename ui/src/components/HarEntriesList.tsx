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

    // Reverse entries for displaying in ascending order
    // const entries = harStore.data.currentPagedResults.value.slice();
    return <>
        {/*{!isKafka && harStore.data.latestErrorType === ErrorType.TIMEOUT && <div>Timed out - many entries. Try to remove filters and try again</div>}*/}
        {/*{!isKafka && harStore.data.latestErrorType === ErrorType.GENERAL && <div>Error getting entries</div>}*/}
        {/*{!isKafka && harStore.data.isInitialized && harStore.data.fetchedCount === 0 && <div>No entries found</div>}*/}
        {/*{isKafka && selectedModelStore.kafka.sampleMessages.isLoading && <LoadingOverlay delay={0}/>}*/}
        <div ref={entriesDiv} className={styles.list}>
            {entries?.map(entry => <HarEntry key={entry.id}
                                             entry={entry}
                                             setFocusedEntryId={setFocusedEntryId}
                                             isSelected={focusedEntryId === entry.id}
            />)}

        </div>
    </>;
};
