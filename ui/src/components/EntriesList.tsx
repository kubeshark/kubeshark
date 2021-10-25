import {EntryItem} from "./EntryListItem/EntryListItem";
import React, {useRef} from "react";
import styles from './style/EntriesList.module.sass';
import ScrollableFeedVirtualized from "react-scrollable-feed-virtualized";
import Api from "../helpers/api";
import down from "./assets/downImg.svg";

interface EntriesListProps {
    entries: any[];
    setEntries: (entries: any[]) => void;
    focusedEntryId: string;
    setFocusedEntryId: (id: string) => void;
    connectionOpen: boolean;
    noMoreDataBottom: boolean;
    setNoMoreDataBottom: (flag: boolean) => void;
    listEntryREF: any;
    onScrollEvent: (isAtBottom:boolean) => void;
    scrollableList: boolean;
}

enum FetchOperator {
    LT = "lt",
    GT = "gt"
}

const api = new Api();

export const EntriesList: React.FC<EntriesListProps> = ({entries, setEntries, focusedEntryId, setFocusedEntryId, connectionOpen, noMoreDataBottom, setNoMoreDataBottom, listEntryREF, onScrollEvent, scrollableList}) => {

    const scrollableRef = useRef(null);

    const getNewEntries = async () => {
        const data = await api.fetchEntries(FetchOperator.GT, entries[entries.length - 1].timestamp);
        let scrollTo;
        if(data.length === 0) {
            setNoMoreDataBottom(true);
        }
        scrollTo = document.getElementById(entries?.[entries.length -1]?.id);
        let newEntries = [...entries, ...data];
        setEntries(newEntries);
        if(scrollTo) {
            scrollTo.scrollIntoView({behavior: "smooth"});
        }
    }

    return <>
            <div className={styles.list}>
                <div id="list" ref={listEntryREF} className={styles.list}>
                    <ScrollableFeedVirtualized ref={scrollableRef} itemHeight={48} marginTop={10} onScroll={(isAtBottom) => onScrollEvent(isAtBottom)}>
                        {false /* TODO: why there is a need for something here (not necessarily false)? */}
                        {entries.map(entry => <EntryItem key={entry.id}
                                                        entry={entry}
                                                        setFocusedEntryId={setFocusedEntryId}
                                                        isSelected={focusedEntryId === entry.id.toString()}
                                                        style={{}}/>)}
                    </ScrollableFeedVirtualized>
                    {!connectionOpen && !noMoreDataBottom && <div className={styles.fetchButtonContainer}>
                        <div className={styles.styledButton} onClick={() => getNewEntries()}>Fetch more entries</div>
                    </div>}
                    <button type="button"
                        className={`${styles.btnLive} ${scrollableList ? styles.showButton : styles.hideButton}`}
                        onClick={(_) => scrollableRef.current.jumpToBottom()}>
                        <img alt="down" src={down} />
                    </button>
                </div>

                {entries?.length > 0 && <div className={styles.footer}>
                    <div><b>{entries?.length}</b> requests</div>
                    <div>Started listening at <span style={{marginRight: 5, fontWeight: 600, fontSize: 13}}>{new Date(+entries[0].timestamp)?.toLocaleString()}</span></div>
                </div>}
            </div>
    </>;
};
