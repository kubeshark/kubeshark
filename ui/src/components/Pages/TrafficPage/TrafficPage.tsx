import React, { useEffect, useRef, useState } from "react";
import { Button } from "@material-ui/core";
import "./TrafficPage.sass";
import Api, {MizuWebsocketURL,getToken} from "../../../helpers/api";
import debounce from 'lodash/debounce';
import {useSetRecoilState} from "recoil";
import OasModal from "../../OasModal/OasModal";
import {useCommonStyles} from "../../../helpers/commonStyle"
import serviceMapModalOpenAtom from "../../../recoil/serviceMapModalOpen";
import TrafficViewer  from "@up9/mizu-common"
import "@up9/mizu-common/dist/index.css"

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

  const commonClasses = useCommonStyles();
  const [message, setMessage] = useState(null);
  const [error, setError] = useState(null);
  const [isOpen, setisOpen] = useState(false);
  const ws = useRef(null);

  const [openOasModal, setOpenOasModal] = useState(false);
  const handleOpenModal = () => setOpenOasModal(true);
  const handleCloseModal = () => setOpenOasModal(false);
  const openServiceMapModalDebounce = debounce(() => {
    setServiceMapModalOpen(true)
  }, 500);

  const onMessage = (e) => {setMessage(e)}
  const onError = (e) => setError(e)
  const onOpen = () => {setisOpen(true)}
  const onClose = () => setisOpen(false)

  const openScoket = (query = "") => {
    let websocketUrl = MizuWebsocketURL;
    if (getToken()) {
      websocketUrl += `/${getToken()}`;
    }
    ws.current = new WebSocket(websocketUrl)
    ws.current.addEventListener("message",onMessage)
    ws.current.addEventListener("error",onError)
    ws.current.addEventListener("open",onOpen)
    ws.current.addEventListener("close",onClose)
  }

  const closeWs = () => {
      ws.current.readyState === WebSocketReadyState.OPEN && ws.current.close();
      ws.current.removeEventListener("message",onMessage)
      ws.current.removeEventListener("error",onError)
      ws.current.removeEventListener("open",onOpen)
      ws.current.removeEventListener("close",onClose)
  }
  
  
  const sendQuery = (query) =>{
      if(ws.current && (ws.current.readyState === WebSocketReadyState.OPEN)){
          ws.current.send(query)
      }
  }

  useEffect(() => {
      return () => {
        if(ws.current)
          closeWs()
      }
  },[])

  return (
    
    <>
      <div className="TrafficPageHeader">
        <div style={{ display: 'flex' }}>
          {window["isOasEnabled"] && <Button type="submit" variant="contained" className={commonClasses.button} style={{ marginRight: 25 }} onClick={handleOpenModal}>
            Show OAS
          </Button>}
          {window["isServiceMapEnabled"] && <Button variant="contained" className={commonClasses.button} onClick={openServiceMapModalDebounce}>
            Service Map
          </Button>}
          {window["isOasEnabled"] && <OasModal openModal={openOasModal} handleCloseModal={handleCloseModal}/>}
        </div>
      </div>
      <TrafficViewer setAnalyzeStatus={setAnalyzeStatus} message={message} error={error} isOpen={isOpen} closeWs={closeWs}
                     sendQuery={sendQuery} openSocket={openScoket} trafficViewerApiProp={api} />
    </>
  );
};
