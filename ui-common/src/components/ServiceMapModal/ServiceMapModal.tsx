import React, { useState, useEffect, useCallback, useMemo } from "react";
import { Box, Fade, Modal, Backdrop, Button } from "@material-ui/core";
import { toast } from "react-toastify";
import spinnerStyle from '../UI/style/Spinner.module.sass';
import spinnerImg from 'assets/spinner.svg';
import Graph from "react-graph-vis";
import debounce from 'lodash/debounce';
import ServiceMapOptions from './ServiceMapOptions'
import { useCommonStyles } from "../../helpers/commonStyle";
import refreshIcon from "assets/refresh.svg";
import closeIcon from "assets/close.svg"
import styles from './ServiceMapModal.module.sass'
import SelectList from "../UI/SelectList";
import { GraphData, ServiceMapGraph } from "./ServiceMapModalTypes"
import { Utils } from "../../helpers/Utils";
import { TOAST_CONTAINER_ID } from "../../configs/Consts";
import Resizeable from "../UI/Resizeable"

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
    padding: "25px 15px"
};

interface LegentLabelProps {
    color: string,
    name: string
}

const LegentLabel: React.FC<LegentLabelProps> = ({ color, name }) => {
    return <React.Fragment>
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
            <span title={name}>{name}</span>
            <span style={{ background: color }} className={styles.colorBlock}></span>
        </div>
    </React.Fragment>
}

const protocols = [
    { key: "HTTP", value: "HTTP", component: <LegentLabel color="#205cf5" name="HTTP" /> },
    { key: "HTTP/2", value: "HTTP/2", component: <LegentLabel color='#244c5a' name="HTTP/2" /> },
    { key: "gRPC", value: "gRPC", component: <LegentLabel color='#244c5a' name="gRPC" /> },
    { key: "AMQP", value: "AMQP", component: <LegentLabel color='#ff6600' name="AMQP" /> },
    { key: "KAFKA", value: "KAFKA", component: <LegentLabel color='#000000' name="KAFKA" /> },
    { key: "REDIS", value: "REDIS", component: <LegentLabel color='#a41e11' name="REDIS" /> },]


interface ServiceMapModalProps {
    isOpen: boolean;
    onOpen: () => void;
    onClose: () => void;
    getServiceMapDataApi: () => Promise<any>
}

export const ServiceMapModal: React.FC<ServiceMapModalProps> = ({ isOpen, onClose, getServiceMapDataApi }) => {
    const commonClasses = useCommonStyles();
    const [isLoading, setIsLoading] = useState<boolean>(true);
    const [graphData, setGraphData] = useState<GraphData>({ nodes: [], edges: [] });
    const [checkedProtocols, setCheckedProtocols] = useState(protocols.map(x => x.key))
    const [checkedServices, setCheckedServices] = useState([])
    const [serviceMapApiData, setServiceMapApiData] = useState<ServiceMapGraph>({ edges: [], nodes: [] })
    const [servicesSearchVal, setServicesSearchVal] = useState("")
    const [graphOptions, setGraphOptions] = useState(ServiceMapOptions);

    const getServiceMapData = useCallback(async () => {
        try {
            setIsLoading(true)

            const serviceMapData: ServiceMapGraph = await getServiceMapDataApi()
            setServiceMapApiData(serviceMapData)
            const newGraphData: GraphData = { nodes: [], edges: [] }

            if (serviceMapData.nodes) {
                newGraphData.nodes = serviceMapData.nodes.map(mapNodesDatatoGraph)
            }

            if (serviceMapData.edges) {
                newGraphData.edges = serviceMapData.edges.map(mapEdgesDatatoGraph)
            }

            setGraphData(newGraphData)
        } catch (ex) {
            toast.error("An error occurred while loading Mizu Service Map, see console for mode details", { containerId: TOAST_CONTAINER_ID });
            console.error(ex);
        } finally {
            setIsLoading(false)
        }
        // eslint-disable-next-line
    }, [isOpen])

    const mapNodesDatatoGraph = node => {
        return {
            id: node.id,
            value: node.count,
            label: (node.entry.name === "unresolved") ? node.name : `${node.entry.name} (${node.name})`,
            title: "Count: " + node.name,
            isResolved: node.entry.resolved
        }
    }

    const mapEdgesDatatoGraph = edge => {
        return {
            from: edge.source.id,
            to: edge.destination.id,
            value: edge.count,
            label: edge.count.toString(),
            color: {
                color: edge.protocol.backgroundColor,
                highlight: edge.protocol.backgroundColor
            },
        }
    }
    const mapToKeyValForFilter = (arr) => arr.map(mapNodesDatatoGraph)
        .map((edge) => { return { key: edge.label, value: edge.label } })
        .sort((a, b) => { return a.key.localeCompare(b.key) });

    const getServicesForFilter = useMemo(() => {

        const resolved = mapToKeyValForFilter(serviceMapApiData.nodes?.filter(x => x.resolved))
        const unResolved = mapToKeyValForFilter(serviceMapApiData.nodes?.filter(x => !x.resolved))
        return [...resolved, ...unResolved]
    }, [serviceMapApiData])

    const filterServiceMap = useCallback((newProtocolsFilters?: any[], newServiceFilters?: string[]) => {
        const filterProt = newProtocolsFilters || checkedProtocols
        const filterService = newServiceFilters || checkedServices
        setCheckedProtocols(filterProt)
        setCheckedServices(filterService)
        const newGraphData: GraphData = {
            nodes: serviceMapApiData.nodes?.map(mapNodesDatatoGraph).filter(node => filterService.includes(node.label)),
            edges: serviceMapApiData.edges?.filter(edge => filterProt.includes(edge.protocol.abbr)).map(mapEdgesDatatoGraph)
        }
        setGraphData(newGraphData);
    }, [checkedProtocols, checkedServices, serviceMapApiData])



    useEffect(() => {
        if (checkedServices.length > 0) return // only after refresh
        filterServiceMap(checkedProtocols, getServicesForFilter.map(x => x.key).filter(serviceName => !Utils.isIpAddress(serviceName)))
    }, [getServicesForFilter, checkedServices])

    useEffect(() => {
        getServiceMapData()
    }, [getServiceMapData])

    useEffect(() => {
        if (graphData?.nodes?.length === 0) return;
        let options = { ...graphOptions };
        options.physics.barnesHut.avoidOverlap = graphData?.nodes?.length > 10 ? 0 : 1;
        setGraphOptions(options);
        // eslint-disable-next-line
    }, [graphData?.nodes?.length])

    const refreshServiceMap = debounce(() => {
        getServiceMapData();
    }, 500);

    return (
        <Modal
            aria-labelledby="transition-modal-title"
            aria-describedby="transition-modal-description"
            open={isOpen}
            onClose={onClose}
            closeAfterTransition
            BackdropComponent={Backdrop}
            BackdropProps={{ timeout: 500 }}>
            <Fade in={isOpen}>
                <Box sx={modalStyle}>
                    <div className={styles.modalContainer}>
                        <div className={styles.filterSection}>
                            <Resizeable minWidth={170}>
                                <div className={styles.filterWrapper}>
                                    <div className={styles.protocolsFilterList}>
                                        <SelectList items={protocols} checkBoxWidth="5%" tableName={"Protocols"} multiSelect={true}
                                            checkedValues={checkedProtocols} setCheckedValues={filterServiceMap} tableClassName={styles.filters} />
                                    </div>
                                    <div className={styles.separtorLine}></div>
                                    <div className={styles.servicesFilter}>
                                        <input className={commonClasses.textField + ` ${styles.servicesFilterSearch}`} placeholder="search service" value={servicesSearchVal} onChange={(event) => setServicesSearchVal(event.target.value)} />
                                        <div className={styles.servicesFilterList}>
                                            <SelectList items={getServicesForFilter} tableName={"Services"} tableClassName={styles.filters} multiSelect={true} searchValue={servicesSearchVal}
                                                checkBoxWidth="5%" checkedValues={checkedServices} setCheckedValues={(newServicesForFilter) => filterServiceMap(null, newServicesForFilter)} />
                                        </div>
                                    </div>
                                </div>
                            </Resizeable>
                        </div>
                        <div className={styles.graphSection}>
                            <div style={{ display: "flex", justifyContent: "space-between" }}>
                                <Button style={{ marginLeft: "3%" }}
                                    startIcon={<img src={refreshIcon} className="custom" alt="refresh" style={{ marginRight: "8%" }}></img>}
                                    size="medium"
                                    variant="contained"
                                    className={commonClasses.outlinedButton + " " + commonClasses.imagedButton}
                                    onClick={refreshServiceMap}
                                >
                                    Refresh
                                </Button>
                                <img src={closeIcon} alt="close" onClick={() => onClose()} style={{ cursor: "pointer", userSelect: "none" }}></img>
                            </div>
                            {isLoading && <div className={spinnerStyle.spinnerContainer}>
                                <img alt="spinner" src={spinnerImg} style={{ height: 50 }} />
                            </div>}
                            {!isLoading && <div style={{ height: "100%", width: "100%" }}>
                                <Graph
                                    graph={graphData}
                                    options={graphOptions}
                                />
                            </div>
                            }
                        </div>
                    </div>
                </Box>
            </Fade>
        </Modal>
    );
}
