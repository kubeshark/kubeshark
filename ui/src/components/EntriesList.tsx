import React, {useCallback, useEffect, useMemo, useRef, useState} from "react";
import styles from './style/EntriesList.module.sass';
import ScrollableFeedVirtualized from "react-scrollable-feed-virtualized";
import {EntryItem} from "./EntryListItem/EntryListItem";
import down from "./assets/downImg.svg";
import spinner from './assets/spinner.svg';
import Api from "../helpers/api";

interface EntriesListProps {
    entries: any[];
    setEntries: any;
    query: string;
    listEntryREF: any;
    onSnapBrokenEvent: () => void;
    isSnappedToBottom: boolean;
    setIsSnappedToBottom: any;
    queriedCurrent: number;
    setQueriedCurrent: any;
    queriedTotal: number;
    startTime: number;
    noMoreDataTop: boolean;
    setNoMoreDataTop: (flag: boolean) => void;
    focusedEntryId: string;
    setFocusedEntryId: (id: string) => void;
    updateQuery: any;
    leftOffTop: number;
    setLeftOffTop: (leftOffTop: number) => void;
    reconnectWebSocket: any;
}

const api = new Api();

export const EntriesList: React.FC<EntriesListProps> = ({entries, setEntries, query, listEntryREF, onSnapBrokenEvent, isSnappedToBottom, setIsSnappedToBottom, queriedCurrent, setQueriedCurrent, queriedTotal, startTime, noMoreDataTop, setNoMoreDataTop, focusedEntryId, setFocusedEntryId, updateQuery, leftOffTop, setLeftOffTop, reconnectWebSocket}) => {
    const [loadMoreTop, setLoadMoreTop] = useState(false);
    const [isLoadingTop, setIsLoadingTop] = useState(false);
    const scrollableRef = useRef(null);

    useEffect(() => {
        const list = document.getElementById('list').firstElementChild;
        list.addEventListener('scroll', (e) => {
            const el: any = e.target;
            if(el.scrollTop === 0) {
                setLoadMoreTop(true);
            } else {
                setNoMoreDataTop(false);
                setLoadMoreTop(false);
            }
        });
    }, [setLoadMoreTop, setNoMoreDataTop]);

    const memoizedEntries = useMemo(() => {
        return entries;
    },[entries]);

    const getOldEntries = useCallback(async () => {
        if (leftOffTop <= 0) {
            return;
        }
        setIsLoadingTop(true);
        setLoadMoreTop(false);
        const data = await api.fetchEntries(leftOffTop, -1, query, 100, 3000);
        setLeftOffTop(data.meta.leftOff);

        let scrollTo: boolean;
        if (data.meta.leftOff === 0) {
            setNoMoreDataTop(true);
            scrollTo = false;
        } else {
            scrollTo = true;
        }
        setIsLoadingTop(false);

        let incomingEntries = [];
        data.data.forEach((entry: any) => {
            incomingEntries.push(
                <EntryItem
                    key={entry.id}
                    entry={entry}
                    focusedEntryId={focusedEntryId}
                    setFocusedEntryId={setFocusedEntryId}
                    style={{}}
                    updateQuery={updateQuery}
                    headingMode={false}
                />
            );
        });
        const newEntries = [...incomingEntries, ...entries];
        setEntries(newEntries);

        setQueriedCurrent(queriedCurrent + data.meta.current);

        if (scrollTo) {
            scrollableRef.current.scrollToIndex(incomingEntries.length - 1);
        }
    },[setLoadMoreTop, setIsLoadingTop, entries, setEntries, query, setNoMoreDataTop, focusedEntryId, setFocusedEntryId, updateQuery, leftOffTop, setLeftOffTop, queriedCurrent, setQueriedCurrent]);

    useEffect(() => {
        if(!loadMoreTop || noMoreDataTop) return;
        getOldEntries();
    }, [loadMoreTop, noMoreDataTop, getOldEntries]);

    return <>
            <div className={styles.list}>
                <div id="list" ref={listEntryREF} className={styles.list}>
                    {isLoadingTop && <div className={styles.spinnerContainer}>
                        <img alt="spinner" src={spinner} style={{height: 25}}/>
                    </div>}
                    {noMoreDataTop && <div id="noMoreDataTop" className={styles.noMoreDataAvailable}>No more data available</div>}
                    <ScrollableFeedVirtualized ref={scrollableRef} itemHeight={48} marginTop={10} onSnapBroken={onSnapBrokenEvent}>
                        {false /* TODO: why there is a need for something here (not necessarily false)? */}
                        {memoizedEntries}
                    </ScrollableFeedVirtualized>
                    <button type="button"
                        className={`${styles.btnLive} ${isSnappedToBottom ? styles.hideButton : styles.showButton}`}
                        onClick={(_) => {
                            reconnectWebSocket();
                            scrollableRef.current.jumpToBottom();
                            setIsSnappedToBottom(true);
                        }}>
                        <img alt="down" src={down} />
                    </button>
                </div>

                <div className={styles.footer}>
                    <div>Displaying <b>{entries?.length}</b> results out of <b>{queriedTotal}</b> total</div>
                    {startTime !== 0 && <div>Started listening at <span style={{marginRight: 5, fontWeight: 600, fontSize: 13}}>{new Date(startTime).toLocaleString()}</span></div>}
                </div>
            </div>
    </>;
};
