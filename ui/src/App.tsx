import { useState} from 'react';
import './App.sass';
import {Header} from "./components/Header/Header";
import {TrafficPage} from "./components/Pages/TrafficPage/TrafficPage";
import { ServiceMapModal } from './components/ServiceMapModal/ServiceMapModal';
import {useRecoilState} from "recoil";
import serviceMapModalOpenAtom from "./recoil/serviceMapModalOpen";
import oasModalOpenAtom from './recoil/oasModalOpen/atom';
import {OasModal} from '@up9/mizu-common';
import Api from './helpers/api';

const api = Api.getInstance()

const App = () => {

    const [analyzeStatus, setAnalyzeStatus] = useState(null);
    const [serviceMapModalOpen, setServiceMapModalOpen] = useRecoilState(serviceMapModalOpenAtom);
    const [oasModalOpen, setOasModalOpen] = useRecoilState(oasModalOpenAtom)

    return (
        <div className="mizuApp">
        <Header analyzeStatus={analyzeStatus} />
        <TrafficPage setAnalyzeStatus={setAnalyzeStatus}/>
        {window["isServiceMapEnabled"] && <ServiceMapModal
                isOpen={serviceMapModalOpen}
                onOpen={() => setServiceMapModalOpen(true)}
                onClose={() => setServiceMapModalOpen(false)}
            />}
        {window["isOasEnabled"] && <OasModal
            getOasServices={api.getOasServices}
            getOasByService={api.getOasByService}
            openModal={oasModalOpen}
            handleCloseModal={() => setOasModalOpen(false)}
        />}
    </div>
    );
}

export default App;
