import {HarEntry} from "./HarEntry";
import React, {useCallback, useEffect, useMemo, useState} from "react";
import styles from './style/HarEntriesList.module.sass';
import spinner from './assets/spinner.svg';
import ScrollableFeed from "react-scrollable-feed";
import {StatusType} from "./HarFilters";
import Api from "../helpers/api";

interface HarEntriesListProps {
    entries: any[];
    setEntries: (entries: any[]) => void;
    focusedEntryId: string;
    setFocusedEntryId: (id: string) => void;
    connectionOpen: boolean;
    noMoreDataTop: boolean;
    setNoMoreDataTop: (flag: boolean) => void;
    noMoreDataBottom: boolean;
    setNoMoreDataBottom: (flag: boolean) => void;
    methodsFilter: Array<string>;
    statusFilter: Array<string>;
    pathFilter: string
}

enum FetchOperator {
    LT = "lt",
    GT = "gt"
}

const api = new Api();

export const HarEntriesList: React.FC<HarEntriesListProps> = ({entries, setEntries, focusedEntryId, setFocusedEntryId, connectionOpen, noMoreDataTop, setNoMoreDataTop, noMoreDataBottom, setNoMoreDataBottom, methodsFilter, statusFilter, pathFilter}) => {

    const [loadMoreTop, setLoadMoreTop] = useState(false);
    const [isLoadingTop, setIsLoadingTop] = useState(false);

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
                <div id="list" className={styles.list}>
                    {isLoadingTop && <div className={styles.spinnerContainer}>
                        <img alt="spinner" src={spinner} style={{height: 25}}/>
                    </div>}
                    <ScrollableFeed>
                        {noMoreDataTop && !connectionOpen && <div id="noMoreDataTop" className={styles.noMoreDataAvailable}>No more data available</div>}
                        {filteredEntries.map(entry => <HarEntry key={entry.id}
                                                     entry={entry}
                                                     setFocusedEntryId={setFocusedEntryId}
                                                     isSelected={focusedEntryId === entry.id}/>)}
                        {!connectionOpen && !noMoreDataBottom && <div className={styles.fetchButtonContainer}>
                            <div className={styles.styledButton} onClick={() => getNewEntries()}>Fetch more entries</div>
                        </div>}
                    </ScrollableFeed>
                </div>

                {entries?.length > 0 && <div className={styles.footer}>
                    <div><b>{filteredEntries?.length !== entries.length && `${filteredEntries?.length} / `} {entries?.length}</b> requests</div>
                    <div>Started listening at <span style={{marginRight: 5, fontWeight: 600, fontSize: 13}}>{new Date(+entries[0].timestamp)?.toLocaleString()}</span></div>
                </div>}
            </div>
    </>;
};
