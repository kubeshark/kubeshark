import React, {useRef, useState} from "react";
import styles from '../../style/EntriesList.module.sass';
import ScrollableFeed from "react-scrollable-feed";
import {EntryItem} from "../../EntryListItem/EntryListItem";
import spinner from '../../assets/spinner.svg';

interface PagingEntriesListProps {
    entries: any[];
    listEntryREF: any;
}

export const PagingEntriesList: React.FC<PagingEntriesListProps> = ({entries, listEntryREF}) => {

    const [loadMoreTop, setLoadMoreTop] = useState(false);
    const [isLoadingTop, setIsLoadingTop] = useState(false);
    const scrollableRef = useRef(null);

    console.log(entries);

    return <>
            <div className={styles.list}>
                <div id="list" ref={listEntryREF} className={styles.list} >
                    {isLoadingTop && <div className={styles.spinnerContainer}>
                        <img alt="spinner" src={spinner} style={{height: 25}}/>
                    </div>}
                    <ScrollableFeed ref={scrollableRef} onScroll={(isAtBottom) => console.log(isAtBottom)}>
                        {entries}
                    </ScrollableFeed>
                </div>

                {entries?.length > 0 && <div className={styles.footer}>
                    <div><b>{entries?.length !== entries.length && `${entries?.length} / `} {entries?.length}</b> requests</div>
                    <div>Started listening at <span style={{marginRight: 5, fontWeight: 600, fontSize: 13}}>{new Date(+entries[0].timestamp)?.toLocaleString()}</span></div>
                </div>}
            </div>
    </>;
};
