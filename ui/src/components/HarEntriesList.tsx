import {HarEntry} from "./HarEntry";
import React, {useEffect, useState} from "react";
import styles from './style/HarEntriesList.module.sass';
import spinner from './assets/spinner.svg';
import ScrollableFeed from "react-scrollable-feed";

interface HarEntriesListProps {
    entries: any[];
    setEntries: (entries: any[]) => void;
    focusedEntryId: string;
    setFocusedEntryId: (id: string) => void
}

export const HarEntriesList: React.FC<HarEntriesListProps> = ({entries, setEntries, focusedEntryId, setFocusedEntryId}) => {

    const [loadMore, setLoadMore] = useState(false);
    const [topEntryId, setTopEntryId] = useState(null);

    useEffect(() => {
        if(loadMore)
            fetchData();
        else {
            if(topEntryId) {
                const element = document.getElementById(topEntryId);
                if(element)
                    element.scrollIntoView();
            }
        }
    }, [loadMore]);


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

    function uuidv4() {
        return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
            var r = Math.random() * 16 | 0, v = c == 'x' ? r : (r & 0x3 | 0x8);
            return v.toString(16);
        });
    }

    const fetchData = () => {
        if(entries[5]) {
            setTopEntryId(entries[0].id);
            const temp = [...entries];
            temp.unshift({...entries[5], id: uuidv4()});
            temp.unshift({...entries[5], id: uuidv4()});
            temp.unshift({...entries[5], id: uuidv4()});
            temp.unshift({...entries[5], id: uuidv4()});
            temp.unshift({...entries[5], id: uuidv4()});
            temp.unshift({...entries[5], id: uuidv4()});
            temp.unshift({...entries[5], id: uuidv4()});

            setTimeout(() => {setEntries(temp); setLoadMore(false)}, 1000)
        }
    };

    return <>
            <div className={styles.list}>
                <div id="list" className={styles.list}>
                    {loadMore && <div style={{display: "flex", justifyContent: "center", marginBottom: 10}}><img alt="spinner" src={spinner} style={{height: 25}}/></div>}
                    <ScrollableFeed>
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
