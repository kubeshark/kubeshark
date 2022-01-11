import React, { useState, useEffect, useCallback } from "react";
import { Box, Fade, Modal, Backdrop, Button } from "@material-ui/core";
import Api from "../../helpers/api";
import spinnerStyle from '../style/Spinner.module.sass';
import spinnerImg from '../assets/spinner.svg';
import Graph from "react-graph-vis";
import variables from '../../variables.module.scss';
import debounce from 'lodash/debounce';

interface GraphData {
    nodes: Node[];
    edges: Edge[];
}

interface Node {
    id: number;
    value: number;
    label: string;
    title: string;
    color?: string;
}

interface Edge {
    from: number;
    to: number;
    value: number;
    title: string;
    color?: string;
}

interface ServiceMapNode {
    name: string;
    id: number;
    protocol: string;
    count: number;
}

interface ServiceMapEdge {
    source: ServiceMapNode;
    destination: ServiceMapNode;
    count: number;
}

interface ServiceMapGraph {
    nodes: ServiceMapNode[];
    edges: ServiceMapEdge[];
}

interface ServiceMapModalProps {
    isOpen: boolean;
    onOpen: () => void;
    onClose: () => void;
}

function getProtocolColor(protocol: string): string {
    let color;
    switch (protocol) {
        case "http": {
            color = "#27AE60"
            break;
        }
        case "https": {
            // TODO: https protocol color
            break;
        }
        case "redis": {
            color = "#A41E11"
            break;
        }
        case "amqp": {
            color = "#FF6600"
            break;
        }
        case "grpc": {
            color = "#244C5A"
            break;
        }
        case "kafka": {
            color = "#000000"
            break;
        }
        default: {
            color = variables.lightBlueColor
            break;
        }
    }
    return color
}

const api = Api.getInstance();

export const ServiceMapModal: React.FC<ServiceMapModalProps> = ({ isOpen, onOpen, onClose }) => {
    const [isLoading, setIsLoading] = useState<boolean>(true);
    const [graphData, setGraphData] = useState<GraphData>({
        nodes: [],
        edges: []
    });

    const options = {
        layout: {
            hierarchical: false,
            randomSeed: 1 // always on node 1
        },
        nodes: {
            shape: "dot",
            color: variables.blueColor
        },
        height: "750px",
    };

    const modalStyle = {
        position: 'absolute',
        top: '10%',
        left: '50%',
        transform: 'translate(-50%, 0%)',
        width: '80vw',
        bgcolor: 'background.paper',
        borderRadius: '5px',
        boxShadow: 24,
        p: 4,
        color: '#000',
    };

    const getServiceMapData = useCallback(async () => {
        console.log("getServiceMapData called")
        try {
            setIsLoading(true)

            const serviceMapData: ServiceMapGraph = await api.serviceMapData()
            console.log(serviceMapData)

            if (serviceMapData.nodes != null) {
                for (let i = 0; i < serviceMapData.nodes.length; i++) {
                    graphData.nodes.push({
                        id: serviceMapData.nodes[i].id,
                        value: serviceMapData.nodes[i].count,
                        label: serviceMapData.nodes[i].name,
                        title: "Count: " + serviceMapData.nodes[i].name,
                    });
                }
            }

            if (serviceMapData.edges != null) {
                for (let i = 0; i < serviceMapData.edges.length; i++) {
                    graphData.edges.push({
                        from: serviceMapData.edges[i].source.id,
                        to: serviceMapData.edges[i].destination.id,
                        value: serviceMapData.edges[i].count,
                        title: "Count: " + serviceMapData.edges[i].count,
                        color: getProtocolColor(serviceMapData.edges[i].source.protocol)
                    });
                }
            }

            setGraphData(graphData)
            setIsLoading(false)

        } catch (error) {
            setIsLoading(false)
            console.error(error);
        }
    }, [graphData])

    useEffect(() => {
        getServiceMapData()
    }, [getServiceMapData])

    const resetServiceMap = debounce(async () => {
        try {
            const serviceMapResetResponse = await api.serviceMapReset();
            if (serviceMapResetResponse["status"] === "enabled") {
                refreshServiceMap()
            }

        } catch (error) {
            console.error(error);
        }
    }, 500);

    const refreshServiceMap = debounce(() => {
        console.log("refreshServiceMap called")

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
                    {!isLoading && <div>
                        <Button
                            variant="contained"
                            style={{
                                margin: "0px 0px 0px 0px",
                                backgroundColor: variables.blueGray,
                                fontWeight: 600,
                                borderRadius: "4px",
                                color: "#fff",
                                textTransform: "none",
                            }}
                            onClick={() => onClose()}
                        >
                            Close
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
                            options={options}
                            style={{
                                border: "1px solid lightgray",
                            }}
                        />
                    </div>}
                </Box>
            </Fade>
        </Modal>
    );

}