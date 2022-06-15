import './App.sass';
import { Header } from "./components/Header/Header";
import { TrafficPage } from "./components/Pages/TrafficPage/TrafficPage";
import { ServiceMapModal } from '@up9/mizu-common';
import { useRecoilState } from "recoil";
import serviceMapModalOpenAtom from "./recoil/serviceMapModalOpen";
import oasModalOpenAtom from './recoil/oasModalOpen/atom';
import { OasModal } from '@up9/mizu-common';
import Api from './helpers/api';
import {ThemeProvider, StyledEngineProvider, createTheme, Button} from '@mui/material';
import {TrafficStatsModal} from '@up9/mizu-common';
import {useState} from "react";

const api = Api.getInstance()

const App = () => {

    const [serviceMapModalOpen, setServiceMapModalOpen] = useRecoilState(serviceMapModalOpenAtom);
    const [oasModalOpen, setOasModalOpen] = useRecoilState(oasModalOpenAtom)
    const [trafficStatsModalOpen, setTrafficStatsModalOpen] = useState(false);
    const [stats, setStats] = useState(null);

    const onClick = async () => {
        try {
            const res = await api.getStats();
            setStats(res);
            setTrafficStatsModalOpen(true)
        } catch (e) {
            console.error(e)
        }
    }

    return (
        <StyledEngineProvider injectFirst>
            <ThemeProvider theme={createTheme(({}))}>
                <div className="mizuApp">
                    <Header />
                    <Button onClick={onClick}>STATS</Button>
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
                    <TrafficStatsModal isOpen={trafficStatsModalOpen} onClose={() => {setTrafficStatsModalOpen(false)}} data={stats}/>
                </div>
            </ThemeProvider>
        </StyledEngineProvider>
    );
}

export default App;
