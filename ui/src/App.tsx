import React, {useState} from 'react';
import './App.sass';
import {Header} from "./components/Header/Header";
import {TrafficPage} from "./components/Pages/TrafficPage/TrafficPage";
import { ServiceMapModal } from './components/ServiceMapModal/ServiceMapModal';
import {useRecoilState} from "recoil";
import serviceMapModalOpenAtom from "./recoil/serviceMapModalOpen";

const App = () => {

    const [analyzeStatus, setAnalyzeStatus] = useState(null);
    const [serviceMapModalOpen, setServiceMapModalOpen] = useRecoilState(serviceMapModalOpenAtom);

    return (
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
