import React, { useState, useEffect, useCallback } from "react";
import { Box, Fade, Modal, Backdrop, Button } from "@material-ui/core";
import { toast } from "react-toastify";
import Api from "../../helpers/api";
import spinnerStyle from '../style/Spinner.module.sass';
import spinnerImg from '../assets/spinner.svg';
import Graph from "react-graph-vis";
import variables from '../../variables.module.scss';
import debounce from 'lodash/debounce';
import ServiceMapOptions from './ServiceMapOptions'

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
    top: '10%',
    left: '50%',
    transform: 'translate(-50%, 0%)',
    width: '80vw',
    height: '80vh',
    bgcolor: 'background.paper',
    borderRadius: '5px',
    boxShadow: 24,
    p: 4,
    color: '#000',
};
const buttonStyle: any = {
    margin: "0px 0px 0px 10px",
    backgroundColor: variables.blueColor,
    fontWeight: 600,
    borderRadius: "4px",
    color: "#fff",
    textTransform: "none",
};

const api = Api.getInstance();

export const ServiceMapModal: React.FC<ServiceMapModalProps> = ({ isOpen, onOpen, onClose }) => {
    const [isLoading, setIsLoading] = useState<boolean>(true);
    const [graphData, setGraphData] = useState<GraphData>(null);

    const getServiceMapData = useCallback(async () => {
        try {
            setIsLoading(true)

            const serviceMapData: ServiceMapGraph = await api.serviceMapData()
            const newGraphData: GraphData = { nodes: [], edges: [] }

            if (serviceMapData.nodes) {
                newGraphData.nodes = serviceMapData.nodes.map(node => {
                    return {
                        id: node.id,
                        value: node.count,
                        label: (node.entry.name === "unresolved") ? node.name : `${node.entry.name} (${node.name})`,
                        title: "Count: " + node.name,
                    }
                })
            }

            if (serviceMapData.edges) {
                newGraphData.edges = serviceMapData.edges.map(edge => {
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
                })
            }

            setGraphData(newGraphData)

        } catch (ex) {
            toast.error("An error occurred while loading Mizu Service Map, see console for mode details");
            console.error(ex);
        } finally {
            setIsLoading(false)
        }
    }, [])

    useEffect(() => {
        getServiceMapData()
    }, [getServiceMapData])

    const resetServiceMap = debounce(async () => {
        try {
            const serviceMapResetResponse = await api.serviceMapReset();
            if (serviceMapResetResponse["status"] === "enabled") {
                refreshServiceMap()
            }

        } catch (ex) {
            toast.error("An error occurred while resetting Mizu Service Map, see console for mode details");
            console.error(ex);
        }
    }, 500);

    const refreshServiceMap = debounce(() => {
        // close and re-open modal
        onClose()
        onOpen()
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
                        <Button
                            variant="contained"
                            style={buttonStyle}
                            onClick={() => onClose()}
                        >
                            Close
                        </Button>
                        <Button
                            variant="contained"
                            style={buttonStyle}
                            onClick={resetServiceMap}
                        >
                            Reset
                        </Button>
                        <Button
                            variant="contained"
                            style={{
                                margin: "0px 0px 0px 10px",
                                backgroundColor: variables.blueColor,
                                fontWeight: 600,
                                borderRadius: "4px",
                                color: "#fff",
                                textTransform: "none",
                            }}
                            onClick={refreshServiceMap}
                        >
                            Refresh
                        </Button>
                        <Graph
                            graph={graphData}
                            options={ServiceMapOptions}
                        />
                    </div>}
                </Box>
            </Fade>
        </Modal>
    );

}