import React, { useEffect, useMemo, useRef, useState } from "react";
import { Filters } from "../../Filters";
import { EntriesList } from "../../EntriesList";
import { makeStyles, Button } from "@material-ui/core";
import "./TrafficPage.sass";
import styles from '../../style/EntriesList.module.sass';
import {EntryDetailed} from "../../EntryDetailed";
import playIcon from '../../assets/run.svg';
import pauseIcon from '../../assets/pause.svg';
import variables from '../../../variables.module.scss';
import {StatusBar} from "../../UI/StatusBar";
import Api, {MizuWebsocketURL} from "../../../helpers/api";
import { toast } from 'react-toastify';
import debounce from 'lodash/debounce';
import {useRecoilState, useRecoilValue, useSetRecoilState} from "recoil";
import tappingStatusAtom from "../../../recoil/tappingStatus";
import entriesAtom from "../../../recoil/entries";
import focusedEntryIdAtom from "../../../recoil/focusedEntryId";
import websocketConnectionAtom, {WsConnectionStatus} from "../../../recoil/wsConnection";
import queryAtom from "../../../recoil/query";
import {useCommonStyles} from "../../../helpers/commonStyle"
import {TLSWarning} from "../../TLSWarning/TLSWarning";
import serviceMapModalOpenAtom from "../../../recoil/serviceMapModalOpen";
import serviceMap from "../../assets/serviceMap.svg";
import services from "../../assets/services.svg";
import oasModalOpenAtom from "../../../recoil/oasModalOpen/atom";

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

interface TrafficPageProps {
  setAnalyzeStatus?: (status: any) => void;
}

const api = Api.getInstance();

export const TrafficPage: React.FC<TrafficPageProps> = ({setAnalyzeStatus}) => {
    const commonClasses = useCommonStyles();
    const classes = useLayoutStyles();
    const [tappingStatus, setTappingStatus] = useRecoilState(tappingStatusAtom);
    const [entries, setEntries] = useRecoilState(entriesAtom);
    const [focusedEntryId, setFocusedEntryId] = useRecoilState(focusedEntryIdAtom);
    const [wsConnection, setWsConnection] = useRecoilState(websocketConnectionAtom);
    const setServiceMapModalOpen = useSetRecoilState(serviceMapModalOpenAtom);
    const [openOasModal, setOpenOasModal] = useRecoilState(oasModalOpenAtom);
    const query = useRecoilValue(queryAtom);

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
        }, 500),
      []
    ) as (query: string) => void;

    useEffect(() => {
        handleQueryChange(query);
    }, [query, handleQueryChange]);

    const ws = useRef(null);

    const listEntry = useRef(null);
    const openWebSocket = (query: string, resetEntries: boolean) => {
        if (resetEntries) {
            setFocusedEntryId(null);
            setEntries([]);
            setQueriedCurrent(0);
            setLeftOffTop(null);
            setNoMoreDataTop(false);
        }
        ws.current = new WebSocket(MizuWebsocketURL);
        ws.current.onopen = () => {
            setWsConnection(WsConnectionStatus.Connected);
            ws.current.send(query);
            ws.current.send("");
        }
        ws.current.onclose = () => {
            setWsConnection(WsConnectionStatus.Closed);
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
      ws.current.onmessage = (e) => {
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
      };
    }

    useEffect(() => {
          (async () => {
                openWebSocket("leftOff(-1)", true);
                try{
                    const tapStatusResponse = await api.tapStatus();
                    setTappingStatus(tapStatusResponse);
                    if(setAnalyzeStatus) {
                        const analyzeStatusResponse = await api.analyzeStatus();
                        setAnalyzeStatus(analyzeStatusResponse);
                    }
                } catch (error) {
                    console.error(error);
                }
            })()
            // eslint-disable-next-line
        }, []);

    const toggleConnection = () => {
      ws.current.close();
      if (wsConnection !== WsConnectionStatus.Connected) {
          if (query) {
              openWebSocket(`(${query}) and leftOff(-1)`, true);
          } else {
              openWebSocket(`leftOff(-1)`, true);
          }
          scrollableRef.current.jumpToBottom();
          setIsSnappedToBottom(true);
      }
    }

    const onTLSDetected = (destAddress: string) => {
        addressesWithTLS.add(destAddress);
        setAddressesWithTLS(new Set(addressesWithTLS));

        if (!userDismissedTLSWarning) {
            setShowTLSWarning(true);
        }
    };

    const getConnectionStatusClass = (isContainer) => {
        const container = isContainer ? "Container" : "";
        switch (wsConnection) {
            case WsConnectionStatus.Connected:
                return "greenIndicator" + container;
            default:
                return "redIndicator" + container;
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
            ws.current.close();
        }
    }

    const handleOpenOasModal = () => {
      ws.current.close();
      setOpenOasModal(true);
    }

    const openServiceMapModalDebounce = debounce(() => {
        ws.current.close();
        setServiceMapModalOpen(true);
    }, 500);

  return (
    <div className="TrafficPage">
      <div className="TrafficPageHeader">
        <div className="TrafficPageStreamStatus">
          <img className="playPauseIcon" style={{ visibility: wsConnection === WsConnectionStatus.Connected ? "visible" : "hidden" }} alt="pause"
            src={pauseIcon} onClick={toggleConnection} />
          <img className="playPauseIcon" style={{ position: "absolute", visibility: wsConnection === WsConnectionStatus.Connected ? "hidden" : "visible" }} alt="play"
            src={playIcon} onClick={toggleConnection} />
          <div className="connectionText">
            {getConnectionTitle()}
            <div className={"indicatorContainer " + getConnectionStatusClass(true)}>
              <div className={"indicator " + getConnectionStatusClass(false)} />
            </div>
          </div>
        </div>
        <div style={{ display: 'flex' }}>
          {window["isOasEnabled"] && <Button
            startIcon={<img className="custom" src={services} alt="services"></img>}
            size="large"
            type="submit"
            variant="contained"
            className={commonClasses.outlinedButton + " " + commonClasses.imagedButton}
            style={{ marginRight: 25 }}
            onClick={handleOpenOasModal}
          >
            Show OAS
          </Button>}
          {window["isServiceMapEnabled"] && <Button
            startIcon={<img src={serviceMap} className="custom" alt="service-map" style={{marginRight:"8%"}}></img>}
            size="large"
            variant="contained"
            className={commonClasses.outlinedButton + " " + commonClasses.imagedButton}
            onClick={openServiceMapModalDebounce}
          >
            Service Map
          </Button>}
        </div>
      </div>
      {<div className="TrafficPage-Container">
        <div className="TrafficPage-ListContainer">
          <Filters
            backgroundColor={queryBackgroundColor}
            ws={ws.current}
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
              ws={ws.current}
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
      {tappingStatus && !openOasModal && <StatusBar />}
       <TLSWarning showTLSWarning={showTLSWarning}
                   setShowTLSWarning={setShowTLSWarning}
                   addressesWithTLS={addressesWithTLS}
                   setAddressesWithTLS={setAddressesWithTLS}
                   userDismissedTLSWarning={userDismissedTLSWarning}
                   setUserDismissedTLSWarning={setUserDismissedTLSWarning} />
    </div>
  );
};
