import React, {useEffect, useState} from "react";
// import {HarFilters} from "./HarFilters";
import {HarEntriesList} from "./HarEntriesList";
import {makeStyles} from "@material-ui/core";
import "./style/HarPage.sass";
import styles from './style/HarEntriesList.module.sass';
import {HAREntryDetailed} from "./HarEntryDetailed";
import useWebSocket from 'react-use-websocket';

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

export const HarPage: React.FC = () => {

    const classes = useLayoutStyles();

    const [entries, setEntries] = useState([] as any);
    const [focusedEntryId, setFocusedEntryId] = useState(null);
    const [selectedHarEntry, setSelectedHarEntry] = useState(null);
    const [connectionOpen, setConnectionOpen] = useState(false);

    const socketUrl = 'ws://localhost:8899/ws';
    const {lastMessage} = useWebSocket(socketUrl, {
        onOpen: () => setConnectionOpen(true),
        onClose: () => setConnectionOpen(false),
        shouldReconnect: (closeEvent) => true});

    useEffect(() => {
        if(!lastMessage?.data) return;
        const entry = JSON.parse(lastMessage.data);
        if(!focusedEntryId) setFocusedEntryId(entry.id)
        let newEntries = [...entries];
        if(entries.length === 1000) {
            newEntries = newEntries.splice(1)
        }
        setEntries([...newEntries, entry])
    },[lastMessage?.data])

    useEffect(() => {
        if(!focusedEntryId) return;
        setSelectedHarEntry(null)
        fetch(`http://localhost:8899/api/entries/${focusedEntryId}`)
            .then(response => response.json())
            .then(data => setSelectedHarEntry(data));
    },[focusedEntryId])

    return (
        <div className="HarPage">
            <div style={{padding: "0 24px 24px 24px"}}>
                <div className="connectionText">
                    {connectionOpen ? "connected, waiting for traffic" : "not connected"}
                    <div className={connectionOpen ? "greenIndicator" : "redIndicator"}/>
                </div>
            </div>
            {entries.length > 0 && <div className="HarPage-Container">
                <div className="HarPage-ListContainer">
                    {/*<HarFilters />*/}
                    <div className={styles.container}>
                        <HarEntriesList entries={entries} focusedEntryId={focusedEntryId} setFocusedEntryId={setFocusedEntryId}/>
                    </div>
                </div>
                <div className={classes.details}>
                    {selectedHarEntry && <HAREntryDetailed harEntry={selectedHarEntry} classes={{root: classes.harViewer}}/>}
                </div>
            </div>}
        </div>
    )
};
