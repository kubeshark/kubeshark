import React, {useEffect, useMemo, useRef, useState} from "react";
import {Filters} from "./Filters";
import {EntriesList} from "./EntriesList";
import {makeStyles} from "@material-ui/core";
import TrafficViewerStyles from "./TrafficViewer.module.sass";
import styles from '../style/EntriesList.module.sass';
import {EntryDetailed} from "./EntryDetailed";
import playIcon from 'assets/run.svg';
import pauseIcon from 'assets/pause.svg';
import variables from '../../variables.module.scss';
import {ToastContainer} from 'react-toastify';
import debounce from 'lodash/debounce';
import {RecoilRoot, RecoilState, useRecoilState, useRecoilValue, useSetRecoilState} from "recoil";
import entriesAtom from "../../recoil/entries";
import focusedEntryIdAtom from "../../recoil/focusedEntryId";
import queryAtom from "../../recoil/query";
import {TLSWarning} from "../TLSWarning/TLSWarning";
import trafficViewerApiAtom from "../../recoil/TrafficViewerApi"
import TrafficViewerApi from "./TrafficViewerApi";
import {StatusBar} from "../UI/StatusBar";
import tappingStatusAtom from "../../recoil/tappingStatus/atom";
import {TOAST_CONTAINER_ID} from "../../configs/Consts";
import leftOffTopAtom from "../../recoil/leftOffTop";

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
  trafficViewerApiProp: TrafficViewerApi,
  actionButtons?: JSX.Element,
  isShowStatusBar?: boolean,
  webSocketUrl: string,
  isCloseWebSocket: boolean,
  isDemoBannerView: boolean
}

export const TrafficViewer: React.FC<TrafficViewerProps> = ({
                                                              setAnalyzeStatus, trafficViewerApiProp,
                                                              actionButtons, isShowStatusBar, webSocketUrl,
                                                              isCloseWebSocket, isDemoBannerView
                                                            }) => {

  const classes = useLayoutStyles();

  const setEntries = useSetRecoilState(entriesAtom);
  const setFocusedEntryId = useSetRecoilState(focusedEntryIdAtom);
  const query = useRecoilValue(queryAtom);
  const setTrafficViewerApiState = useSetRecoilState(trafficViewerApiAtom as RecoilState<TrafficViewerApi>)
  const [tappingStatus, setTappingStatus] = useRecoilState(tappingStatusAtom);
  const [noMoreDataTop, setNoMoreDataTop] = useState(false);
  const [isSnappedToBottom, setIsSnappedToBottom] = useState(true);
  const [wsReadyState, setWsReadyState] = useState(0);

  const [queryBackgroundColor, setQueryBackgroundColor] = useState("#f5f5f5");

  const setLeftOffTop = useSetRecoilState(leftOffTopAtom);
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

  useEffect(() => {
    isCloseWebSocket && closeWebSocket()
  }, [isCloseWebSocket])

  useEffect(() => {
    reopenConnection()
  }, [webSocketUrl])

  const ws = useRef(null);

  const openEmptyWebSocket = () => {
    if (query) {
      openWebSocket(`(${query}) and leftOff(-1)`, true);
    } else {
      openWebSocket(`leftOff(-1)`, true);
    }
  }

  const closeWebSocket = () => {
    if (ws?.current?.readyState === WebSocket.OPEN) {
      ws.current.close();
      return true;
    }
  }

  const listEntry = useRef(null);
  const openWebSocket = (query: string, resetEntries: boolean) => {
    if (resetEntries) {
      setFocusedEntryId(null);
      setEntries([]);
      setLeftOffTop(null);
      setNoMoreDataTop(false);
    }
    try {
      ws.current = new WebSocket(webSocketUrl);
      sendQueryWhenWsOpen(query);

      ws.current.onopen = () => {
        setWsReadyState(ws?.current?.readyState);
      }

      ws.current.onclose = () => {
        setWsReadyState(ws?.current?.readyState);
      }
      ws.current.onerror = (event) => {
        console.error("WebSocket error:", event);
        if (ws?.current?.readyState === WebSocket.OPEN) {
          ws.current.close();
        }
      }
    } catch (e) {
    }
  }

  const sendQueryWhenWsOpen = (query) => {
    setTimeout(() => {
      if (ws?.current?.readyState === WebSocket.OPEN) {
        ws.current.send(JSON.stringify({"query": query, "enableFullEntries": false}));
      } else {
        sendQueryWhenWsOpen(query);
      }
    }, 500)
  }


  useEffect(() => {
    setTrafficViewerApiState({...trafficViewerApiProp, webSocket: {close: closeWebSocket}});
    (async () => {
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
  }, []);

  const toggleConnection = () => {
    if (!closeWebSocket()) {
      openEmptyWebSocket();
      scrollableRef.current.jumpToBottom();
      setIsSnappedToBottom(true);
    }
  }

  const reopenConnection = async () => {
    closeWebSocket()
    openEmptyWebSocket();
    scrollableRef.current.jumpToBottom();
    setIsSnappedToBottom(true);
  }

  useEffect(() => {
    return () => {
      if (ws?.current?.readyState === WebSocket.OPEN) {
        ws.current.close();
      }
    };
  }, []);

  const getConnectionIndicator = () => {
    switch (wsReadyState) {
      case WebSocket.OPEN:
        return <div
          className={`${TrafficViewerStyles.indicatorContainer} ${TrafficViewerStyles.greenIndicatorContainer}`}>
          <div className={`${TrafficViewerStyles.indicator} ${TrafficViewerStyles.greenIndicator}`}/>
        </div>
      default:
        return <div
          className={`${TrafficViewerStyles.indicatorContainer} ${TrafficViewerStyles.redIndicatorContainer}`}>
          <div className={`${TrafficViewerStyles.indicator} ${TrafficViewerStyles.redIndicator}`}/>
        </div>
    }
  }

  const getConnectionTitle = () => {
    switch (wsReadyState) {
      case WebSocket.OPEN:
        return "streaming live traffic"
      default:
        return "streaming paused";
    }
  }

  const onSnapBrokenEvent = () => {
    setIsSnappedToBottom(false);
    if (ws?.current?.readyState === WebSocket.OPEN) {
      ws.current.close();
    }
  }

  return (
    <div className={TrafficViewerStyles.TrafficPage}>
      {tappingStatus && isShowStatusBar && <StatusBar isDemoBannerView={isDemoBannerView}/>}
      <div className={TrafficViewerStyles.TrafficPageHeader}>
        <div className={TrafficViewerStyles.TrafficPageStreamStatus}>
          <img className={TrafficViewerStyles.playPauseIcon}
               style={{visibility: wsReadyState === WebSocket.OPEN ? "visible" : "hidden"}} alt="pause"
               src={pauseIcon} onClick={toggleConnection}/>
          <img className={TrafficViewerStyles.playPauseIcon}
               style={{position: "absolute", visibility: wsReadyState === WebSocket.OPEN ? "hidden" : "visible"}}
               alt="play"
               src={playIcon} onClick={toggleConnection}/>
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
            reopenConnection={reopenConnection}
          />
          <div className={styles.container}>
            <EntriesList
              listEntryREF={listEntry}
              onSnapBrokenEvent={onSnapBrokenEvent}
              isSnappedToBottom={isSnappedToBottom}
              setIsSnappedToBottom={setIsSnappedToBottom}
              noMoreDataTop={noMoreDataTop}
              setNoMoreDataTop={setNoMoreDataTop}
              openWebSocket={openWebSocket}
              scrollableRef={scrollableRef}
              ws={ws}
            />
          </div>
        </div>
        <div className={classes.details} id="rightSideContainer">
          <EntryDetailed/>
        </div>
      </div>}
      <TLSWarning showTLSWarning={showTLSWarning}
                  setShowTLSWarning={setShowTLSWarning}
                  addressesWithTLS={addressesWithTLS}
                  setAddressesWithTLS={setAddressesWithTLS}
                  userDismissedTLSWarning={userDismissedTLSWarning}
                  setUserDismissedTLSWarning={setUserDismissedTLSWarning}/>
    </div>
  );
};

const MemoiedTrafficViewer = React.memo(TrafficViewer)
const TrafficViewerContainer: React.FC<TrafficViewerProps> = ({
                                                                setAnalyzeStatus, trafficViewerApiProp,
                                                                actionButtons, isShowStatusBar = true,
                                                                webSocketUrl, isCloseWebSocket, isDemoBannerView
                                                              }) => {
  return <RecoilRoot>
    <MemoiedTrafficViewer actionButtons={actionButtons} isShowStatusBar={isShowStatusBar} webSocketUrl={webSocketUrl}
                          isCloseWebSocket={isCloseWebSocket} trafficViewerApiProp={trafficViewerApiProp}
                          setAnalyzeStatus={setAnalyzeStatus} isDemoBannerView={isDemoBannerView}/>
    <ToastContainer enableMultiContainer containerId={TOAST_CONTAINER_ID}
                    position="bottom-right"
                    autoClose={5000}
                    hideProgressBar={false}
                    newestOnTop={false}
                    closeOnClick
                    rtl={false}
                    pauseOnFocusLoss
                    draggable
                    pauseOnHover/>
  </RecoilRoot>
}

export default TrafficViewerContainer
