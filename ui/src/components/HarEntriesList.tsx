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
}

export const HarEntriesList: React.FC<HarEntriesListProps> = ({entries, setEntries, focusedEntryId, setFocusedEntryId, connectionOpen}) => {

    const [loadMore, setLoadMore] = useState(false);
    const [isLoading, setIsLoading] = useState(false);
    const [topEntry, setTopEntry] = useState(null);
    const [noMoreData, setNoMoreData] = useState(false);

    useEffect(() => {
        if(loadMore && !connectionOpen && !noMoreData)
            fetchData();
    }, [loadMore, connectionOpen, noMoreData]);

    useEffect(() => {
        const element = noMoreData ? document.getElementById("noMoreData") :  document.getElementById(topEntry?.id);
        if(element)
            element.scrollIntoView();
    },[loadMore])


    useEffect(() => {
        const list = document.getElementById('list').firstElementChild;
        list.addEventListener('scroll', (e) => {
            const el: any = e.target;
            // console.log(el.scrollTop);
            // console.log(el.clientHeight);
            // console.log(el.scrollHeight);
            // if(el.scrollTop + el.clientHeight === el.scrollHeight) { // scroll down
            //     setLoadMore(true);
            // }
            if(el.scrollTop === 0) {
                setLoadMore(true);
            }
        });
    }, []);

    const fetchData = () => {
        setTopEntry(entries[0]);
        setIsLoading(true);
        fetch(`http://localhost:8899/api/entries?limit=20&operator=lt&timestamp=${entries[0].timestamp}`)
            .then(response => response.json())
            .then((data: any[]) => {
                if(data.length === 0) {
                    setNoMoreData(true);
                }
                setEntries([...data, ...entries]);
                setLoadMore(false);
                setIsLoading(false)
            });
    };

    return <>
            <div className={styles.list}>
                <div id="list" className={styles.list}>
                    {isLoading && <div style={{display: "flex", justifyContent: "center", marginBottom: 10}}><img alt="spinner" src={spinner} style={{height: 25}}/></div>}
                    <ScrollableFeed>
                        {noMoreData && !connectionOpen && <div id="noMoreData" style={{textAlign: "center", fontWeight: 600, color: "rgba(255,255,255,0.75)"}}>No more data available</div>}
                        {entries?.map(entry => <HarEntry key={entry.id}
                                                     entry={entry}
                                                     setFocusedEntryId={setFocusedEntryId}
                                                     isSelected={focusedEntryId === entry.id}/>)}
                    </ScrollableFeed>

                </div>

                {entries?.length > 0 && <div className={styles.footer}>
                    <div><b>{entries?.length}</b> requests</div>
                    <div>Started listening at <span style={{marginRight: 5, fontWeight: 600, fontSize: 13}}>{new Date(+entries[0].timestamp*1000)?.toLocaleString()}</span></div>
                </div>}
            </div>
    </>;
};
