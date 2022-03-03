import { Box, Fade, FormControl, MenuItem, Modal, Backdrop, ListSubheader } from "@material-ui/core";
import { useCallback, useEffect, useState } from "react";
import { RedocStandalone } from "redoc";
import Api from "../../helpers/api";
import { Select } from "../UI/Select";
import closeIcon from "../assets/closeIcon.svg";
import { toast } from 'react-toastify';
import style from './OasModal.module.sass';
import opnApiLogo from '../assets/openApiLogo.png'

const modalStyle = {
  position: 'absolute',
  top: '6%',
  left: '50%',
  transform: 'translate(-50%, 0%)',
  width: '89vw',
  height: '82vh',
  bgcolor: 'background.paper',
  borderRadius: '5px',
  boxShadow: 24,
  p: 4,
  color: '#000',
};

const api = Api.getInstance();
const ipAddressWithPortRegex = new RegExp('([0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3}):([0-9]{1,5})');

const redocThemeOptions = {     
  theme:{
    codeBlock:{
      backgroundColor:"#11171a",
    },
    colors:{
      responses:{
        error:{
          tabTextColor:"#1b1b29"
        },
        info:{
          tabTextColor:"#1b1b29",
        },
        success:{
          tabTextColor:"#0c0b1a"
        },
      },
      text:{
        primary:"#1b1b29",
        secondary:"#4d4d4d"
      }
    },
    rightPanel:{
      backgroundColor:"#253237",
    },
    sidebar:{
      backgroundColor:"#ffffff"
    },
    typography:{
      code:{
        color:"#0c0b1a"
      }
    }
  }
}

const OasModal = ({ openModal, handleCloseModal }) => { 
  const [oasServices, setOasServices] = useState([] as string[])
  const [selectedServiceName, setSelectedServiceName] = useState("");
  const [selectedServiceSpec, setSelectedServiceSpec] = useState(null);
  const [resolvedServices, setResolvedServices] = useState([]);
  const [unResolvedServices, setUnResolvedServices] = useState([]);

  const onSelectedOASService = useCallback( async (selectedService) => {
    if (!!selectedService){
      setSelectedServiceName(selectedService);
      if(oasServices.length === 0){
        return
      }
      try {
        const data = await api.getOasByService(selectedService);
        setSelectedServiceSpec(data);
      } catch (e) {
        toast.error("Error occurred while fetching service OAS spec");
        console.error(e);
      }
    }
  },[oasServices.length])

  const resolvedArrayBuilder = useCallback(async(services) => {
    const resServices = [];
    const unResServices = [];
    services.forEach(s => {
      if(ipAddressWithPortRegex.test(s)){
        unResServices.push(s);
      }
      else {
        resServices.push(s);
      }
    });

    resServices.sort();
    unResServices.sort();
    onSelectedOASService(resServices[0]);
    setResolvedServices(resServices);
    setUnResolvedServices(unResServices);
  },[onSelectedOASService])

  useEffect(() => {
    (async () => {
      try {
        const services = await api.getOasServices();
        resolvedArrayBuilder(services);
        setOasServices(services);
      } catch (e) {
        console.error(e);
      }
    })();
  }, [openModal,resolvedArrayBuilder]);



  return (
    <Modal
      aria-labelledby="transition-modal-title"
      aria-describedby="transition-modal-description"
      open={openModal}
      onClose={handleCloseModal}
      closeAfterTransition
      BackdropComponent={Backdrop}
      BackdropProps={{
          timeout: 500,
      }}
    >
      <Fade in={openModal}>
        <Box sx={modalStyle}>
          <div className={style.boxContainer}>
            <div className={style.selectHeader}>
              <div><img src={opnApiLogo} alt="openApi" className={style.openApilogo}/></div>
                <div className={style.title}>OpenAPI selected service: </div>
                <div className={style.selectContainer} >
                  <FormControl>
                    <Select
                      labelId="service-select-label"
                      id="service-select"
                      placeholder="Show OAS"
                      value={selectedServiceName}
                      onChange={onSelectedOASService}
                    >
                      <ListSubheader disableSticky={true}>Resolved</ListSubheader>
                      {resolvedServices.map((service) => (
                        <MenuItem key={service} value={service}>
                          {service}
                        </MenuItem>
                      ))}
                      <ListSubheader disableSticky={true}>UnResolved</ListSubheader>
                      {unResolvedServices.map((service) => (
                        <MenuItem key={service} value={service}>
                          {service}
                        </MenuItem>
                      ))} 
                    </Select>
                  </FormControl>
                </div>
            </div>
            <div style={{ cursor: "pointer" }}>
              <img src={closeIcon} alt="close" onClick={handleCloseModal} />
            </div>
          </div>
          <div className={style.redoc}>
          {selectedServiceSpec && <RedocStandalone 
                                    spec={selectedServiceSpec}   
                                    options={redocThemeOptions}/>}
          </div>
        </Box>
      </Fade>
    </Modal>
  );
};

export default OasModal;
