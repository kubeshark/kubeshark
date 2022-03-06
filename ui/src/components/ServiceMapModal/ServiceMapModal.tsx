import React, { useState, useEffect, useCallback } from "react";
import { Box, Fade, Modal, Backdrop, Button } from "@material-ui/core";
import { toast } from "react-toastify";
import Api from "../../helpers/api";
import spinnerStyle from '../style/Spinner.module.sass';
import './ServiceMapModal.sass';
import spinnerImg from '../assets/spinner.svg';
import Graph from "react-graph-vis";
import debounce from 'lodash/debounce';
import ServiceMapOptions from './ServiceMapOptions'
import { useCommonStyles } from "../../helpers/commonStyle";
import refresh from "../assets/refresh.svg";
import close from "../assets/close.svg";

interface GraphData {
    nodes: Node[];
    edges: Edge[];
}

interface Node {
    id: number;
    value: number;
    label: string;
    title?: string;
    color?: object;
}

interface Edge {
    from: number;
    to: number;
    value: number;
    label: string;
    title?: string;
    color?: object;
    font?: object;
}

interface ServiceMapNode {
    id: number;
    name: string;
    entry: Entry;
    count: number;
}

interface ServiceMapEdge {
    source: ServiceMapNode;
    destination: ServiceMapNode;
    count: number;
    protocol: Protocol;
}

interface ServiceMapGraph {
    nodes: ServiceMapNode[];
    edges: ServiceMapEdge[];
}

interface Entry {
    ip: string;
    port: string;
    name: string;
}

interface Protocol {
    name: string;
    abbr: string;
    macro: string;
    version: string;
    backgroundColor: string;
    foregroundColor: string;
    fontSize: number;
    referenceLink: string;
    ports: string[];
    priority: number;
}

interface ServiceMapModalProps {
    isOpen: boolean;
    onOpen: () => void;
    onClose: () => void;
}

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

const api = Api.getInstance();

export const ServiceMapModal: React.FC<ServiceMapModalProps> = ({ isOpen, onOpen, onClose }) => {
    const commonClasses = useCommonStyles();
    const [isLoading, setIsLoading] = useState<boolean>(true);
    const [graphData, setGraphData] = useState<GraphData>({ nodes: [], edges: [] });

    const getServiceMapData = useCallback(async () => {
        try {
            setIsLoading(true)

            const serviceMapData: ServiceMapGraph = await api.serviceMapData()
            const newGraphData: GraphData = { nodes: [], edges: [] }

            if (serviceMapData.nodes) {
                newGraphData.nodes = serviceMapData.nodes.map<Node>(node => {
                    return {
                        id: node.id,
                        value: node.count,
                        label: (node.entry.name === "unresolved") ? node.name : `${node.entry.name} (${node.name})`,
                        title: "Count: " + node.name,
                    }
                })
            }

            if (serviceMapData.edges) {
                newGraphData.edges = serviceMapData.edges.map<Edge>(edge => {
                    return {
                        from: edge.source.id,
                        to: edge.destination.id,
                        value: edge.count,
                        label: edge.count.toString(),
                        color: {
                            color: edge.protocol.backgroundColor,
                            highlight: edge.protocol.backgroundColor
                        },
                        font: {
                            color: edge.protocol.backgroundColor,
                            strokeColor: edge.protocol.backgroundColor
                        },
                    }
                })
            }

            setGraphData(newGraphData)

        } catch (ex) {
            toast.error("An error occurred while loading Mizu Service Map, see console for mode details");
            console.error(ex);
        } finally {
            setIsLoading(false)
        }
        // eslint-disable-next-line
    }, [isOpen])

    useEffect(() => {
        getServiceMapData();
        return () => setGraphData({ nodes: [], edges: [] })
    }, [getServiceMapData])

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
            BackdropProps={{
                timeout: 500,
            }}
            style={{ overflow: 'auto' }}
        >
            <Fade in={isOpen}>
                <Box sx={modalStyle}>
                    {isLoading && <div className={spinnerStyle.spinnerContainer}>
                        <img alt="spinner" src={spinnerImg} style={{ height: 50 }} />
                    </div>}
                    {!isLoading && <div style={{ height: "100%", width: "100%" }}>
                    <div style={{ display: "flex", justifyContent: "space-between" }}> 
                        <div>
                        <Button
                            startIcon={<img src={refresh} className="custom" alt="refresh" style={{ marginRight:"8%"}}></img>}
                            size="medium"
                            variant="contained"
                            className={commonClasses.outlinedButton + " " + commonClasses.imagedButton}
                            onClick={refreshServiceMap}
                        >
                            Refresh
                        </Button>
                        </div>
                        <img src={close} alt="close" onClick={() => onClose()} style={{cursor:"pointer"}}></img>
                    </div>
                        <Graph
                            graph={graphData}
                            options={ServiceMapOptions}
                        />
                        <div className='legend-scale'>
                            <ul className='legend-labels'>
                                <li><span style={{ background: '#205cf5' }}></span>HTTP</li>
                                <li><span style={{ background: '#244c5a' }}></span>HTTP/2</li>
                                <li><span style={{ background: '#244c5a' }}></span>gRPC</li>
                                <li><span style={{ background: '#ff6600' }}></span>AMQP</li>
                                <li><span style={{ background: '#000000' }}></span>KAFKA</li>
                                <li><span style={{ background: '#a41e11' }}></span>REDIS</li>
                            </ul>
                        </div>
                    </div>}
                </Box>
            </Fade>
        </Modal>
    );

}