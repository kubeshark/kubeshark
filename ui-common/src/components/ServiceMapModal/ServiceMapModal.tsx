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
import filterIcon from "assets/filter-icon.svg";
import filterIconClicked from "assets/filter-icon-clicked.svg";
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
    bgcolor: '#F0F5FF',
    borderRadius: '5px',
    boxShadow: 24,
    p: 4,
    color: '#000',
    padding: "1px 1px",
    paddingBottom: "15px"
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
    { key: "HTTP", value: "HTTP", component: <LegentLabel color="#494677" name="HTTP" /> },
    { key: "HTTP/2", value: "HTTP/2", component: <LegentLabel color='#F7B202' name="HTTP/2" /> },
    { key: "gRPC", value: "gRPC", component: <LegentLabel color='#219653' name="gRPC" /> },
    { key: "GQL", value: "GQL", component: <LegentLabel color='#e10098' name="GQL" /> },
    { key: "AMQP", value: "AMQP", component: <LegentLabel color='#F86818' name="AMQP" /> },
    { key: "KAFKA", value: "KAFKA", component: <LegentLabel color='#0C0B1A' name="KAFKA" /> },
    { key: "REDIS", value: "REDIS", component: <LegentLabel color='#DB2156' name="REDIS" /> },]


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
    const [isFilterClicked, setIsFilterClicked] = useState(true)

    const getServiceMapData = useCallback(async () => {
        try {
            setIsLoading(true)

            const serviceMapData: ServiceMapGraph = await getServiceMapDataApi()
            setServiceMapApiData(serviceMapData)
            const newGraphData: GraphData = { nodes: [], edges: [] }
            newGraphData.nodes = serviceMapData.nodes.map(mapNodesDatatoGraph)
            newGraphData.edges = serviceMapData.edges.map(mapEdgesDatatoGraph)
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

    useEffect(() => {
        const newGraphData: GraphData = {
            nodes: serviceMapApiData.nodes?.map(mapNodesDatatoGraph).filter(node => checkedServices.includes(node.label)),
            edges: serviceMapApiData.edges?.filter(edge => checkedProtocols.includes(edge.protocol.abbr)).map(mapEdgesDatatoGraph)
        }
        setGraphData(newGraphData);
    }, [checkedServices, checkedProtocols, serviceMapApiData])

    const onProtocolsChange = (newProtocolsFiltersnewProt) => {
        const filterProt = newProtocolsFiltersnewProt || checkedProtocols
        setCheckedProtocols(filterProt)
    }

    const onServiceChanges = (newServiceFilters) => {
        const filterService = newServiceFilters || checkedServices
        setCheckedServices([...filterService])
    }

    useEffect(() => {
        if (checkedServices.length == 0)
            setCheckedServices(getServicesForFilter.map(x => x.key).filter(serviceName => !Utils.isIpAddress(serviceName)))
    }, [getServicesForFilter])

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
                    <div className={styles.closeIcon}>
                        <img src={closeIcon} alt="close" onClick={() => onClose()} style={{ cursor: "pointer", userSelect: "none" }}></img>
                    </div>
                    <div className={styles.headerContainer}>
                        <div className={styles.headerSection}>
                            <span className={styles.title}>Services</span>
                            <Button size="medium"
                                variant="contained"
                                startIcon={<img src={isFilterClicked ? filterIconClicked : filterIcon} className="custom" alt="refresh" style={{ height: "26px", width: "26px" }}></img>}
                                className={commonClasses.button + " " + commonClasses.imagedButton + " " + `${isFilterClicked ? commonClasses.button : commonClasses.outlinedButton}`}
                                onClick={() => setIsFilterClicked(prevState => !prevState)}
                                style={{ textTransform: 'unset' }}>
                                Filter
                            </Button >
                            <Button style={{ marginLeft: "2%", textTransform: 'unset' }}
                                startIcon={<img src={refreshIcon} className="custom" alt="refresh"></img>}
                                size="medium"
                                variant="contained"
                                className={commonClasses.outlinedButton + " " + commonClasses.imagedButton}
                                onClick={refreshServiceMap}
                            >
                                Refresh
                            </Button>
                        </div>
                    </div>

                    <div className={styles.modalContainer}>
                        <div className={styles.filterSection + ` ${isFilterClicked ? styles.show : ""}`}>
                            <Resizeable minWidth={170} maxWidth={320}>
                                <div className={styles.filterWrapper}>
                                    <div className={styles.protocolsFilterList}>
                                        <h3 className={styles.subHeader} style={{ marginLeft: "10px" }}>
                                            PROTOCOLS
                                            <span className={styles.totalSelected}>&nbsp;({checkedProtocols.length})</span>
                                        </h3>
                                        <SelectList items={protocols} checkBoxWidth="5%" tableName={"All"} multiSelect={true}
                                            checkedValues={checkedProtocols} setCheckedValues={onProtocolsChange} tableClassName={styles.filters} />
                                    </div>
                                    <div className={styles.servicesFilter}>
                                        <h3 className={styles.subHeader} style={{ marginLeft: "10px" }}>
                                            SERVICES
                                            <span className={styles.totalSelected}>&nbsp;({checkedServices.length})</span>
                                        </h3>
                                        <input className={commonClasses.textField + ` ${styles.servicesFilterSearch}`} placeholder="search service" value={servicesSearchVal} onChange={(event) => setServicesSearchVal(event.target.value)} />
                                        <div className={styles.servicesFilterList}>
                                            <SelectList items={getServicesForFilter} tableName={"All"} tableClassName={styles.filters} multiSelect={true} searchValue={servicesSearchVal}
                                                checkBoxWidth="5%" checkedValues={checkedServices} setCheckedValues={onServiceChanges} />
                                        </div>
                                    </div>
                                </div>
                            </Resizeable>
                        </div>
                        <div className={styles.graphSection}>
                            <div style={{ display: "flex", justifyContent: "space-between" }}>
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
