import { Box, Fade, FormControl, MenuItem, Modal } from "@material-ui/core";
import { useEffect, useState } from "react";
import { RedocStandalone } from "redoc";
import Api from "../../helpers/api";
import { Select } from "../UI/Select";
import closeIcon from "../assets/closeIcon.svg";
import { arrayElementsComparetion } from "../../helpers/utils";
import { toast } from 'react-toastify';

const api = new Api();

const OasModal = ({ openModal, handleCloseModal, entries }) => {
  const [oasServices, setOASservices] = useState([]);
  const [selectedOASService, setSelectedOASService] = useState("");
  const [oasService, setServiceOAS] = useState(null);

  const noOasServiceSelectedMessage = "Please Select OasService";

  useEffect(() => {
    (async () => {
      try {
        const services = await api.getOASAServices();
        if (!arrayElementsComparetion(oasServices, services)) setOASservices(services);
      } catch (e) {
        toast.error(e.message);
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

  return (
    <Modal
      aria-labelledby="transition-modal-title"
      aria-describedby="transition-modal-description"
      open={openModal}
      onClose={handleCloseModal}
      closeAfterTransition
      hideBackdrop={true}
      style={{ overflow: "auto", backgroundColor: "#ffffff", color:"black" }}
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
          {oasService && <RedocStandalone spec={oasService} />}
          <div className="NotSelectedMessage">
            {!selectedOASService && noOasServiceSelectedMessage}
          </div>
        </Box>
      </Fade>
    </Modal>
  );
};

export default OasModal;
