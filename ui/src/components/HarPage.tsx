import React, {useEffect, useRef, useState} from "react";
// import {HarFilters} from "./HarFilters";
import {HarEntriesList} from "./HarEntriesList";
import {makeStyles} from "@material-ui/core";
import "./style/HarPage.sass";
import styles from './style/HarEntriesList.module.sass';
import {HAREntryDetailed} from "./HarEntryDetailed";
import playIcon from './assets/play.svg';
import pauseIcon from './assets/pause.svg';

const useLayoutStyles = makeStyles(() => ({
    details: {
        flex: "0 0 50%",
        width: "45vw",
        padding: "12px 24px",
        backgroundColor: "#090b14",
        borderLeft: "2px #11162a solid"
    },

    harViewer: {
        display: 'flex',
        overflowY: 'auto',
        height: "calc(100% - 58px)",
        padding: 5,
        paddingBottom: 0
    }
}));

enum ConnectionStatus {
    Closed,
    Connection,
    Paused
}

export const HarPage: React.FC = () => {

    const classes = useLayoutStyles();

    const [entries, setEntries] = useState([] as any);
    const [focusedEntryId, setFocusedEntryId] = useState(null);
    const [selectedHarEntry, setSelectedHarEntry] = useState(null);
    const [connectionOpen, setConnectionOpen] = useState(false);
    const [noMoreDataTop, setNoMoreDataTop] = useState(false);

    const ws = useRef(null);

    const openWebSocket = () => {
        ws.current = new WebSocket("ws://localhost:8899/ws");
        ws.current.onopen = () => setConnectionOpen(true);
        ws.current.onclose = () => setConnectionOpen(false);
    }

    if(ws.current) {
        ws.current.onmessage = e => {
            if(!e?.data) return;
            const entry = JSON.parse(e.data);
            if(!focusedEntryId) setFocusedEntryId(entry.id)
            let newEntries = [...entries];
            if(entries.length === 1000) {
                newEntries = newEntries.splice(1);
                setNoMoreDataTop(false);
            }
            setEntries([...newEntries, entry])
        }
    }

    useEffect(() => {
        openWebSocket();
    }, []);


    useEffect(() => {
        if(!focusedEntryId) return;
        setSelectedHarEntry(null)
        fetch(`http://localhost:8899/api/entries/${focusedEntryId}`)
            .then(response => response.json())
            .then(data => setSelectedHarEntry(data));
    },[focusedEntryId])

    const toggleConnection = () => {
        if(connectionOpen) {
            ws.current.close();
        } else {
            openWebSocket();
        }
    }

    return (
        <div className="HarPage">
            <div style={{padding: "0 24px 24px 24px", display: "flex", alignItems: "center"}}>
                <img style={{cursor: "pointer", marginRight: 15, height: 20}} alt="pause" src={connectionOpen ? pauseIcon : playIcon} onClick={toggleConnection}/>
                <div className="connectionText">
                    {connectionOpen ? "connected, waiting for traffic" : "not connected"}
                    <div className={connectionOpen ? "greenIndicator" : "redIndicator"}/>
                </div>
            </div>
            {entries.length > 0 && <div className="HarPage-Container">
                <div className="HarPage-ListContainer">
                    {/*<HarFilters />*/}
                    <div className={styles.container}>
                        <HarEntriesList entries={entries} setEntries={setEntries} focusedEntryId={focusedEntryId} setFocusedEntryId={setFocusedEntryId} connectionOpen={connectionOpen} noMoreDataTop={noMoreDataTop} setNoMoreDataTop={setNoMoreDataTop}/>
                    </div>
                </div>
                <div className={classes.details}>
                    {selectedHarEntry && <HAREntryDetailed harEntry={selectedHarEntry} classes={{root: classes.harViewer}}/>}
                </div>
            </div>}
        </div>
    )
};
