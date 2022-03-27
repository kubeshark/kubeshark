import { Box, Fade, FormControl, MenuItem, Modal, Backdrop, ListSubheader } from "@material-ui/core";
import { useCallback, useEffect, useState } from "react";
import { RedocStandalone } from "redoc";
import closeIcon from "assets/closeIcon.svg";
import { toast } from 'react-toastify';
import style from './OasModal.module.sass';
import openApiLogo from 'assets/openApiLogo.png'
import { redocThemeOptions } from "./redocThemeOptions";
import React from "react";
import { TOAST_CONTAINER_ID } from "../../configs/Consts";
import { Select } from "../UI/Select";


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

const ipAddressWithPortRegex = new RegExp('([0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3}):([0-9]{1,5})');

const OasModal = ({ openModal, handleCloseModal, getOasServices, getOasByService }) => {
  const [oasServices, setOasServices] = useState([] as string[])
  const [selectedServiceName, setSelectedServiceName] = useState("");
  const [selectedServiceSpec, setSelectedServiceSpec] = useState(null);
  const [resolvedServices, setResolvedServices] = useState([]);
  const [unResolvedServices, setUnResolvedServices] = useState([]);

  const onSelectedOASService = useCallback(async (selectedService) => {
    if (!!selectedService) {
      setSelectedServiceName(selectedService);
      if (oasServices.length === 0) {
        return
      }
      try {
        const data = await getOasByService(selectedService);
        setSelectedServiceSpec(data);
      } catch (e) {
        toast.error("Error occurred while fetching service OAS spec", { containerId: TOAST_CONTAINER_ID });
        console.error(e);
      }
    }
  }, [oasServices.length])

  const resolvedArrayBuilder = useCallback(async (services) => {
    const resServices = [];
    const unResServices = [];
    services.forEach(s => {
      if (ipAddressWithPortRegex.test(s)) {
        unResServices.push(s);
      }
      else {
        resServices.push(s);
      }
    });

    resServices.sort();
    unResServices.sort();
    if (resServices.length > 0) {
      onSelectedOASService(resServices[0]);
    }
    else {
      onSelectedOASService(unResServices[0]);
    }

    setResolvedServices(resServices);
    setUnResolvedServices(unResServices);
  }, [onSelectedOASService])

  useEffect(() => {
    (async () => {
      try {
        const services = await getOasServices();
        resolvedArrayBuilder(services);
        setOasServices(services);
      } catch (e) {
        console.error(e);
      }
    })();
  }, [openModal, resolvedArrayBuilder]);


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
              <div><img src={openApiLogo} alt="openApi" className={style.openApilogo} /></div>
              <div className={style.title}>OpenApi </div>
            </div>
            <div style={{ cursor: "pointer" }}>
              <img src={closeIcon} alt="close" onClick={handleCloseModal} />
            </div>
          </div>
          <div className={style.selectContainer} >
            <FormControl>
              <Select
                labelId="service-select-label"
                id="service-select"
                value={selectedServiceName}
                onChangeCb={onSelectedOASService}
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
          <div className={style.borderLine}></div>
          <div className={style.redoc}>
            {selectedServiceSpec && <RedocStandalone
              spec={selectedServiceSpec}
              options={redocThemeOptions} />}
          </div>
        </Box>
      </Fade>
    </Modal>
  );
};

export default OasModal;