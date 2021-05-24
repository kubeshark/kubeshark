import React, {useEffect, useRef, useState} from "react";
import {HarFilters} from "./HarFilters";
import {HarEntriesList} from "./HarEntriesList";
import {makeStyles} from "@material-ui/core";
import "./style/HarPage.sass";
import styles from './style/HarEntriesList.module.sass';
import {HAREntryDetailed} from "./HarEntryDetailed";
import playIcon from './assets/play.svg';
import pauseIcon from './assets/pause.svg';
import variables from './style/variables.module.scss';
import {StatusBar} from "./StatusBar";

const useLayoutStyles = makeStyles(() => ({
    details: {
        flex: "0 0 50%",
        width: "45vw",
        padding: "12px 24px",
        borderRadius: 4,
        marginTop: 15,
        background: variables.headerBackgoundColor
    },

    harViewer: {
        display: 'flex',
        overflowY: 'auto',
        height: "calc(100% - 70px)",
        padding: 5,
        paddingBottom: 0
    }
}));

enum ConnectionStatus {
    Closed,
    Connected,
    Paused
}

export const HarPage: React.FC = () => {

    const classes = useLayoutStyles();

    const [entries, setEntries] = useState([] as any);
    const [focusedEntryId, setFocusedEntryId] = useState(null);
    const [selectedHarEntry, setSelectedHarEntry] = useState(null);
    const [connection, setConnection] = useState(ConnectionStatus.Closed);
    const [noMoreDataTop, setNoMoreDataTop] = useState(false);
    const [noMoreDataBottom, setNoMoreDataBottom] = useState(false);

    const [methodsFilter, setMethodsFilter] = useState([]);
    const [statusFilter, setStatusFilter] = useState([]);
    const [pathFilter, setPathFilter] = useState("");

    const [apiStatus, setApiStatus] = useState(null);

    const ws = useRef(null);

    const openWebSocket = () => {
        ws.current = new WebSocket("ws://localhost:8899/ws");
        ws.current.onopen = () => setConnection(ConnectionStatus.Connected);
        ws.current.onclose = () => setConnection(ConnectionStatus.Closed);
    }

    if(ws.current) {
        ws.current.onmessage = e => {
            if(!e?.data) return;
            const message = JSON.parse(e.data);

            switch (message.messageType) {
                case "entry":
                    const entry = message.data
                    if(connection === ConnectionStatus.Paused) {
                        setNoMoreDataBottom(false)
                        return;
                    }
                    if(!focusedEntryId) setFocusedEntryId(entry.id)
                    let newEntries = [...entries];
                    if(entries.length === 1000) {
                        newEntries = newEntries.splice(1);
                        setNoMoreDataTop(false);
                    }
                    setEntries([...newEntries, entry])
                    break
                case "status":
                    console.log('status update', message);
                    setApiStatus(message.tappingStatus);
                    break
                default:
                    console.error(`unsupported websocket message type, Got: ${message.messageType}`)
            }
        }
    }

    useEffect(() => {
        openWebSocket();
        fetch(`http://localhost:8899/api/tapStatus`)
            .then(response => response.json())
            .then(data => setApiStatus(data));
    }, []);


    useEffect(() => {
        if(!focusedEntryId) return;
        setSelectedHarEntry(null)
        fetch(`http://localhost:8899/api/entries/${focusedEntryId}`)
            .then(response => response.json())
            .then(data => setSelectedHarEntry(data));
    },[focusedEntryId])

    const toggleConnection = () => {
        setConnection(connection === ConnectionStatus.Connected ? ConnectionStatus.Paused : ConnectionStatus.Connected );
    }

    const getConnectionStatusClass = (isContainer) => {
        const container = isContainer ? "Container" : "";
        switch (connection) {
            case ConnectionStatus.Paused:
                return "orangeIndicator" + container;
            case ConnectionStatus.Connected:
                return "greenIndicator" + container;
            default:
                return "redIndicator" + container;
        }
    }

    const getConnectionTitle = () => {
        switch (connection) {
            case ConnectionStatus.Paused:
                return "traffic paused";
            case ConnectionStatus.Connected:
                return "connected, waiting for traffic"
            default:
                return "not connected";
        }
    }

    return (
        <div className="HarPage">
            <div className="harPageHeader">
                <img style={{cursor: "pointer", marginRight: 15, height: 30}} alt="pause" src={connection === ConnectionStatus.Connected ? pauseIcon : playIcon} onClick={toggleConnection}/>
                <div className="connectionText">
                    {getConnectionTitle()}
                    <div className={"indicatorContainer " + getConnectionStatusClass(true)}>
                        <div className={"indicator " +  getConnectionStatusClass(false)} />
                    </div>
                </div>
            </div>
            {entries.length > 0 && <div className="HarPage-Container">
                <div className="HarPage-ListContainer">
                    <HarFilters methodsFilter={methodsFilter}
                                setMethodsFilter={setMethodsFilter}
                                statusFilter={statusFilter}
                                setStatusFilter={setStatusFilter}
                                pathFilter={pathFilter}
                                setPathFilter={setPathFilter}
                    />
                    <div className={styles.container}>
                        <HarEntriesList entries={entries}
                                        setEntries={setEntries}
                                        focusedEntryId={focusedEntryId}
                                        setFocusedEntryId={setFocusedEntryId}
                                        connectionOpen={connection === ConnectionStatus.Connected}
                                        noMoreDataBottom={noMoreDataBottom}
                                        setNoMoreDataBottom={setNoMoreDataBottom}
                                        noMoreDataTop={noMoreDataTop}
                                        setNoMoreDataTop={setNoMoreDataTop}
                                        methodsFilter={methodsFilter}
                                        statusFilter={statusFilter}
                                        pathFilter={pathFilter}
                        />
                    </div>
                </div>
                <div className={classes.details}>
                    {selectedHarEntry && <HAREntryDetailed harEntry={selectedHarEntry} classes={{root: classes.harViewer}}/>}
                </div>
            </div>}
            {apiStatus && <StatusBar tappingStatus={apiStatus}/>}
        </div>
    )
};
