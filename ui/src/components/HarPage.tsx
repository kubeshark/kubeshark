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
        backgroundColor: "#171c30",
        padding: "12px 24px",
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

    const socketUrl = 'ws://localhost:8899/ws';
    const {lastMessage} = useWebSocket(socketUrl, {shouldReconnect: (closeEvent) => true});

    useEffect(() => {
        if(!lastMessage?.data) return;
        const entry = JSON.parse(lastMessage.data);
        if(!focusedEntryId) setFocusedEntryId(entry.id)
        setEntries([...entries, entry])
    },[lastMessage?.data])

    useEffect(() => {
        if(!focusedEntryId) return;
        fetch(`http://localhost:8899/api/entries/${focusedEntryId}`)
            .then(response => response.json())
            .then(data => setSelectedHarEntry(data));
    },[focusedEntryId])

    return (
        <div className="HarPage">
            <div className="HarPage-Container">
                <div className="HarPage-ListContainer">
                    {/*<HarFilters />*/}
                    <div className={styles.container}>
                        <HarEntriesList entries={entries} focusedEntryId={focusedEntryId} setFocusedEntryId={setFocusedEntryId}/>
                    </div>
                </div>
                {selectedHarEntry && <div className={classes.details}>
                    <HAREntryDetailed harEntry={selectedHarEntry} classes={{root: classes.harViewer}}/>
                </div>}
            </div>
        </div>
    )
};
