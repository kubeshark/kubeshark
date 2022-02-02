import React, {useEffect, useMemo} from "react";
import styles from './style/EntriesList.module.sass';
import ScrollableFeedVirtualized from "react-scrollable-feed-virtualized";
import Moment from 'moment';
import {EntryItem} from "./EntryListItem/EntryListItem";
import down from "./assets/downImg.svg";
import spinner from './assets/spinner.svg';
import {useRecoilState, useRecoilValue} from "recoil";
import entriesAtom from "../recoil/entries";
import wsConnectionAtom, {WsConnectionStatus} from "../recoil/wsConnection";
import queryAtom from "../recoil/query";

interface EntriesListProps {
    listEntryREF: any;
    onSnapBrokenEvent: () => void;
    isSnappedToBottom: boolean;
    setIsSnappedToBottom: any;
    queriedTotal: number;
    startTime: number;
    noMoreDataTop: boolean;
    setNoMoreDataTop: (flag: boolean) => void;
    leftOffTop: number;
    setLeftOffTop: (leftOffTop: number) => void;
    ws: any;
    openWebSocket: (query: string, resetEntries: boolean) => void;
    leftOffBottom: number;
    truncatedTimestamp: number;
    scrollableRef: any;
    loadMoreTop: boolean;
    setLoadMoreTop: any;
    isLoadingTop: boolean;
    getOldEntries: () => void;
}

export const EntriesList: React.FC<EntriesListProps> = ({listEntryREF, onSnapBrokenEvent, isSnappedToBottom, setIsSnappedToBottom, queriedTotal, startTime, noMoreDataTop, setNoMoreDataTop, leftOffTop, setLeftOffTop, ws, openWebSocket, leftOffBottom, truncatedTimestamp, scrollableRef, loadMoreTop, setLoadMoreTop, isLoadingTop, getOldEntries}) => {

    const [entries] = useRecoilState(entriesAtom);
    const wsConnection = useRecoilValue(wsConnectionAtom);
    const query = useRecoilValue(queryAtom);
    const isWsConnectionClosed = wsConnection === WsConnectionStatus.Closed;

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

    useEffect(() => {
        if(!isWsConnectionClosed || !loadMoreTop || noMoreDataTop) return;
        getOldEntries();
    }, [loadMoreTop, noMoreDataTop, getOldEntries, isWsConnectionClosed]);

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
                            style={{}}
                            headingMode={false}
                        />)}
                    </ScrollableFeedVirtualized>
                    <button type="button"
                        title="Fetch old records"
                        className={`${styles.btnOld} ${!scrollbarVisible}`}
                        onClick={(_) => {
                            ws.close();
                            getOldEntries();
                        }}>
                        <img alt="down" src={down} />
                    </button>
                    <button type="button"
                        title="Snap to bottom"
                        className={`${styles.btnLive} ${isSnappedToBottom && !isWsConnectionClosed ? styles.hideButton : styles.showButton}`}
                        onClick={(_) => {
                            if (isWsConnectionClosed) {
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
                    <div>Displaying <b id="entries-length">{entries?.length}</b> results out of <b id="total-entries">{queriedTotal}</b> total</div>
                    {startTime !== 0 && <div>Started listening at <span style={{marginRight: 5, fontWeight: 600, fontSize: 13}}>{Moment(truncatedTimestamp ? truncatedTimestamp : startTime).utc().format('MM/DD/YYYY, h:mm:ss.SSS A')}</span></div>}
                </div>
            </div>
    </>;
};
