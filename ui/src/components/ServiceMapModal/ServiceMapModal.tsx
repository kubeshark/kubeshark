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
    name: string;
    id: number;
    protocol: Protocol;
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

const api = Api.getInstance();

export const ServiceMapModal: React.FC<ServiceMapModalProps> = ({ isOpen, onOpen, onClose }) => {
    const [isLoading, setIsLoading] = useState<boolean>(true);
    const [graphData, setGraphData] = useState<GraphData>({
        nodes: [],
        edges: []
    });

    const options = {
        physics: {
            enabled: true,
            solver: 'barnesHut',
            barnesHut: {
                theta: 0.5,
                gravitationalConstant: -2000,
                centralGravity: 0.3,
                springLength: 180,
                springConstant: 0.04,
                damping: 0.09,
                avoidOverlap: 1
            },
        },
        layout: {
            hierarchical: false,
            randomSeed: 1 // always on node 1
        },
        nodes: {
            shape: 'dot',
            chosen: true,
            color: {
                background: '#27AE60',
                border: '#000000',
                highlight: {
                    background: '#27AE60',
                    border: '#000000',
                },
            },
            font: {
                color: '#343434',
                size: 14, // px
                face: 'arial',
                background: 'none',
                strokeWidth: 0, // px
                strokeColor: '#ffffff',
                align: 'center',
                multi: false,
            },
            borderWidth: 1.5,
            borderWidthSelected: 2.5,
            labelHighlightBold: true,
            opacity: 1,
            shadow: true,
        },
        edges: {
            chosen: true,
            dashes: false,
            arrowStrikethrough: false,
            arrows: {
                to: {
                    enabled: true,
                },
                middle: {
                    enabled: false,
                },
                from: {
                    enabled: false,
                }
            },
            smooth: {
                enabled: true,
                type: 'dynamic',
                roundness: 1.0
            },
            font: {
                color: '#343434',
                size: 12, // px
                face: 'arial',
                background: 'none',
                strokeWidth: 2, // px
                strokeColor: '#ffffff',
                align: 'horizontal',
                multi: false,
            },
            labelHighlightBold: true,
            selectionWidth: 1,
            shadow: true,
        },
        height: '750px',
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
        try {
            setIsLoading(true)

            const serviceMapData: ServiceMapGraph = await api.serviceMapData()

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
                        label: serviceMapData.edges[i].count.toString(),
                        color: {
                            color: serviceMapData.edges[i].source.protocol.backgroundColor,
                            highlight: serviceMapData.edges[i].source.protocol.backgroundColor
                        },
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
                        />
                    </div>}
                </Box>
            </Fade>
        </Modal>
    );

}