import {HarEntry} from "./HarEntry";
import React, {useEffect, useState} from "react";
import styles from './style/HarEntriesList.module.sass';
import spinner from './assets/spinner.svg';
import ScrollableFeed from "react-scrollable-feed";

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
}

enum FetchOperator {
    LT = "lt",
    GT = "gt"
}

export const HarEntriesList: React.FC<HarEntriesListProps> = ({entries, setEntries, focusedEntryId, setFocusedEntryId, connectionOpen, noMoreDataTop, setNoMoreDataTop, noMoreDataBottom, setNoMoreDataBottom}) => {

    const [loadMoreTop, setLoadMoreTop] = useState(false);
    const [isLoadingTop, setIsLoadingTop] = useState(false);

    useEffect(() => {
        if(loadMoreTop && !connectionOpen && !noMoreDataTop)
            fetchData(FetchOperator.LT);
    }, [loadMoreTop, connectionOpen, noMoreDataTop]);

    useEffect(() => {
        const list = document.getElementById('list').firstElementChild;
        list.addEventListener('scroll', (e) => {
            const el: any = e.target;
            if(el.scrollTop === 0) {
                setLoadMoreTop(true);
            }
        });
    }, []);

    const fetchData = async (operator) => {

        const timestamp = operator === FetchOperator.LT ? entries[0].timestamp : entries[entries.length - 1].timestamp;
        if(operator === FetchOperator.LT)
            setIsLoadingTop(true);

        fetch(`http://localhost:8899/api/entries?limit=50&operator=${operator}&timestamp=${timestamp}`)
            .then(response => response.json())
            .then((data: any[]) => {
                let scrollTo;
                if(operator === FetchOperator.LT) {
                    if(data.length === 0) {
                        setNoMoreDataTop(true);
                        scrollTo = document.getElementById("noMoreDataTop");
                    } else {
                        scrollTo = document.getElementById(entries[0].id);
                    }
                    let newEntries = [...data, ...entries];
                    if(newEntries.length >= 1000) {
                        newEntries.splice(1000);
                    }
                    setEntries(newEntries);
                    setLoadMoreTop(false);
                    setIsLoadingTop(false)
                    if(scrollTo) {
                        scrollTo.scrollIntoView();
                    }
                }

                if(operator === FetchOperator.GT) {
                    if(data.length === 0) {
                        setNoMoreDataBottom(true);
                    }
                    scrollTo = document.getElementById(entries[entries.length -1].id);
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
            });
    };

    return <>
            <div className={styles.list}>
                <div id="list" className={styles.list}>
                    {isLoadingTop && <div style={{display: "flex", justifyContent: "center", marginBottom: 10}}><img alt="spinner" src={spinner} style={{height: 25}}/></div>}
                    <ScrollableFeed>
                        {noMoreDataTop && !connectionOpen && <div id="noMoreDataTop" style={{textAlign: "center", fontWeight: 600, color: "rgba(255,255,255,0.75)"}}>No more data available</div>}
                        {entries?.map(entry => <HarEntry key={entry.id}
                                                     entry={entry}
                                                     setFocusedEntryId={setFocusedEntryId}
                                                     isSelected={focusedEntryId === entry.id}/>)}
                        {!connectionOpen && !noMoreDataBottom && <div style={{width: "100%", display: "flex", justifyContent: "center", marginTop: 12, fontWeight: 600, color: "rgba(255,255,255,0.75)"}}>
                            <div className={styles.styledButton} onClick={() => fetchData(FetchOperator.GT)}>Fetch more entries</div>
                        </div>}
                    </ScrollableFeed>
                </div>

                {entries?.length > 0 && <div className={styles.footer}>
                    <div><b>{entries?.length}</b> requests</div>
                    <div>Started listening at <span style={{marginRight: 5, fontWeight: 600, fontSize: 13}}>{new Date(+entries[0].timestamp)?.toLocaleString()}</span></div>
                </div>}
            </div>
    </>;
};
