import {EntryItem} from "./EntryListItem/EntryListItem";
import React, {useCallback, useEffect, useMemo, useRef, useState} from "react";
import styles from './style/EntriesList.module.sass';
import spinner from './assets/spinner.svg';
import ScrollableFeed from "react-scrollable-feed";
import {StatusType} from "./Filters";
import Api from "../helpers/api";
import down from "./assets/downImg.svg";

interface HarEntriesListProps {
    entries: any[];
    setEntries: (entries: any[]) => void;
    focusedEntry: any;
    setFocusedEntry: (entry: any) => void;
    connectionOpen: boolean;
    noMoreDataTop: boolean;
    setNoMoreDataTop: (flag: boolean) => void;
    noMoreDataBottom: boolean;
    setNoMoreDataBottom: (flag: boolean) => void;
    methodsFilter: Array<string>;
    statusFilter: Array<string>;
    pathFilter: string
    listEntryREF: any;
    onScrollEvent: (isAtBottom:boolean) => void;
    scrollableList: boolean;
}

enum FetchOperator {
    LT = "lt",
    GT = "gt"
}

const api = new Api();

export const EntriesList: React.FC<HarEntriesListProps> = ({entries, setEntries, focusedEntry, setFocusedEntry, connectionOpen, noMoreDataTop, setNoMoreDataTop, noMoreDataBottom, setNoMoreDataBottom, methodsFilter, statusFilter, pathFilter, listEntryREF, onScrollEvent, scrollableList}) => {

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

    const filterEntries = useCallback((entry) => {
        if(methodsFilter.length > 0 && !methodsFilter.includes(entry.method.toLowerCase())) return;
        if(pathFilter && entry.path?.toLowerCase()?.indexOf(pathFilter) === -1) return;
        if(statusFilter.includes(StatusType.SUCCESS) && entry.statusCode >= 400) return;
        if(statusFilter.includes(StatusType.ERROR) && entry.statusCode < 400) return;
        return entry;
    },[methodsFilter, pathFilter, statusFilter])

    const filteredEntries = useMemo(() => {
        return entries.filter(filterEntries);
    },[entries, filterEntries])

    const getOldEntries = useCallback(async () => {
        setIsLoadingTop(true);
        const data = await api.fetchEntries(FetchOperator.LT, entries[0].timestamp);
        setLoadMoreTop(false);

        let scrollTo;
        if(data.length === 0) {
            setNoMoreDataTop(true);
            scrollTo = document.getElementById("noMoreDataTop");
        } else {
            scrollTo = document.getElementById(filteredEntries?.[0]?.id);
        }
        setIsLoadingTop(false);
        const newEntries = [...data, ...entries];
        if(newEntries.length >= 1000) {
            newEntries.splice(1000);
        }
        setEntries(newEntries);

        if(scrollTo) {
            scrollTo.scrollIntoView();
        }
    },[setLoadMoreTop, setIsLoadingTop, entries, setEntries, filteredEntries, setNoMoreDataTop])

    useEffect(() => {
        if(!loadMoreTop || connectionOpen || noMoreDataTop) return;
        getOldEntries();
    }, [loadMoreTop, connectionOpen, noMoreDataTop, getOldEntries]);

    const getNewEntries = async () => {
        const data = await api.fetchEntries(FetchOperator.GT, entries[entries.length - 1].timestamp);
        let scrollTo;
        if(data.length === 0) {
            setNoMoreDataBottom(true);
        }
        scrollTo = document.getElementById(filteredEntries?.[filteredEntries.length -1]?.id);
        let newEntries = [...entries, ...data];
        if(newEntries.length >= 1000) {
            setNoMoreDataTop(false);
            newEntries = newEntries.slice(-1000);
        }
        setEntries(newEntries);
        if(scrollTo) {
            scrollTo.scrollIntoView({behavior: "smooth"});
        }
    }

    return <>
            <div className={styles.list}>
                <div id="list" ref={listEntryREF} className={styles.list} >
                    {isLoadingTop && <div className={styles.spinnerContainer}>
                        <img alt="spinner" src={spinner} style={{height: 25}}/>
                    </div>}
                    <ScrollableFeed ref={scrollableRef} onScroll={(isAtBottom) => onScrollEvent(isAtBottom)}>
                        {noMoreDataTop && !connectionOpen && <div id="noMoreDataTop" className={styles.noMoreDataAvailable}>No more data available</div>}
                        {filteredEntries.map(entry => <EntryItem key={entry.id}
                                                     entry={entry}
                                                     setFocusedEntry = {setFocusedEntry}
                                                     isSelected={focusedEntry.id === entry.id}/>)}
                        {!connectionOpen && !noMoreDataBottom && <div className={styles.fetchButtonContainer}>
                            <div className={styles.styledButton} onClick={() => getNewEntries()}>Fetch more entries</div>
                        </div>}
                    </ScrollableFeed>
                    <button type="button" 
                        className={`${styles.btnLive} ${scrollableList ? styles.showButton : styles.hideButton}`} 
                        onClick={(_) => scrollableRef.current.scrollToBottom()}>
                        <img alt="down" src={down} />
                    </button>
                </div>

                {entries?.length > 0 && <div className={styles.footer}>
                    <div><b>{filteredEntries?.length !== entries.length && `${filteredEntries?.length} / `} {entries?.length}</b> requests</div>
                    <div>Started listening at <span style={{marginRight: 5, fontWeight: 600, fontSize: 13}}>{new Date(+entries[0].timestamp)?.toLocaleString()}</span></div>
                </div>}
            </div>
    </>;
};
