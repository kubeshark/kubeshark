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
    entriesBuffer: any[];
    setEntriesBuffer: any;
    query: string;
    listEntryREF: any;
    onSnapBrokenEvent: () => void;
    isSnappedToBottom: boolean;
    setIsSnappedToBottom: any;
    queriedCurrent: number;
    queriedTotal: number;
    startTime: number;
    noMoreDataTop: boolean;
    setNoMoreDataTop: (flag: boolean) => void;
    focusedEntryId: string;
    setFocusedEntryId: (id: string) => void;
    updateQuery: any;
    leftOffTop: any;
    setLeftOffTop: (leftOffTop: number) => void;
}

const api = new Api();

export const EntriesList: React.FC<EntriesListProps> = ({entries, setEntries, entriesBuffer, setEntriesBuffer, query, listEntryREF, onSnapBrokenEvent, isSnappedToBottom, setIsSnappedToBottom, queriedCurrent, queriedTotal, startTime, noMoreDataTop, setNoMoreDataTop, focusedEntryId, setFocusedEntryId, updateQuery, leftOffTop, setLeftOffTop}) => {

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
                setLoadMoreTop(false);
            }
        });
    }, []);

    const memoizedEntries = useMemo(() => {
        return entries;
    },[entries]);

    const getOldEntries = useCallback(async () => {
        setIsLoadingTop(true);
        setLoadMoreTop(false);
        const data = await api.fetchEntries(leftOffTop, -1, query, 100, 3000);
        let leftOffTopBak = leftOffTop;
        setLeftOffTop(data.meta.leftOff);

        let scrollTo;
        if (data.meta.leftOff === 0) {
            setNoMoreDataTop(true);
            scrollTo = document.getElementById("noMoreDataTop");
        } else {
            scrollTo = document.getElementById(`entry-${leftOffTopBak}`);
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
        const newEntries = [...incomingEntries, ...entriesBuffer];
        setEntriesBuffer(newEntries);
        setEntries(newEntries);

        if (scrollTo) {
            scrollTo.scrollIntoView({block: "nearest", inline: "nearest"});
        }
    },[setLoadMoreTop, setIsLoadingTop, setEntries, entriesBuffer, setEntriesBuffer, query, setNoMoreDataTop, focusedEntryId, setFocusedEntryId, updateQuery, leftOffTop, setLeftOffTop]);

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
                    <ScrollableFeedVirtualized ref={scrollableRef} itemHeight={48} marginTop={10} onSnapBroken={onSnapBrokenEvent}>
                        {noMoreDataTop && <div id="noMoreDataTop" className={styles.noMoreDataAvailable}>No more data available</div>}
                        {memoizedEntries}
                    </ScrollableFeedVirtualized>
                    <button type="button"
                        className={`${styles.btnLive} ${isSnappedToBottom ? styles.hideButton : styles.showButton}`}
                        onClick={(_) => {
                            scrollableRef.current.jumpToBottom();
                            setIsSnappedToBottom(true);
                        }}>
                        <img alt="down" src={down} />
                    </button>
                </div>

                <div className={styles.footer}>
                    <div>Displaying <b>{entries?.length}</b> results (queried <b>{queriedCurrent}</b>/<b>{queriedTotal}</b>)</div>
                    {startTime !== 0 && <div>Started listening at <span style={{marginRight: 5, fontWeight: 600, fontSize: 13}}>{new Date(startTime).toLocaleString()}</span></div>}
                </div>
            </div>
    </>;
};
