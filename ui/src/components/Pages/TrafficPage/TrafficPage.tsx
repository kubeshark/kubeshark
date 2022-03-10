import React, { useEffect, useRef, useState } from "react";
import { Button } from "@material-ui/core";
import Api, {getWebsocketUrl} from "../../../helpers/api";
import debounce from 'lodash/debounce';
import {useSetRecoilState, useRecoilState} from "recoil";
import {useCommonStyles} from "../../../helpers/commonStyle"
import serviceMapModalOpenAtom from "../../../recoil/serviceMapModalOpen";
import TrafficViewer  from "@up9/mizu-common"
import "@up9/mizu-common/dist/index.css"
import oasModalOpenAtom from "../../../recoil/oasModalOpen/atom";
import serviceMap from "../../assets/serviceMap.svg";	
import services from "../../assets/services.svg";	
import tappingStatusAtom from "../../../recoil/tappingStatus/atom";
import {StatusBar} from "@up9/mizu-common"

enum WebSocketReadyState{
  CONNECTING,
  OPEN,
  CLOSING,
  CLOSED
}

interface TrafficPageProps {
  setAnalyzeStatus?: (status: any) => void;
}

const api = Api.getInstance();

export const TrafficPage: React.FC<TrafficPageProps> = ({setAnalyzeStatus}) => {
  const setServiceMapModalOpen = useSetRecoilState(serviceMapModalOpenAtom);
  const [tappingStatus, setTappingStatus] = useRecoilState(tappingStatusAtom);
  const [openOasModal,setOpenOasModal] = useRecoilState(oasModalOpenAtom);

  const commonClasses = useCommonStyles();
  const [message, setMessage] = useState(null);
  const [error, setError] = useState(null);
  const [isOpen, setisOpen] = useState(false);
  const ws = useRef(null);

  const openServiceMapModalDebounce = debounce(() => {
    setServiceMapModalOpen(true)
  }, 500);

  const onMessage = (e) => {setMessage(e)}
  const onError = (e) => setError(e)
  const onOpen = () => {setisOpen(true)}
  const onClose = () => setisOpen(false)

  const openScoket = (query = "") => {
    ws.current = new WebSocket(getWebsocketUrl())
    ws.current.addEventListener("message",onMessage)
    ws.current.addEventListener("error",onError)
    ws.current.addEventListener("open",onOpen)
    ws.current.addEventListener("close",onClose)
  }

  const closeSocket = () => {
    ws.current.readyState === WebSocketReadyState.OPEN && ws.current.close();
    ws.current.removeEventListener("message",onMessage)
    ws.current.removeEventListener("error",onError)
    ws.current.removeEventListener("open",onOpen)
    ws.current.removeEventListener("close",onClose)
}


const sendQuery = (query: string) =>{
    if(ws.current && (ws.current.readyState === WebSocketReadyState.OPEN)){
      ws.current.send(JSON.stringify({"query": query, "enableFullEntries": false}));
    }
}

const trafficViewerApi = {...api, webSocket:{open : openScoket, close: closeSocket, sendQuery: sendQuery}}

  const handleOpenOasModal = () => {	
    ws.current.close();	
    setOpenOasModal(true);	
  }

  useEffect(() => {
      return () => {
        if(ws.current)
        closeSocket()
      }
  },[])

  return (
    
    <>
      <div className="TrafficPageHeader">
      <div style={{ display: 'flex', height: "100%" }}>	
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
      <TrafficViewer setAnalyzeStatus={setAnalyzeStatus} setTappingStatus={setTappingStatus} message={message} error={error} isOpen={isOpen}
                     trafficViewerApiProp={trafficViewerApi} />
      {tappingStatus && !openOasModal && <StatusBar/>}
    </>
  );
};