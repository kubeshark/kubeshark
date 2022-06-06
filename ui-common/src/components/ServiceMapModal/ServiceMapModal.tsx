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

const protocolDisplayNameMap = {
    "GQL": "GraphQL"
}

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

type ProtocolType = {
    key: string;
    value: string;
    component: JSX.Element;
};


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
    const [checkedProtocols, setCheckedProtocols] = useState([])
    const [checkedServices, setCheckedServices] = useState([])
    const [serviceMapApiData, setServiceMapApiData] = useState<ServiceMapGraph>({ edges: [], nodes: [] })
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
    const mapToKeyValForFilter = useCallback((arr) => arr.map(mapNodesDatatoGraph)
        .map((edge) => { return { key: edge.label, value: edge.label } })
        .sort((a, b) => { return a.key.localeCompare(b.key) }), [])

    const getProtocolsForFilter = useMemo(() => {
        return serviceMapApiData.edges.reduce<ProtocolType[]>((returnArr, currentValue, currentIndex, array) => {
            if (!returnArr.find(prot => prot.key === currentValue.protocol.abbr))
                returnArr.push({
                    key: currentValue.protocol.abbr, value: currentValue.protocol.abbr,
                    component: <LegentLabel color={currentValue.protocol.backgroundColor}
                        name={protocolDisplayNameMap[currentValue.protocol.abbr] ? protocolDisplayNameMap[currentValue.protocol.abbr] : currentValue.protocol.abbr} />
                })
            return returnArr
        }, new Array<ProtocolType>())
    }, [serviceMapApiData])

    const getServicesForFilter = useMemo(() => {
        const resolved = mapToKeyValForFilter(serviceMapApiData.nodes?.filter(x => x.resolved))
        const unResolved = mapToKeyValForFilter(serviceMapApiData.nodes?.filter(x => !x.resolved))
        return [...resolved, ...unResolved]
    }, [mapToKeyValForFilter, serviceMapApiData.nodes])

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
        if (checkedServices.length === 0)
            setCheckedServices(getServicesForFilter.map(x => x.key).filter(serviceName => !Utils.isIpAddress(serviceName)))
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [getServicesForFilter])

    useEffect(() => {
        if (checkedProtocols.length === 0) {
            setCheckedProtocols(getProtocolsForFilter.map(x => x.key))
        }
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [getProtocolsForFilter])

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
                                className={commonClasses.outlinedButton + " " + commonClasses.imagedButton + ` ${isFilterClicked ? commonClasses.clickedButton : ""}`}
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
                                    <div className={styles.card}>
                                        <SelectList items={getProtocolsForFilter} checkBoxWidth="5%" tableName={"PROTOCOLS"} multiSelect={true}
                                            checkedValues={checkedProtocols} setCheckedValues={onProtocolsChange} tableClassName={styles.filters}
                                            inputSearchClass={styles.servicesFilterSearch} isFilterable={false} />
                                    </div>
                                    <div className={styles.servicesFilterWrapper + ` ${styles.card}`}>
                                        <div className={styles.servicesFilterList}>
                                            <SelectList items={getServicesForFilter} tableName={"SERVICES"} tableClassName={styles.filters} multiSelect={true}
                                                checkBoxWidth="5%" checkedValues={checkedServices} setCheckedValues={onServiceChanges} inputSearchClass={styles.servicesFilterSearch} />
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
