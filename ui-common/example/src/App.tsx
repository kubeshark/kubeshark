

 // eslint-disable-next-line @typescript-eslint/no-unused-vars
import TrafficViewer from '@up9/mizu-common';
import "@up9/mizu-common/dist/index.css"
import {  useRef, useState } from 'react';

import Api, {MizuWebsocketURL,getToken} from "./api";

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

  const onMessage = (e:any) => {setMessage(e)}
  const onError = (e:any) => setError(e)
  const onOpen = () => {setisOpen(true)}
  const onClose = () => setisOpen(false)

  const openScoket = () => {
    let websocketUrl = MizuWebsocketURL;
    const tk = getToken()
    if (tk) {
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
  
  
  const sendQuery = (query : any) =>{
      if(ws.current && (ws.current.readyState === WebSocketReadyState.OPEN)){
          ws.current.send(query)
      }
  }

  return <>

    <TrafficViewer message={{}} isOpen={false} closeWs={closeWs} sendQuery={sendQuery} openSocket={openScoket} trafficViewerApiProp={api} ></TrafficViewer>
  </>
}

export default App