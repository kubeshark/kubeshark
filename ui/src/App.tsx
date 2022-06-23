import './App.sass';
import { Header } from "./components/Header/Header";
import { TrafficPage } from "./components/Pages/TrafficPage/TrafficPage";
import { ServiceMapModal } from '@up9/mizu-common';
import { useRecoilState } from "recoil";
import serviceMapModalOpenAtom from "./recoil/serviceMapModalOpen";
import oasModalOpenAtom from './recoil/oasModalOpen/atom';
import trafficStatsModalOpenAtom from "./recoil/trafficStatsModalOpen";
import { OasModal } from '@up9/mizu-common';
import Api from './helpers/api';
import {ThemeProvider, StyledEngineProvider, createTheme} from '@mui/material';
import { TrafficStatsModal } from '@up9/mizu-common';

const api = Api.getInstance()

const App = () => {

    const [serviceMapModalOpen, setServiceMapModalOpen] = useRecoilState(serviceMapModalOpenAtom);
    const [oasModalOpen, setOasModalOpen] = useRecoilState(oasModalOpenAtom)
    const [trafficStatsModalOpen, setTrafficStatsModalOpen] = useRecoilState(trafficStatsModalOpenAtom);

    return (
        <StyledEngineProvider injectFirst>
            <ThemeProvider theme={createTheme(({}))}>
                <div className="mizuApp">
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
                    <TrafficStatsModal isOpen={trafficStatsModalOpen} onClose={() => setTrafficStatsModalOpen(false)} getPieStatsDataApi={api.getPieStats} getTimelineStatsDataApi={api.getTimelineStats}/>
                </div>
            </ThemeProvider>
        </StyledEngineProvider>
    );
}

export default App;
