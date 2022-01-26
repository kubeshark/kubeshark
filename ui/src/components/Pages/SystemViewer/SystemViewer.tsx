import React from "react";
import {Outlet} from "react-router-dom";
import {ServiceMapModal} from "../../ServiceMapModal/ServiceMapModal";
import {EntHeader} from "../../Header/EntHeader";
import {useRecoilState} from "recoil";
import serviceMapModalOpenAtom from "../../../recoil/serviceMapModalOpen";

const SystemViewer = ({isFirstLogin, setIsFirstLogin}) => {

    const [serviceMapModalOpen, setServiceMapModalOpen] = useRecoilState(serviceMapModalOpenAtom);

    return <>
        <EntHeader isFirstLogin={isFirstLogin} setIsFirstLogin={setIsFirstLogin} />
        <Outlet/>
        {window["isServiceMapEnabled"] && <ServiceMapModal
            isOpen={serviceMapModalOpen}
            onOpen={() => setServiceMapModalOpen(true)}
            onClose={() => setServiceMapModalOpen(false)}
        />}
    </>
}

export default SystemViewer;
