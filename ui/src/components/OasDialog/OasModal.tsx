import { Box, Fade, FormControl, MenuItem, Modal } from "@material-ui/core";
import { useEffect, useState } from "react";
import { RedocStandalone } from "redoc";
import Api from "../../helpers/api";
import { Select } from "../UI/Select";
import closeIcon from "../assets/closeIcon.svg";

const api = new Api();

const OasDModal = ({ openModal, handleCloseModal, entries }) => {
  const [oasServices, setOASservices] = useState([]);
  const [selectedOASService, setSelectedOASService] = useState("");
  const [serviceOAS, setServiceOAS] = useState(null);

  useEffect(() => {
    (async () => {
      try {
        const services = await api.getOASAServices();
        if (!areEqual(oasServices, services)) setOASservices(services);
      } catch (e) {
        console.error(e);
      }
    })();
  }, [entries,oasServices]);

  const onSelectedOASService = async (selectedService) => {
    setSelectedOASService(selectedService);
    if(oasServices.length === 0){
      return
    }
    try {
      const data = await api.getOASAByService(selectedService);
      setServiceOAS(data);
    } catch (e) {
      console.error(e);
    }
  };

  function areEqual(array1, array2) {
    if (array1.length === array2.length) {
      return array1.every((element) => {
        if (array2.includes(element)) {
          return true;
        }
        return false;
      });
    }
    return false;
  }

  return (
    <Modal
      aria-labelledby="transition-modal-title"
      aria-describedby="transition-modal-description"
      open={openModal}
      onClose={handleCloseModal}
      closeAfterTransition
      hideBackdrop={true}
      style={{ overflow: "auto", backgroundColor: "#ffffff" }}
    >
      <Fade in={openModal}>
        <Box>
          <div
            style={{
              display: "flex",
              justifyContent: "space-between",
              padding: "1%",
            }}
          >
            <div style={{ marginLeft: "40%" }}>
              <FormControl>
                <Select
                  labelId="service-select-label"
                  id="service-select"
                  label="Show OAS"
                  placeholder="Show OAS"
                  value={selectedOASService}
                  onChange={onSelectedOASService}
                >
                  {oasServices.map((service) => (
                    <MenuItem key={service} value={service}>
                      {service}
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>
            </div>
            <div style={{ cursor: "pointer" }}>
              <img src={closeIcon} alt="Back" onClick={handleCloseModal} />
            </div>
          </div>
          {serviceOAS && <RedocStandalone spec={serviceOAS} />}
        </Box>
      </Fade>
    </Modal>
  );
};

export default OasDModal;
