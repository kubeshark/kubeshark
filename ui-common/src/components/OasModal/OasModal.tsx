import { Box, Fade, FormControl, MenuItem, Modal, Backdrop } from "@material-ui/core";
import { useCallback, useEffect, useState } from "react";
import { RedocStandalone } from "redoc";
import closeIcon from "assets/closeIcon.svg";
import { toast } from 'react-toastify';
import style from './OasModal.module.sass';
import openApiLogo from 'assets/openApiLogo.png'
import { redocThemeOptions } from "./redocThemeOptions";
import React from "react";
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


const OasModal = ({ openModal, handleCloseModal, getOasServices, getOasByService }) => {
  const [oasServices, setOasServices] = useState([] as string[])
  const [selectedServiceName, setSelectedServiceName] = useState("");
  const [selectedServiceSpec, setSelectedServiceSpec] = useState(null);

  const onSelectedOASService = async (selectedService) => {
    if (oasServices.length === 0) {
      setSelectedServiceSpec(null);
      setSelectedServiceName("");
      return
    }
    else {
      setSelectedServiceName(selectedService ? selectedService : oasServices[0]);
    }
    try {
      const data = await getOasByService(selectedService ? selectedService : oasServices[0]);
      setSelectedServiceSpec(data);
      } catch (e) {
        toast.error("Error occurred while fetching service OAS spec");
        console.error(e);
      }  
  };

  useEffect(() => {
    (async () => {
      try {
        const services = await getOasServices();
        setOasServices(services);
      } catch (e) {
        console.error(e);
      }
    })();
  }, [openModal]);

  useEffect(() => {
    onSelectedOASService(null);
  },[oasServices])

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
              <div><img src={openApiLogo} alt="openAPI" className={style.openApilogo} /></div>
              <div className={style.title}>OpenAPI</div>
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
                  {oasServices.map((service) => (
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