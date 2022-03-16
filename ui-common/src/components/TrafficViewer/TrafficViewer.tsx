import React, { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { Filters } from "./Filters";
import { EntriesList } from "./EntriesList";
import { makeStyles } from "@material-ui/core";
import TrafficViewerStyles from "./TrafficViewer.module.sass";
import styles from '../style/EntriesList.module.sass';
import { EntryDetailed } from "./EntryDetailed";
import playIcon from 'assets/run.svg';
import pauseIcon from 'assets/pause.svg';
import variables from '../../variables.module.scss';
import { toast } from 'react-toastify';
import debounce from 'lodash/debounce';
import { RecoilRoot, RecoilState, useRecoilState, useRecoilValue, useSetRecoilState } from "recoil";
import entriesAtom from "../../recoil/entries";
import focusedEntryIdAtom from "../../recoil/focusedEntryId";
import websocketConnectionAtom, { WsConnectionStatus } from "../../recoil/wsConnection";
import queryAtom from "../../recoil/query";
import { TLSWarning } from "../TLSWarning/TLSWarning";
import trafficViewerApiAtom from "../../recoil/TrafficViewerApi"
import TrafficViewerApi from "./TrafficViewerApi";
import { StatusBar } from "../UI/StatusBar";
import tappingStatusAtom from "../../recoil/tappingStatus/atom";


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
    display: "flex",
    overflowY: "auto",
    height: "calc(100% - 70px)",
    padding: 5,
    paddingBottom: 0,
    overflow: "auto",
  },
}));

interface TrafficViewerProps {
  setAnalyzeStatus?: (status: any) => void;
  api?: any
  message?: {}
  error?: {}
  isWebSocketOpen: boolean
  trafficViewerApiProp: TrafficViewerApi,
  actionButtons?: JSX.Element,
  isShowStatusBar?: boolean
}

const TrafficViewer: React.FC<TrafficViewerProps> = ({ setAnalyzeStatus, message, error, isWebSocketOpen, trafficViewerApiProp, actionButtons, isShowStatusBar }) => {
  const classes = useLayoutStyles();

  const [entries, setEntries] = useRecoilState(entriesAtom);
  const [focusedEntryId, setFocusedEntryId] = useRecoilState(focusedEntryIdAtom);
  const [wsConnection, setWsConnection] = useRecoilState(websocketConnectionAtom);
  const query = useRecoilValue(queryAtom);
  const [queryToSend, setQueryToSend] = useState("")
  const setTrafficViewerApiState = useSetRecoilState(trafficViewerApiAtom as RecoilState<TrafficViewerApi>)
  const [tappingStatus, setTappingStatus] = useRecoilState(tappingStatusAtom);


  const [noMoreDataTop, setNoMoreDataTop] = useState(false);
  const [isSnappedToBottom, setIsSnappedToBottom] = useState(true);

  const [queryBackgroundColor, setQueryBackgroundColor] = useState("#f5f5f5");

  const [queriedCurrent, setQueriedCurrent] = useState(0);
  const [queriedTotal, setQueriedTotal] = useState(0);
  const [leftOffBottom, setLeftOffBottom] = useState(0);
  const [leftOffTop, setLeftOffTop] = useState(null);
  const [truncatedTimestamp, setTruncatedTimestamp] = useState(0);

  const [startTime, setStartTime] = useState(0);
  const scrollableRef = useRef(null);

  const [showTLSWarning, setShowTLSWarning] = useState(false);
  const [userDismissedTLSWarning, setUserDismissedTLSWarning] = useState(false);
  const [addressesWithTLS, setAddressesWithTLS] = useState(new Set<string>());

  const handleQueryChange = useMemo(
    () =>
      debounce(async (query: string) => {
        if (!query) {
          setQueryBackgroundColor("#f5f5f5");
        } else {
          const data = await trafficViewerApiProp.validateQuery(query);
          if (!data) {
            return;
          }
          if (data.valid) {
            setQueryBackgroundColor("#d2fad2");
          } else {
            setQueryBackgroundColor("#fad6dc");
          }
        }
      }, 500),
    []
  ) as (query: string) => void;

  useEffect(() => {
    handleQueryChange(query);
  }, [query, handleQueryChange]);


  const listEntry = useRef(null);
  const openWebSocket = (query: string, resetEntries: boolean) => {
    if (resetEntries) {
      setFocusedEntryId(null);
      setEntries([]);
      setQueriedCurrent(0);
      setLeftOffTop(null);
      setNoMoreDataTop(false);
    }
    setQueryToSend(query)
    trafficViewerApiProp.webSocket.open();
  }

  const onmessage = useCallback((e) => {
    if (!e?.data) return;
    const message = JSON.parse(e.data);
    switch (message.messageType) {
      case "entry":
        const entry = message.data;
        if (!focusedEntryId) setFocusedEntryId(entry.id.toString());
        const newEntries = [...entries, entry];
        if (newEntries.length === 10001) {
          setLeftOffTop(newEntries[0].entry.id);
          newEntries.shift();
          setNoMoreDataTop(false);
        }
        setEntries(newEntries);
        break;
      case "status":
        setTappingStatus(message.tappingStatus);
        break;
      case "analyzeStatus":
        setAnalyzeStatus(message.analyzeStatus);
        break;
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
        setQueriedCurrent(queriedCurrent + message.data.current);
        setQueriedTotal(message.data.total);
        setLeftOffBottom(message.data.leftOff);
        setTruncatedTimestamp(message.data.truncatedTimestamp);
        if (leftOffTop === null) {
          setLeftOffTop(message.data.leftOff - 1);
        }
        break;
      case "startTime":
        setStartTime(message.data);
        break;
      default:
        console.error(
          `unsupported websocket message type, Got: ${message.messageType}`
        );
    }

    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [message]);

  useEffect(() => {
    onmessage(message)
  }, [message, onmessage])

  useEffect(() => {
    onerror(error)
  }, [error])


  useEffect(() => {
    isWebSocketOpen ? setWsConnection(WsConnectionStatus.Connected) : setWsConnection(WsConnectionStatus.Closed)
    trafficViewerApiProp.webSocket.sendQuery(queryToSend)
  }, [isWebSocketOpen, queryToSend, setWsConnection])

  const onerror = (event) => {
    console.error("WebSocket error:", event);
    if (query) {
      openWebSocket(`(${query}) and leftOff(${leftOffBottom})`, false);
    } else {
      openWebSocket(`leftOff(${leftOffBottom})`, false);
    }
  }

  useEffect(() => {
    (async () => {
      setTrafficViewerApiState(trafficViewerApiProp)
      openWebSocket("leftOff(-1)", true);
      try {
        const tapStatusResponse = await trafficViewerApiProp.tapStatus();
        setTappingStatus(tapStatusResponse);
        if (setAnalyzeStatus) {
          const analyzeStatusResponse = await trafficViewerApiProp.analyzeStatus();
          setAnalyzeStatus(analyzeStatusResponse);
        }
      } catch (error) {
        console.error(error);
      }
    })()
    // eslint-disable-next-line
  }, []);

  const toggleConnection = () => {
    if (wsConnection === WsConnectionStatus.Closed) {

      if (query) {
        openWebSocket(`(${query}) and leftOff(-1)`, true);
      } else {
        openWebSocket(`leftOff(-1)`, true);
      }
      scrollableRef.current.jumpToBottom();
      setIsSnappedToBottom(true);
    }
    else if (wsConnection === WsConnectionStatus.Connected) {
      trafficViewerApiProp.webSocket.close()
      setWsConnection(WsConnectionStatus.Closed);
    }
  }

  const onTLSDetected = (destAddress: string) => {
    addressesWithTLS.add(destAddress);
    setAddressesWithTLS(new Set(addressesWithTLS));

    if (!userDismissedTLSWarning) {
      setShowTLSWarning(true);
    }
  };

  const getConnectionIndicator = () => {
    switch (wsConnection) {
      case WsConnectionStatus.Connected:
        return <div className={`${TrafficViewerStyles.indicatorContainer} ${TrafficViewerStyles.greenIndicatorContainer}`}>
          <div className={`${TrafficViewerStyles.indicator} ${TrafficViewerStyles.greenIndicator}`} />
        </div>
      default:
        return <div className={`${TrafficViewerStyles.indicatorContainer} ${TrafficViewerStyles.redIndicatorContainer}`}>
          <div className={`${TrafficViewerStyles.indicator} ${TrafficViewerStyles.redIndicator}`} />
        </div>
    }
  }

  const getConnectionTitle = () => {
    switch (wsConnection) {
      case WsConnectionStatus.Connected:
        return "streaming live traffic"
      default:
        return "streaming paused";
    }
  }

  const onSnapBrokenEvent = () => {
    setIsSnappedToBottom(false);
    if (wsConnection === WsConnectionStatus.Connected) {
      trafficViewerApiProp.webSocket.close()
    }
  }

  return (
    <div className={TrafficViewerStyles.TrafficPage}>
      {tappingStatus && isShowStatusBar && <StatusBar />}
      <div className={TrafficViewerStyles.TrafficPageHeader}>
        <div className={TrafficViewerStyles.TrafficPageStreamStatus}>
          <img className={TrafficViewerStyles.playPauseIcon} style={{ visibility: wsConnection === WsConnectionStatus.Connected ? "visible" : "hidden" }} alt="pause"
            src={pauseIcon} onClick={toggleConnection} />
          <img className={TrafficViewerStyles.playPauseIcon} style={{ position: "absolute", visibility: wsConnection === WsConnectionStatus.Connected ? "hidden" : "visible" }} alt="play"
            src={playIcon} onClick={toggleConnection} />
          <div className={TrafficViewerStyles.connectionText}>
            {getConnectionTitle()}
            {getConnectionIndicator()}
          </div>
        </div>
        {actionButtons}
      </div>
      {<div className={TrafficViewerStyles.TrafficPageContainer}>
        <div className={TrafficViewerStyles.TrafficPageListContainer}>
          <Filters
            backgroundColor={queryBackgroundColor}
            openWebSocket={openWebSocket}

          />
          <div className={styles.container}>
            <EntriesList
              listEntryREF={listEntry}
              onSnapBrokenEvent={onSnapBrokenEvent}
              isSnappedToBottom={isSnappedToBottom}
              setIsSnappedToBottom={setIsSnappedToBottom}
              queriedCurrent={queriedCurrent}
              setQueriedCurrent={setQueriedCurrent}
              queriedTotal={queriedTotal}
              setQueriedTotal={setQueriedTotal}
              startTime={startTime}
              noMoreDataTop={noMoreDataTop}
              setNoMoreDataTop={setNoMoreDataTop}
              leftOffTop={leftOffTop}
              setLeftOffTop={setLeftOffTop}
              openWebSocket={openWebSocket}
              leftOffBottom={leftOffBottom}
              truncatedTimestamp={truncatedTimestamp}
              setTruncatedTimestamp={setTruncatedTimestamp}
              scrollableRef={scrollableRef}
            />
          </div>
        </div>
        <div className={classes.details} id="rightSideContainer">
          {focusedEntryId && <EntryDetailed />}
        </div>
      </div>}
      <TLSWarning showTLSWarning={showTLSWarning}
        setShowTLSWarning={setShowTLSWarning}
        addressesWithTLS={addressesWithTLS}
        setAddressesWithTLS={setAddressesWithTLS}
        userDismissedTLSWarning={userDismissedTLSWarning}
        setUserDismissedTLSWarning={setUserDismissedTLSWarning} />
    </div>
  );
};

const MemoiedTrafficViewer = React.memo(TrafficViewer)
const TrafficViewerContainer: React.FC<TrafficViewerProps> = ({ setAnalyzeStatus, message, isWebSocketOpen, trafficViewerApiProp, actionButtons, isShowStatusBar = true }) => {
  return <RecoilRoot>
    <MemoiedTrafficViewer message={message} isWebSocketOpen={isWebSocketOpen} actionButtons={actionButtons} isShowStatusBar={isShowStatusBar}
      trafficViewerApiProp={trafficViewerApiProp} setAnalyzeStatus={setAnalyzeStatus} />
  </RecoilRoot>
}

export default TrafficViewerContainer 