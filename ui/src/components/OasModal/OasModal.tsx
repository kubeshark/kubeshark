import { Box, Fade, FormControl, MenuItem, Modal, Backdrop, ListSubheader } from "@material-ui/core";
import { useEffect, useState } from "react";
import { RedocStandalone } from "redoc";
import Api from "../../helpers/api";
import { Select } from "../UI/Select";
import closeIcon from "../assets/closeIcon.svg";
import { toast } from 'react-toastify';
import style from './OasModal.module.sass';

const modalStyle = {
  position: 'absolute',
  top: '10%',
  left: '50%',
  transform: 'translate(-50%, 0%)',
  width: '80vw',
  height: '80vh',
  bgcolor: 'background.paper',
  borderRadius: '5px',
  boxShadow: 24,
  p: 4,
  color: '#000',
};

const api = Api.getInstance();
const noOasServiceSelectedMessage = "Please Select OasService";
const ipAddressWithPortRegex = new RegExp('([0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3}):([0-9]{1,5})');

const OasModal = ({ openModal, handleCloseModal }) => { 
  const [oasServices, setOasServices] = useState([] as string[])
  const [selectedServiceName, setSelectedServiceName] = useState("");
  const [selectedServiceSpec, setSelectedServiceSpec] = useState(null);
  const [resolvedServices, setResolvedServices] = useState([]);
  const [unResolvedServices, setUnResolvedServices] = useState([]);

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
  }, [openModal]);

  const onSelectedOASService = async (selectedService) => {
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
  };

  const resolvedArrayBuilder = async (services) => {
    var resServices = [];
    var unResServices = [];
    services.map(s => {
      if(ipAddressWithPortRegex.test(s)){
        unResServices.push(s);
      }
      else {
        resServices.push(s);
      }
    })
    setResolvedServices(resServices);
    setUnResolvedServices(unResServices);
  }

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
          <div
            style={{
              display: "flex",
              justifyContent: "space-between",
              padding: "1%",
            }}>
            <div className={style.selectHeader}>
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
              <img src={closeIcon} alt="Back" onClick={handleCloseModal} />
            </div>
          </div>
          <div className={style.redoc}>
          {selectedServiceSpec && <RedocStandalone spec={selectedServiceSpec} 
              options={{               
                theme:{
                  colors:{
                    responses:{
                      info:{
                        color:"#0d0b1a",
                        tabTextColor:"#1b1b28"
                      },
                      success:{
                        backgroundColor:"#ffffff"
                      }
                    }
                  },
                  rightPanel:{
                    backgroundColor:"#f7f9fc",
                    textColor:"#5f5d87"
                  },
                  sidebar:{
                    backgroundColor:"#ffffff"
                  }
              }
              }}/>}
            </div>
          {/* <div className={style.NotSelectedMessage}>
            {!selectedServiceName && noOasServiceSelectedMessage}
          </div> */}
        </Box>
      </Fade>
    </Modal>
  );
};

export default OasModal;
