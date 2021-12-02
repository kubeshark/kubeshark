import React, {useEffect, useRef, useState} from "react";
import {Filters} from "./Filters";
import {EntriesList} from "./EntriesList";
import {EntryItem} from "./EntryListItem/EntryListItem";
import {makeStyles} from "@material-ui/core";
import "./style/TrafficPage.sass";
import styles from './style/EntriesList.module.sass';
import {EntryDetailed} from "./EntryDetailed";
import playIcon from './assets/run.svg';
import pauseIcon from './assets/pause.svg';
import variables from '../variables.module.scss';
import {StatusBar} from "./UI/StatusBar";
import Api, {MizuWebsocketURL} from "../helpers/api";
import { ToastContainer, toast } from 'react-toastify';
import 'react-toastify/dist/ReactToastify.css';

const useLayoutStyles = makeStyles(() => ({
    details: {
        flex: "0 0 50%",
        width: "45vw",
        padding: "12px 24px",
        borderRadius: 4,
        marginTop: 15,
        background: variables.headerBackgroundColor,
    },

    viewer: {
        display: 'flex',
        overflowY: 'auto',
        height: "calc(100% - 70px)",
        padding: 5,
        paddingBottom: 0,
        overflow: "auto",
    }
}));

enum ConnectionStatus {
    Closed,
    Connected,
}

interface TrafficPageProps {
    setAnalyzeStatus: (status: any) => void;
    onTLSDetected: (destAddress: string) => void;
}

const api = new Api();

export const TrafficPage: React.FC<TrafficPageProps> = ({setAnalyzeStatus, onTLSDetected}) => {

    const classes = useLayoutStyles();

    const [entries, setEntries] = useState([] as any);
    const [focusedEntryId, setFocusedEntryId] = useState(null);
    const [selectedEntryData, setSelectedEntryData] = useState(null);
    const [connection, setConnection] = useState(ConnectionStatus.Closed);

    const [noMoreDataTop, setNoMoreDataTop] = useState(false);

    const [tappingStatus, setTappingStatus] = useState(null);

    const [isSnappedToBottom, setIsSnappedToBottom] = useState(true);

    const [query, setQuery] = useState("");
    const [queryBackgroundColor, setQueryBackgroundColor] = useState("#f5f5f5");
    const [addition, updateQuery] = useState("");

    const [queriedCurrent, setQueriedCurrent] = useState(0);
    const [queriedTotal, setQueriedTotal] = useState(0);
    const [leftOffBottom, setLeftOffBottom] = useState(0);
    const [leftOffTop, setLeftOffTop] = useState(null);

    const [startTime, setStartTime] = useState(0);

    useEffect(() => {
        (async function() {
            if (!query) {
                setQueryBackgroundColor("#f5f5f5")
            } else {
                const data = await api.validateQuery(query);
                if (!data) {
                    return;
                }
                if (data.valid) {
                    setQueryBackgroundColor("#d2fad2");
                } else {
                    setQueryBackgroundColor("#fad6dc");
                }
            }
        })();
    }, [query]);

    useEffect(() => {
        if (query) {
            setQuery(`${query} and ${addition}`);
        } else {
            setQuery(addition);
        }
        // eslint-disable-next-line
    }, [addition]);

    const ws = useRef(null);

    const listEntry = useRef(null);

    const openWebSocket = (query: string, resetEntries: boolean) => {
        if (resetEntries) {
            setFocusedEntryId(null);
            setEntries([]);
        }
        ws.current = new WebSocket(MizuWebsocketURL);
        ws.current.onopen = () => {
            setConnection(ConnectionStatus.Connected);
            ws.current.send(query);
        }
        ws.current.onclose = () => {
            setConnection(ConnectionStatus.Closed);
        }
        ws.current.onerror = (event) => {
            console.error("WebSocket error:", event);
            if (query) {
                openWebSocket(`(${query}) and leftOff(${leftOffBottom})`, false);
            } else {
                openWebSocket(`leftOff(${leftOffBottom})`, false);
            }
        }
    }

    if (ws.current) {
        ws.current.onmessage = e => {
            if (!e?.data) return;
            const message = JSON.parse(e.data);
            switch (message.messageType) {
                case "entry":
                    const entry = message.data;
                    var focusThis = false;
                    if (!focusedEntryId) {
                        focusThis = true;
                        setFocusedEntryId(entry.id.toString());
                    }
                    setEntries([
                        ...entries,
                        <EntryItem
                            key={entry.id}
                            entry={entry}
                            focusedEntryId={focusThis ? entry.id.toString() : focusedEntryId}
                            setFocusedEntryId={setFocusedEntryId}
                            style={{}}
                            updateQuery={updateQuery}
                            headingMode={false}
                        />
                    ]);
                    break
                case "status":
                    setTappingStatus(message.tappingStatus);
                    break
                case "analyzeStatus":
                    setAnalyzeStatus(message.analyzeStatus);
                    break
                case "outboundLink":
                    onTLSDetected(message.Data.DstIP);
                    break;
                case "toast":
                    toast[message.data.type](message.data.text, {
                        position: "bottom-right",
                        theme: "colored",
                        autoClose: message.data.autoClose,
                        hideProgressBar: false,
                        closeOnClick: true,
                        pauseOnHover: true,
                        draggable: true,
                        progress: undefined,
                    });
                    break;
                case "queryMetadata":
                    setQueriedCurrent(message.data.current);
                    setQueriedTotal(message.data.total);
                    setLeftOffBottom(message.data.leftOff);
                    if (leftOffTop === null) {
                        setLeftOffTop(message.data.leftOff);
                    }
                    break;
                case "startTime":
                    setStartTime(message.data);
                    break;
                case "focusEntry":
                    // To achieve selecting only one entry, render all elements in the buffer
                    // with the current `focusedEntryId` value.
                    entries.forEach((entry: any, i: number) => {
                        entries[i] = React.cloneElement(entry, {
                            focusedEntryId: focusedEntryId
                        });
                    });
                    break;
                default:
                    console.error(`unsupported websocket message type, Got: ${message.messageType}`)
            }
        }
    }

    useEffect(() => {
        (async () => {
            openWebSocket("leftOff(-1)", true);
            try{
                const tapStatusResponse = await api.tapStatus();
                setTappingStatus(tapStatusResponse);
                const analyzeStatusResponse = await api.analyzeStatus();
                setAnalyzeStatus(analyzeStatusResponse);
            } catch (error) {
                console.error(error);
            }
        })()
        // eslint-disable-next-line
    }, []);


    useEffect(() => {
        if (!focusedEntryId) return;
        setSelectedEntryData(null);

        if (ws.current.readyState === WebSocket.OPEN) {
            ws.current.send(focusedEntryId);
        } else {
            entries.forEach((entry: any, i: number) => {
                entries[i] = React.cloneElement(entry, {
                    focusedEntryId: focusedEntryId
                });
            });
        }

        (async () => {
            try {
                const entryData = await api.getEntry(focusedEntryId);
                setSelectedEntryData(entryData);
            } catch (error) {
                if (error.response) {
                    toast[error.response.data.type](`Entry[${focusedEntryId}]: ${error.response.data.msg}`, {
                        position: "bottom-right",
                        theme: "colored",
                        autoClose: error.response.data.autoClose,
                        hideProgressBar: false,
                        closeOnClick: true,
                        pauseOnHover: true,
                        draggable: true,
                        progress: undefined,
                    });
                }
                console.error(error);
            }
        })();
        // eslint-disable-next-line
    }, [focusedEntryId])

    const toggleConnection = () => {
        if (connection === ConnectionStatus.Connected) {
            ws.current.close();
        } else {
            if (query) {
                openWebSocket(`(${query}) and leftOff(${leftOffBottom})`, false);
            } else {
                openWebSocket(`leftOff(${leftOffBottom})`, false);
            }
        }
    }

    const getConnectionStatusClass = (isContainer) => {
        const container = isContainer ? "Container" : "";
        switch (connection) {
            case ConnectionStatus.Connected:
                return "greenIndicator" + container;
            default:
                return "redIndicator" + container;
        }
    }

    const getConnectionTitle = () => {
        switch (connection) {
            case ConnectionStatus.Connected:
                return "connected, waiting for traffic"
            default:
                return "not connected";
        }
    }

    const onSnapBrokenEvent = () => {
        setIsSnappedToBottom(false)
    }

    return (
        <div className="TrafficPage">
            <div className="TrafficPageHeader">
                <img className="playPauseIcon" style={{visibility: connection === ConnectionStatus.Connected ? "visible" : "hidden"}} alt="pause"
                    src={pauseIcon} onClick={toggleConnection}/>
                <img className="playPauseIcon" style={{position: "absolute", visibility: connection === ConnectionStatus.Connected ? "hidden" : "visible"}} alt="play"
                    src={playIcon} onClick={toggleConnection}/>
                <div className="connectionText">
                    {getConnectionTitle()}
                    <div className={"indicatorContainer " + getConnectionStatusClass(true)}>
                        <div className={"indicator " + getConnectionStatusClass(false)}/>
                    </div>
                </div>
            </div>
            {<div className="TrafficPage-Container">
                <div className="TrafficPage-ListContainer">
                    <Filters
                        query={query}
                        setQuery={setQuery}
                        backgroundColor={queryBackgroundColor}
                        ws={ws.current}
                        openWebSocket={openWebSocket}
                    />
                    <div className={styles.container}>
                        <EntriesList
                            entries={entries}
                            setEntries={setEntries}
                            query={query}
                            listEntryREF={listEntry}
                            onSnapBrokenEvent={onSnapBrokenEvent}
                            isSnappedToBottom={isSnappedToBottom}
                            setIsSnappedToBottom={setIsSnappedToBottom}
                            queriedCurrent={queriedCurrent}
                            queriedTotal={queriedTotal}
                            startTime={startTime}
                            noMoreDataTop={noMoreDataTop}
                            setNoMoreDataTop={setNoMoreDataTop}
                            focusedEntryId={focusedEntryId}
                            setFocusedEntryId={setFocusedEntryId}
                            updateQuery={updateQuery}
                            leftOffTop={leftOffTop}
                            setLeftOffTop={setLeftOffTop}
                        />
                    </div>
                </div>
                <div className={classes.details}>
                    {selectedEntryData && <EntryDetailed entryData={selectedEntryData} updateQuery={updateQuery}/>}
                </div>
            </div>}
            {tappingStatus?.pods != null && <StatusBar tappingStatus={tappingStatus}/>}
            <ToastContainer
                position="bottom-right"
                autoClose={5000}
                hideProgressBar={false}
                newestOnTop={false}
                closeOnClick
                rtl={false}
                pauseOnFocusLoss
                draggable
                pauseOnHover
            />
        </div>
    )
};
