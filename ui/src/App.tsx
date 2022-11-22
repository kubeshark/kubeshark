import './App.sass';
import { Header } from "./components/Header/Header";
import { TrafficPage } from "./components/Pages/TrafficPage/TrafficPage";
import { ServiceMapModal } from './components/modals/ServiceMapModal/ServiceMapModal';
import { useRecoilState } from "recoil";
import serviceMapModalOpenAtom from "./recoil/serviceMapModalOpen";
import oasModalOpenAtom from './recoil/oasModalOpen/atom';
import trafficStatsModalOpenAtom from "./recoil/trafficStatsModalOpen";
import { OasModal } from './components/modals/OasModal/OasModal';
import Api from './helpers/api';
import { ThemeProvider, StyledEngineProvider, createTheme } from '@mui/material';
import { TrafficStatsModal } from './components/modals/TrafficStatsModal/TrafficStatsModal';

const api = Api.getInstance()

const App = () => {

    const [serviceMapModalOpen, setServiceMapModalOpen] = useRecoilState(serviceMapModalOpenAtom);
    const [oasModalOpen, setOasModalOpen] = useRecoilState(oasModalOpenAtom)
    const [trafficStatsModalOpen, setTrafficStatsModalOpen] = useRecoilState(trafficStatsModalOpenAtom);

    return (
        <StyledEngineProvider injectFirst>
            <ThemeProvider theme={createTheme(({}))}>
                <div className="kubesharkApp">
                    <Header />
                    <TrafficPage />
                    {window["isServiceMapEnabled"] && <ServiceMapModal
                        isOpen={serviceMapModalOpen}
                        onOpen={() => setServiceMapModalOpen(true)}
                        onClose={() => setServiceMapModalOpen(false)}
                        getServiceMapDataApi={api.serviceMapData} />}
                    {window["isOasEnabled"] && <OasModal
                        getOasServices={api.getOasServices}
                        getOasByService={api.getOasByService}
                        openModal={oasModalOpen}
                        handleCloseModal={() => setOasModalOpen(false)}
                    />}
                    <TrafficStatsModal isOpen={trafficStatsModalOpen} onClose={() => setTrafficStatsModalOpen(false)} getTrafficStatsDataApi={api.getTrafficStats} />
                </div>
            </ThemeProvider>
        </StyledEngineProvider>
    );
}

export default App;
