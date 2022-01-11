import React, {useCallback, useEffect, useMemo, useState} from "react";
import styles from './style/EntriesList.module.sass';
import ScrollableFeedVirtualized from "react-scrollable-feed-virtualized";
import Moment from 'moment';
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
    setQueriedTotal: any;
    startTime: number;
    noMoreDataTop: boolean;
    setNoMoreDataTop: (flag: boolean) => void;
    focusedEntryId: string;
    setFocusedEntryId: (id: string) => void;
    updateQuery: any;
    leftOffTop: number;
    setLeftOffTop: (leftOffTop: number) => void;
    isWebSocketConnectionClosed: boolean;
    ws: any;
    openWebSocket: (query: string, resetEntries: boolean) => void;
    leftOffBottom: number;
    truncatedTimestamp: number;
    setTruncatedTimestamp: any;
    scrollableRef: any;
}

const api = Api.getInstance();

export const EntriesList: React.FC<EntriesListProps> = ({entries, setEntries, query, listEntryREF, onSnapBrokenEvent, isSnappedToBottom, setIsSnappedToBottom, queriedCurrent, setQueriedCurrent, queriedTotal, setQueriedTotal, startTime, noMoreDataTop, setNoMoreDataTop, focusedEntryId, setFocusedEntryId, updateQuery, leftOffTop, setLeftOffTop, isWebSocketConnectionClosed, ws, openWebSocket, leftOffBottom, truncatedTimestamp, setTruncatedTimestamp, scrollableRef}) => {
    const [loadMoreTop, setLoadMoreTop] = useState(false);
    const [isLoadingTop, setIsLoadingTop] = useState(false);

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
        setLoadMoreTop(false);
        if (leftOffTop === null || leftOffTop <= 0) {
            return;
        }
        setIsLoadingTop(true);
        const data = await api.fetchEntries(leftOffTop, -1, query, 100, 3000);
        if (!data || data.data === null || data.meta === null) {
            setNoMoreDataTop(true);
            setIsLoadingTop(false);
            return;
        }
        setLeftOffTop(data.meta.leftOff);

        let scrollTo: boolean;
        if (data.meta.leftOff === 0) {
            setNoMoreDataTop(true);
            scrollTo = false;
        } else {
            scrollTo = true;
        }
        setIsLoadingTop(false);

        const newEntries = [...data.data.reverse(), ...entries];
        setEntries(newEntries);

        setQueriedCurrent(queriedCurrent + data.meta.current);
        setQueriedTotal(data.meta.total);
        setTruncatedTimestamp(data.meta.truncatedTimestamp);

        if (scrollTo) {
            scrollableRef.current.scrollToIndex(data.data.length - 1);
        }
    },[setLoadMoreTop, setIsLoadingTop, entries, setEntries, query, setNoMoreDataTop, leftOffTop, setLeftOffTop, queriedCurrent, setQueriedCurrent, setQueriedTotal, setTruncatedTimestamp, scrollableRef]);

    useEffect(() => {
        if(!isWebSocketConnectionClosed || !loadMoreTop || noMoreDataTop) return;
        getOldEntries();
    }, [loadMoreTop, noMoreDataTop, getOldEntries, isWebSocketConnectionClosed]);

    const scrollbarVisible = scrollableRef.current?.childWrapperRef.current.clientHeight > scrollableRef.current?.wrapperRef.current.clientHeight;

    return <>
            <div className={styles.list}>
                <div id="list" ref={listEntryREF} className={styles.list}>
                    {isLoadingTop && <div className={styles.spinnerContainer}>
                        <img alt="spinner" src={spinner} style={{height: 25}}/>
                    </div>}
                    {noMoreDataTop && <div id="noMoreDataTop" className={styles.noMoreDataAvailable}>No more data available</div>}
                    <ScrollableFeedVirtualized ref={scrollableRef} itemHeight={48} marginTop={10} onSnapBroken={onSnapBrokenEvent}>
                        {false /* It's because the first child is ignored by ScrollableFeedVirtualized */}
                        {memoizedEntries.map(entry => <EntryItem
                            key={`entry-${entry.id}`}
                            entry={entry}
                            focusedEntryId={focusedEntryId}
                            setFocusedEntryId={setFocusedEntryId}
                            style={{}}
                            updateQuery={updateQuery}
                            headingMode={false}
                        />)}
                    </ScrollableFeedVirtualized>
                    <button type="button"
                        title="Fetch old records"
                        className={`${styles.btnOld} ${!scrollbarVisible && leftOffTop > 0 ? styles.showButton : styles.hideButton}`}
                        onClick={(_) => {
                            ws.close();
                            getOldEntries();
                        }}>
                        <img alt="down" src={down} />
                    </button>
                    <button type="button"
                        title="Snap to bottom"
                        className={`${styles.btnLive} ${isSnappedToBottom && !isWebSocketConnectionClosed ? styles.hideButton : styles.showButton}`}
                        onClick={(_) => {
                            if (isWebSocketConnectionClosed) {
                                if (query) {
                                    openWebSocket(`(${query}) and leftOff(${leftOffBottom})`, false);
                                } else {
                                    openWebSocket(`leftOff(${leftOffBottom})`, false);
                                }
                            }
                            scrollableRef.current.jumpToBottom();
                            setIsSnappedToBottom(true);
                        }}>
                        <img alt="down" src={down} />
                    </button>
                </div>

                <div className={styles.footer}>
                    <div>Displaying <b>{entries?.length}</b> results out of <b>{queriedTotal}</b> total</div>
                    {startTime !== 0 && <div>Started listening at <span style={{marginRight: 5, fontWeight: 600, fontSize: 13}}>{Moment(truncatedTimestamp ? truncatedTimestamp : startTime).utc().format('MM/DD/YYYY, h:mm:ss.SSS A')}</span></div>}
                </div>
            </div>
    </>;
};
