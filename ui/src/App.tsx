import React, {useEffect, useRef, useState} from 'react';
import './App.sass';
import {Header} from "./components/Header/Header";
import {TrafficPage} from "./components/Pages/TrafficPage/TrafficPage";
import { ServiceMapModal } from './components/ServiceMapModal/ServiceMapModal';
import {useRecoilState} from "recoil";
import serviceMapModalOpenAtom from "./recoil/serviceMapModalOpen";
import { MizuWebsocketURL } from './helpers/api';
import Api from "././helpers/api"
import TrafficViewer from "@up9/mizu-common"
import "@up9/mizu-common/dist/index.css"

//  enum WebSocketReadyState{
//     CONNECTING,
//     OPEN,
//     CLOSING,
//     CLOSED
// }

// const api = Api.getInstance()

const App = () => {

    const [analyzeStatus, setAnalyzeStatus] = useState(null);
    const [serviceMapModalOpen, setServiceMapModalOpen] = useRecoilState(serviceMapModalOpenAtom);
    // const [message, setMessage] = useState(null);
    // const [error, setError] = useState(null);
    // const [isOpen, setisOpen] = useState(false);
    // const ws = useRef(null);

    // const onMessage = (e) => {setMessage(e)}
    // const onError = (e) => setError(e)
    // const onOpen = () => {setisOpen(true)}
    // const onClose = () => setisOpen(false)

    // const openScoket = (query = "") => {
    //     ws.current = new WebSocket(MizuWebsocketURL)
    //     ws.current.addEventListener("message",onMessage)
    //     ws.current.addEventListener("error",onError)
    //     ws.current.addEventListener("open",onOpen)
    //     ws.current.addEventListener("close",onClose)
    // }

    // const closeWs = () => {
    //     ws.current.readyState === WebSocketReadyState.OPEN && ws.current.close();
    //     ws.current.removeEventListener("message",onMessage)
    //     ws.current.removeEventListener("error",onError)
    //     ws.current.removeEventListener("open",onOpen)
    //     ws.current.removeEventListener("close",onClose)
    // }
        
    
    // const sendQuery = (query) =>{
    //     if(ws.current && (ws.current.readyState === WebSocketReadyState.OPEN)){
    //         ws.current.send(query)
    //     }
    // }

    // useEffect(() => {
    //     return () => {
    //         closeWs()
    //     }
    // },[])

    return (
        // <div className="mizuApp">
            
        //     <Header analyzeStatus={analyzeStatus} />
        //     <TrafficViewer setAnalyzeStatus={setAnalyzeStatus} message={message} error={error} isOpen={isOpen} closeWs={closeWs}
        //     sendQuery={sendQuery} openSocket={openScoket} trafficViewerApiProp={api} />
        //     {window["isServiceMapEnabled"] && <ServiceMapModal
        //         isOpen={serviceMapModalOpen}
        //         onOpen={() => setServiceMapModalOpen(true)}
        //         onClose={() => setServiceMapModalOpen(false)}
                
        //     />}
        // </div>
        <div className="mizuApp">
        <Header analyzeStatus={analyzeStatus} />
        <TrafficPage setAnalyzeStatus={setAnalyzeStatus}/>
        {window["isServiceMapEnabled"] && <ServiceMapModal
            isOpen={serviceMapModalOpen}
            onOpen={() => setServiceMapModalOpen(true)}
            onClose={() => setServiceMapModalOpen(false)}
        />}
    </div>
    );
}

export default App;
