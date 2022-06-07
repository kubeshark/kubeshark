import './App.sass';
import { Header } from "./components/Header/Header";
import { TrafficPage } from "./components/Pages/TrafficPage/TrafficPage";
import { ServiceMapModal } from '@up9/mizu-common';
import { useRecoilState } from "recoil";
import serviceMapModalOpenAtom from "./recoil/serviceMapModalOpen";
import oasModalOpenAtom from './recoil/oasModalOpen/atom';
import { OasModal } from '@up9/mizu-common';
import Api from './helpers/api';
import { ThemeProvider, Theme, StyledEngineProvider, createTheme } from '@mui/material';

declare module '@mui/styles/defaultTheme' {
    // eslint-disable-next-line @typescript-eslint/no-empty-interface
    interface DefaultTheme extends Theme {}
  }
  
  
  const theme = createTheme(({
      //here you set palette, typography ect...
    }))
  

const api = Api.getInstance()

const App = () => {

    const [serviceMapModalOpen, setServiceMapModalOpen] = useRecoilState(serviceMapModalOpenAtom);
    const [oasModalOpen, setOasModalOpen] = useRecoilState(oasModalOpenAtom)

    return (
        <StyledEngineProvider injectFirst>
        <ThemeProvider theme={theme}>
        <div className="mizuApp">
            <Header />
            <TrafficPage />
            {true && <ServiceMapModal
                isOpen={serviceMapModalOpen}
                onOpen={() => setServiceMapModalOpen(true)}
                onClose={() => setServiceMapModalOpen(false)}
                getServiceMapDataApi={api.serviceMapData} />}
            {true && <OasModal
                getOasServices={api.getOasServices}
                getOasByService={api.getOasByService}
                openModal={oasModalOpen}
                handleCloseModal={() => setOasModalOpen(false)}
            />}
        </div>
        </ThemeProvider>
        </StyledEngineProvider>
    );
}

export default App;
