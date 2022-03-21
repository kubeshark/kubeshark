import TrafficViewer,{useWS, DEFAULT_QUERY} from '@up9/mizu-common';
import "@up9/mizu-common/dist/index.css"
import {useEffect} from 'react';
import Api, {getWebsocketUrl} from "./api";

const api = Api.getInstance()

const App = () => {
  const {message,error,isOpen, openSocket, closeSocket, sendQueryWhenWsOpen} = useWS(getWebsocketUrl())
  const trafficViewerApi = {...api, webSocket:{open : openSocket, close: closeSocket, sendQueryWhenWsOpen: sendQueryWhenWsOpen}}
  sendQueryWhenWsOpen(DEFAULT_QUERY);

  useEffect(() => {
    return () =>{
      closeSocket()
    }
  },[])

  return <>
    <TrafficViewer message={message} error={error} isWebSocketOpen={isOpen}
                   trafficViewerApiProp={trafficViewerApi} ></TrafficViewer>
  </>
}

export default App
