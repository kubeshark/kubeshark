import React, {useEffect, useRef, useState} from "react";
import {Filters} from "./Filters";
import {EntriesList} from "./EntriesList";
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
        background: variables.headerBackgoundColor,
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

    const [tappingStatus, setTappingStatus] = useState(null);

    const [disableScrollList, setDisableScrollList] = useState(false);

    const [query, setQueryDefault] = useState("");
    const [queryBackgroundColor, setQueryBackgroundColor] = useState("#f5f5f5");

    const [queriedCurrent, setQueriedCurrent] = useState(0);
    const [queriedTotal, setQueriedTotal] = useState(0);

    const [startTime, setStartTime] = useState(0);

    const setQuery = async (query) => {
        if (!query) {
            setQueryBackgroundColor("#f5f5f5")
        } else {
            const data = await api.validateQuery(query);
            if (data.valid) {
                setQueryBackgroundColor("#d2fad2")
            } else {
                setQueryBackgroundColor("#fad6dc")
            }
        }
        setQueryDefault(query)
    }

    const updateQuery = (addition) => {
        if (query) {
            setQuery(`${query} and ${addition}`)
        } else {
            setQuery(addition)
        }
    }

    const ws = useRef(null);

    const listEntry = useRef(null);

    const openWebSocket = (query) => {
        setEntries([])
        ws.current = new WebSocket(MizuWebsocketURL);
        ws.current.onopen = () => {
            ws.current.send(query)
            setConnection(ConnectionStatus.Connected);
        }
        ws.current.onclose = () => setConnection(ConnectionStatus.Closed);
    }

    if (ws.current) {
        ws.current.onmessage = e => {
            if (!e?.data) return;
            const message = JSON.parse(e.data);
            switch (message.messageType) {
                case "entry":
                    const entry = message.data
                    if (!focusedEntryId) setFocusedEntryId(entry.id.toString())
                    let newEntries = [...entries];
                    setEntries([...newEntries, entry])
                    if(listEntry.current) {
                        if(isScrollable(listEntry.current.firstChild)) {
                            setDisableScrollList(true)
                        }
                    }
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
                    setQueriedCurrent(message.data.current)
                    setQueriedTotal(message.data.total)
                    break;
                case "startTime":
                    setStartTime(message.data);
                    break;
                default:
                    console.error(`unsupported websocket message type, Got: ${message.messageType}`)
            }
        }
    }

    useEffect(() => {
        (async () => {
            openWebSocket("rlimit(100)");
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
        (async () => {
            try {
                const entryData = await api.getEntry(focusedEntryId);
                setSelectedEntryData(entryData);
            } catch (error) {
                console.error(error);
            }
        })()
    }, [focusedEntryId])

    const toggleConnection = () => {
        if (connection === ConnectionStatus.Connected) {
            ws.current.close();
        } else {
            openWebSocket(query);
            setConnection(ConnectionStatus.Connected);
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

    const onScrollEvent = (isAtBottom) => {
        isAtBottom ? setDisableScrollList(false) : setDisableScrollList(true)
    }

    const isScrollable = (element) => {
        return element.scrollHeight > element.clientHeight;
    };

    return (
        <div className="TrafficPage">
            <div className="TrafficPageHeader">
                {connection !== ConnectionStatus.Closed && <img style={{cursor: "pointer", marginRight: 15, height: 30}} alt="pause"
                     src={connection === ConnectionStatus.Connected ? pauseIcon : playIcon} onClick={toggleConnection}/>}
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
                            focusedEntryId={focusedEntryId}
                            setFocusedEntryId={setFocusedEntryId}
                            connectionOpen={connection === ConnectionStatus.Connected}
                            listEntryREF={listEntry}
                            onScrollEvent={onScrollEvent}
                            scrollableList={disableScrollList}
                            ws={ws.current}
                            openWebSocket={openWebSocket}
                            query={query}
                            updateQuery={updateQuery}
                            queriedCurrent={queriedCurrent}
                            queriedTotal={queriedTotal}
                            startTime={startTime}
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
