

 // eslint-disable-next-line @typescript-eslint/no-unused-vars
import TrafficViewer from '@up9/mizu-common';
import "@up9/mizu-common/dist/index.css"
import {  useEffect, useRef, useState } from 'react';

import Api, {getWebsocketUrl} from "./api";

const api = Api.getInstance()
enum WebSocketReadyState{
  CONNECTING,
  OPEN,
  CLOSING,
  CLOSED
}


const App = () => {
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  const [message, setMessage] = useState(null);
  const [error, setError] = useState(null);
  const [isOpen, setisOpen] = useState(false);
  const ws = useRef(null);
  
  const onMessage = (e: any) => {setMessage(e)}
  const onError = (e: any) => setError(e)
  const onOpen = () => {setisOpen(true)}
  const onClose = () => setisOpen(false)

  const openScoket = () => {
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
  
  const sendQuery = (query: string) => {
      if(ws.current && (ws.current.readyState === WebSocketReadyState.OPEN)){
        ws.current.send(JSON.stringify({"query": query, "enableFullEntries": false}));
      }
  }

  const trafficViewerApi = {...api, webSocket:{open : openScoket, close: closeSocket, sendQuery: sendQuery}}

  useEffect(() => {
      return () => {
        if(ws.current)
          closeSocket()
        }
  },[])

  return <>
    <TrafficViewer message={message} error={error} isOpen={isOpen} 
                   trafficViewerApiProp={trafficViewerApi} setTappingStatus={()=>{}} ></TrafficViewer>
  </>
}

export default App