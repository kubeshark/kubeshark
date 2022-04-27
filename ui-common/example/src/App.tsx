import {useWS, DEFAULT_QUERY} from '@up9/mizu-common';
import "@up9/mizu-common/dist/index.css"
import {useEffect} from 'react';
import {getWebsocketUrl} from "./api";

const App = () => {
  const {closeSocket, sendQueryWhenWsOpen} = useWS(getWebsocketUrl())
  sendQueryWhenWsOpen(DEFAULT_QUERY);

  useEffect(() => {
    return () =>{
      closeSocket()
    }
  },[closeSocket])

  return <>

  </>
}

export default App
