import { Box, Fade, FormControl, MenuItem, Modal } from "@material-ui/core";
import { useEffect, useState } from "react";
import { RedocStandalone } from "redoc";
import Api from "../../../helpers/api";
import { Select } from "../../UI/Select";
import closeIcon from "../../assets/closeIcon.svg";
import { toast } from 'react-toastify';
import './OasModal.sass'

const api = Api.getInstance();
const noOasServiceSelectedMessage = "Please Select OasService";

const OasModal = ({ openModal, handleCloseModal }) => { 
  const [oasServices, setOasServices] = useState([])
  const [selectedServiceName, setSelectedServiceName] = useState("");
  const [selectedServiceSpec, setSelectedServiceSpec] = useState(null);

  useEffect(() => {
    (async () => {
      try {
        const services = await api.getOasServices();
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
                  value={selectedServiceName}
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
          {selectedServiceSpec && <RedocStandalone spec={selectedServiceSpec} />}
          <div className="NotSelectedMessage">
            {!selectedServiceName && noOasServiceSelectedMessage}
          </div>
        </Box>
      </Fade>
    </Modal>
  );
};

export default OasModal;
